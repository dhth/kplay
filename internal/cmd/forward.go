package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	a "github.com/dhth/kplay/internal/awsweb"
	f "github.com/dhth/kplay/internal/forwarder"
	k "github.com/dhth/kplay/internal/kafka"
	"github.com/dhth/kplay/internal/utils"
	"github.com/spf13/cobra"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	envVarConsumerGroup          = "KPLAY_FORWARD_CONSUMER_GROUP"
	envVarFetchBatchSize         = "KPLAY_FORWARD_FETCH_BATCH_SIZE"
	envVarNumUploadWorkers       = "KPLAY_FORWARD_NUM_UPLOAD_WORKERS"
	envVarShutdownTimeoutMillis  = "KPLAY_FORWARD_SHUTDOWN_TIMEOUT_MILLIS"
	envVarPollFetchTimeoutMillis = "KPLAY_FORWARD_POLL_FETCH_TIMEOUT_MILLIS"
	envVarUploadTimeoutMillis    = "KPLAY_FORWARD_UPLOAD_TIMEOUT_MILLIS"
	envVarPollSleepMillis        = "KPLAY_FORWARD_POLL_SLEEP_MILLIS"
	envVarUploadReports          = "KPLAY_FORWARD_UPLOAD_REPORTS"
	envVarReportBatchSize        = "KPLAY_FORWARD_REPORT_BATCH_SIZE"
	envVarRunServer              = "KPLAY_FORWARD_RUN_SERVER"
	envVarHost                   = "KPLAY_FORWARD_SERVER_HOST"
	envVarPort                   = "KPLAY_FORWARD_SERVER_PORT"

	// longest env var
	// KPLAY_FORWARD_POLL_FETCH_TIMEOUT_MILLIS -> 39
	envVarHelpPadding = 42

	s3DestinationPrefix = "arn:aws:s3:::"
	maxProfilesAllowed  = 10

	consumerGroupDefault   = "kplay-forwarder"
	consumerGroupMinLength = 5
	consumerGroupMaxLength = 255

	fetchBatchSizeDefault = 50
	fetchBatchSizeMin     = 1
	fetchBatchSizeMax     = 1000

	numUploadWorkersDefault = 50
	numUploadWorkersMin     = 1
	numUploadWorkersMax     = 500

	shutdownTimeoutMillisDefault = 30 * 1000
	shutdownTimeoutMillisMin     = 10 * 1000
	shutdownTimeoutMillisMax     = 60 * 1000

	pollFetchTimeoutMillisDefault = 10 * 1000
	pollFetchTimeoutMillisMin     = 1 * 1000
	pollFetchTimeoutMillisMax     = 60 * 1000

	pollSleepMillisDefault = 5 * 1000
	pollSleepMillisMin     = 0
	pollSleepMillisMax     = 30 * 60 * 1000

	uploadTimeoutMillisDefault = 10 * 1000
	uploadTimeoutMillisMin     = 1 * 1000
	uploadTimeoutMillisMax     = 60 * 1000

	uploadReportsDefault = false

	reportBatchSizeDefault = 5000
	reportBatchSizeMin     = 1000
	reportBatchSizeMax     = 20000

	runServerDefault = false
	hostDefault      = "127.0.0.1"

	portDefault = 8080
	portMin     = 0
	portMax     = 65535
)

var (
	errTooManyForwardProfiles     = errors.New("too many profiles provided")
	errInvalidDestinationProvided = errors.New("invalid destination provided")
	errDestinationEmpty           = errors.New("destination is empty")
)

func newForwardCmd(configPath *string, homeDir string, debug *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forward <PROFILE>,<PROFILE>,... <DESTINATION>",
		Short: "Consume messages in a kafka topic and forward them to a remote destination",
		Long: fmt.Sprintf(`This command is useful when you want to consume messages in a kafka topic as
part of a consumer group, decode them, and forward the decoded contents to a
remote destination (AWS S3 is the only supported destination for now).

This command is intended to be run in a long running containerised environment;
as such, it accepts configuration via the following environment variables.

- %s consumer group to use (default: %s)
- %s number of records to fetch per batch (default: %d, range: %d-%d)
- %s number of upload workers (default: %d, range: %d-%d)
- %s graceful shutdown timeout in ms (default: %d, range: %d-%d)
- %s kafka polling fetch timeout in ms (default: %d, range: %d-%d)
- %s kafka polling sleep interval in ms (default: %d, range: %d-%d)
- %s upload timeout in ms (default: %d, range: %d-%d)
- %s whether to upload reports of the messages forwarded (default: %v)
- %s report batch size (default: %d, range: %d-%d)
- %s whether to run an http server alongside the forwarder (default: %v)
- %s host to run the server on (default: %s)
- %s port to run the server on (default: %d)

If needed, this command can also start an HTTP server which can be used for
health checks (at /health).
`,
			utils.RightPadTrim(envVarConsumerGroup, envVarHelpPadding), consumerGroupDefault,
			utils.RightPadTrim(envVarFetchBatchSize, envVarHelpPadding), fetchBatchSizeDefault, fetchBatchSizeMin, fetchBatchSizeMax,
			utils.RightPadTrim(envVarNumUploadWorkers, envVarHelpPadding), numUploadWorkersDefault, numUploadWorkersMin, numUploadWorkersMax,
			utils.RightPadTrim(envVarShutdownTimeoutMillis, envVarHelpPadding), shutdownTimeoutMillisDefault, shutdownTimeoutMillisMin, shutdownTimeoutMillisMax,
			utils.RightPadTrim(envVarPollFetchTimeoutMillis, envVarHelpPadding), pollFetchTimeoutMillisDefault, pollFetchTimeoutMillisMin, pollFetchTimeoutMillisMax,
			utils.RightPadTrim(envVarPollSleepMillis, envVarHelpPadding), pollSleepMillisDefault, pollSleepMillisMin, pollSleepMillisMax,
			utils.RightPadTrim(envVarUploadTimeoutMillis, envVarHelpPadding), uploadTimeoutMillisDefault, uploadTimeoutMillisMin, uploadTimeoutMillisMax,
			utils.RightPadTrim(envVarUploadReports, envVarHelpPadding), uploadReportsDefault,
			utils.RightPadTrim(envVarReportBatchSize, envVarHelpPadding), reportBatchSizeDefault, reportBatchSizeMin, reportBatchSizeMax,
			utils.RightPadTrim(envVarRunServer, envVarHelpPadding), runServerDefault,
			utils.RightPadTrim(envVarHost, envVarHelpPadding), hostDefault,
			utils.RightPadTrim(envVarPort, envVarHelpPadding), portDefault,
		),
		Example:      `kplay forward profile-1,profile-2 arn:aws:s3:::bucket-to-forward-messages-to/prefix`,
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPathFromEnvVar := os.Getenv(envVarConfigPath)
			if configPathFromEnvVar != "" && !cmd.Flags().Changed("config-path") {
				*configPath = configPathFromEnvVar
			}

			configPathFull := utils.ExpandTilde(*configPath, homeDir)
			configBytes, err := os.ReadFile(configPathFull)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrCouldntReadConfigFile, err)
			}

			profileNames := strings.Split(args[0], ",")

			configs, err := ParseProfileConfigs(configBytes, profileNames, homeDir)

			if errors.Is(err, errProfileNotFound) {
				return err
			} else if err != nil {
				return fmt.Errorf("%w: %w", ErrConfigInvalid, err)
			}

			if len(configs) == 0 {
				return nil
			}

			if len(configs) > maxProfilesAllowed {
				return fmt.Errorf("%w; provided: %d, upper limit: %d",
					errTooManyForwardProfiles,
					len(configs),
					maxProfilesAllowed,
				)
			}

			forwardBehaviours, err := getBehaviorsFromEnv()
			if err != nil {
				return err
			}

			destinationStr := strings.TrimSpace(args[1])
			if len(destinationStr) == 0 {
				return errDestinationEmpty
			}

			destinationWithoutPrefix, ok := strings.CutPrefix(destinationStr, s3DestinationPrefix)
			if !ok {
				return fmt.Errorf("%w; supported destination prefixes: [%s]", errInvalidDestinationProvided, s3DestinationPrefix)
			}

			if *debug {
				configDebug := make([]string, len(configs))
				for i, c := range configs {
					configDebug[i] = c.Display()
				}
				fmt.Printf(`%s

Destination               %s

%s
`,
					strings.Join(configDebug, "\n"),
					destinationStr,
					forwardBehaviours.Display(),
				)

				return nil
			}

			ctx := cmd.Context()

			awsConfig, err := a.GetAWSConfig(ctx)
			if err != nil {
				return err
			}

			destination, err := f.NewS3Destination(awsConfig, destinationWithoutPrefix)
			if err != nil {
				return err
			}

			var kafkaClients []*kgo.Client

			for _, config := range configs {
				client, err := k.GetKafkaClientForForwarding(
					config.Authentication,
					config.Brokers,
					config.Topic,
					forwardBehaviours.ConsumerGroup,
					&awsConfig,
				)
				if err != nil {
					return err
				}

				pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)

				err = client.Ping(pingCtx)
				pingCancel()
				if err != nil {
					return fmt.Errorf("%w (profile: %q): %s", errCouldntPingBrokers, config.Name, err.Error())
				}

				kafkaClients = append(kafkaClients, client)
			}

			defer func() {
				for _, client := range kafkaClients {
					client.Close()
				}
			}()

			var profileConfigNames []string
			for _, c := range configs {
				profileConfigNames = append(profileConfigNames, c.Name)
			}

			forwarder := f.New(kafkaClients, configs, &destination, forwardBehaviours)
			logStartupInfo(profileConfigNames, destination.Display(), forwardBehaviours)

			return forwarder.Execute(ctx)
		},
	}

	return cmd
}

func getBehaviorsFromEnv() (f.Behaviours, error) {
	var errs []error

	consumerGroup, err := getConstrainedStringEnvVar(envVarConsumerGroup, consumerGroupDefault, consumerGroupMinLength, consumerGroupMaxLength)
	if err != nil {
		errs = append(errs, err)
	}

	fetchBatchSize, err := getUint16EnvVar(
		envVarFetchBatchSize,
		fetchBatchSizeDefault,
		fetchBatchSizeMin,
		fetchBatchSizeMax,
	)
	if err != nil {
		errs = append(errs, err)
	}

	numUploadWorkers, err := getUint16EnvVar(
		envVarNumUploadWorkers,
		numUploadWorkersDefault,
		numUploadWorkersMin,
		numUploadWorkersMax,
	)
	if err != nil {
		errs = append(errs, err)
	}

	shutdownTimeoutMillis, err := getUint16EnvVar(
		envVarShutdownTimeoutMillis,
		shutdownTimeoutMillisDefault,
		shutdownTimeoutMillisMin,
		shutdownTimeoutMillisMax,
	)
	if err != nil {
		errs = append(errs, err)
	}

	pollSleepMillis, err := getUint32EnvVar(
		envVarPollSleepMillis,
		pollSleepMillisDefault,
		pollSleepMillisMin,
		pollSleepMillisMax,
	)
	if err != nil {
		errs = append(errs, err)
	}

	pollFetchTimeoutMillis, err := getUint16EnvVar(
		envVarPollFetchTimeoutMillis,
		pollFetchTimeoutMillisDefault,
		pollFetchTimeoutMillisMin,
		pollFetchTimeoutMillisMax,
	)
	if err != nil {
		errs = append(errs, err)
	}

	uploadTimeoutMillis, err := getUint16EnvVar(
		envVarUploadTimeoutMillis,
		uploadTimeoutMillisDefault,
		uploadTimeoutMillisMin,
		uploadTimeoutMillisMax,
	)
	if err != nil {
		errs = append(errs, err)
	}

	uploadReports, err := getBoolEnvVar(envVarUploadReports, uploadReportsDefault)
	if err != nil {
		errs = append(errs, err)
	}

	reportBatchSize, err := getUint16EnvVar(
		envVarReportBatchSize,
		reportBatchSizeDefault,
		reportBatchSizeMin,
		reportBatchSizeMax,
	)
	if err != nil {
		errs = append(errs, err)
	}

	runServer, err := getBoolEnvVar(envVarRunServer, runServerDefault)
	if err != nil {
		errs = append(errs, err)
	}

	host := getStringEnvVar(envVarHost, hostDefault)

	port, err := getUint16EnvVar(
		envVarPort,
		portDefault,
		portMin,
		portMax,
	)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		if len(errs) == 1 {
			return f.Behaviours{}, errs[0]
		}
		return f.Behaviours{}, fmt.Errorf("multiple issues:\n\n%w", errors.Join(errs...))
	}

	return f.Behaviours{
		ConsumerGroup:                  consumerGroup,
		FetchBatchSize:                 fetchBatchSize,
		NumUploadWorkers:               numUploadWorkers,
		ForwarderShutdownTimeoutMillis: shutdownTimeoutMillis,
		PollSleepMillis:                pollSleepMillis,
		PollFetchTimeoutMillis:         pollFetchTimeoutMillis,
		UploadTimeoutMillis:            uploadTimeoutMillis,
		UploadReports:                  uploadReports,
		ReportBatchSize:                reportBatchSize,
		RunServer:                      runServer,
		ServerHost:                     host,
		ServerPort:                     port,
	}, nil
}

func logStartupInfo(profileConfigNames []string, destination string, behaviours f.Behaviours) {
	slog.Info("starting up")
	slog.Info("input", "profiles", strings.Join(profileConfigNames, ","))
	slog.Info("input", "destination", destination)
	slog.Info("behaviour", "consumer_group", behaviours.ConsumerGroup)
	slog.Info("behaviour", "fetch_batch_size", behaviours.FetchBatchSize)
	slog.Info("behaviour", "upload_workers", behaviours.NumUploadWorkers)
	slog.Info("behaviour", "shutdown_timeout_millis", behaviours.ForwarderShutdownTimeoutMillis)
	slog.Info("behaviour", "poll_sleep_millis", behaviours.PollSleepMillis)
	slog.Info("behaviour", "poll_fetch_timeout_millis", behaviours.PollFetchTimeoutMillis)
	slog.Info("behaviour", "upload_timeout_millis", behaviours.UploadTimeoutMillis)
	if behaviours.UploadReports {
		slog.Info("behaviour", "upload_reports", "true", "report_batch_size", behaviours.ReportBatchSize)
	} else {
		slog.Info("behaviour", "upload_reports", "false")
	}

	if behaviours.RunServer {
		slog.Info("behaviour", "run_server", "true", "host", behaviours.ServerHost, "port", behaviours.ServerPort)
	} else {
		slog.Info("behaviour", "run_server", "false")
	}
}
