package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
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
	envVarForwardConsumerGroup = "KPLAY_FORWARD_CONSUMER_GROUP"
	envVarForwardRunServer     = "KPLAY_FORWARD_RUN_SERVER"
	envVarForwardHost          = "KPLAY_FORWARD_HOST"
	envVarForwardPort          = "KPLAY_FORWARD_PORT"

	forwardS3DestinationPrefix    = "arn:aws:s3:::"
	forwardConsumerGroupMinLength = 5
	forwardMaxProfilesAllowed     = 10
	forwardConsumerGroupDefault   = "kplay-forwarder"
	forwardRunServerDefault       = false
	forwardHostDefault            = "127.0.0.1"
	forwardPortDefault            = 8080
)

var (
	errConsumerGroupTooShort      = errors.New("consumer group is too short")
	errTooManyForwardProfiles     = errors.New("too many profiles provided")
	errInvalidDestinationProvided = errors.New("invalid destination provided")
	errDestinationEmpty           = errors.New("destination is empty")
)

func newForwardCmd(configPath *string, homeDir string, debug *bool) *cobra.Command {
	var (
		forwardConsumerGroup string
		forwardRunServer     bool
		forwardHost          string
		forwardPort          uint16
	)

	cmd := &cobra.Command{
		Use:   "forward <PROFILE>,<PROFILE>,... <DESTINATION>",
		Short: "fetch messages in a kafka topic and forward them to a remote destination",
		Long: fmt.Sprintf(`fetch messages in a kafka topic and forward them to a remote destination.
AWS S3 is the only supported destination for now.

This command uses the following environment variables for optional configuration
as it's designed to be run in long-lived containers where environment-based
configuration is more suitable than command-line flags.

- %s: consumer group to use (default: %s)
- %s: whether to run an http server alongside the forwarder (can be used for
    health checks) (default: %v)
- %s: host to run the server on (default: %s)
- %s: port to run the server on (default: %d)
`,
			envVarForwardConsumerGroup, forwardConsumerGroupDefault,
			envVarForwardRunServer, forwardRunServerDefault,
			envVarForwardHost, forwardHostDefault,
			envVarForwardPort, forwardPortDefault,
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

			if len(configs) > forwardMaxProfilesAllowed {
				return fmt.Errorf("%w; provided: %d, upper limit: %d",
					errTooManyForwardProfiles,
					len(configs),
					forwardMaxProfilesAllowed,
				)
			}

			forwardConsumerGroup = os.Getenv(envVarForwardConsumerGroup)
			if forwardConsumerGroup == "" {
				forwardConsumerGroup = forwardConsumerGroupDefault
			}

			runServerStr := os.Getenv(envVarForwardRunServer)
			if runServerStr != "" {
				var err error
				forwardRunServer, err = strconv.ParseBool(runServerStr)
				if err != nil {
					return fmt.Errorf("invalid value for %s: %q; expected a boolean value", envVarForwardRunServer, runServerStr)
				}
			} else {
				forwardRunServer = false
			}

			forwardHost = os.Getenv(envVarForwardHost)
			if forwardHost == "" {
				forwardHost = forwardHostDefault
			}

			portStr := os.Getenv(envVarForwardPort)
			if portStr != "" {
				port64, err := strconv.ParseUint(portStr, 10, 16)
				if err != nil {
					return fmt.Errorf("invalid value for %s: %q; expected a valid port number (0-65535)", envVarForwardPort, portStr)
				}
				forwardPort = uint16(port64)
			} else {
				forwardPort = forwardPortDefault
			}

			forwarderCg := strings.TrimSpace(forwardConsumerGroup)
			if len(forwarderCg) < forwardConsumerGroupMinLength {
				return fmt.Errorf("%w (%q); needs to be atleast %d characters",
					errConsumerGroupTooShort,
					forwardConsumerGroup,
					forwardConsumerGroupMinLength,
				)
			}

			destinationStr := strings.TrimSpace(args[1])
			if len(destinationStr) == 0 {
				return errDestinationEmpty
			}

			destinationWithoutPrefix, ok := strings.CutPrefix(destinationStr, forwardS3DestinationPrefix)
			if !ok {
				return fmt.Errorf("%w; supported destination prefixes: [%s]", errInvalidDestinationProvided, forwardS3DestinationPrefix)
			}

			forwardBehaviours := f.Behaviours{
				RunServer: forwardRunServer,
				Host:      forwardHost,
				Port:      forwardPort,
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
					forwardConsumerGroup,
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
			slog.Info("starting up",
				"profiles", strings.Join(profileConfigNames, ","),
				"destination", destination.Display(),
				"consumer_group", forwardConsumerGroup,
			)

			forwarder := f.New(kafkaClients, configs, &destination, forwardBehaviours)

			return forwarder.Execute(ctx)
		},
	}

	return cmd
}
