package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	a "github.com/dhth/kplay/internal/awsweb"
	k "github.com/dhth/kplay/internal/kafka"
	"github.com/dhth/kplay/internal/scan"
	t "github.com/dhth/kplay/internal/types"
	"github.com/spf13/cobra"
)

func newScanCommand(
	preRunE func(cmd *cobra.Command, args []string) error,
	config *t.Config,
	consumeBehaviours *t.ConsumeBehaviours,
	fromOffset *string,
	fromTimestamp *string,
	outputDir *string,
	debug *bool,
	defaultOutputDir string,
) *cobra.Command {
	var scanKeyFilterRegexStr string
	var scanNumMessages uint
	var scanSaveMessages bool
	var scanDecode bool
	var scanBatchSize uint

	cmd := &cobra.Command{
		Use:               "scan <PROFILE>",
		Short:             "scan messages in a kafka topic and optionally write them to the local filesystem",
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		PersistentPreRunE: preRunE,
		RunE: func(_ *cobra.Command, _ []string) error {
			if scanBatchSize == 0 {
				return fmt.Errorf("batch size must be greater than 0")
			}

			if scanNumMessages == 0 {
				return fmt.Errorf("count must be greater than 0")
			}

			var keyFilterRegex *regexp.Regexp
			if strings.TrimSpace(scanKeyFilterRegexStr) != "" {
				var regexErr error
				keyFilterRegex, regexErr = regexp.Compile(scanKeyFilterRegexStr)
				if regexErr != nil {
					return fmt.Errorf("%w: %q", errInvalidRegexProvided, scanKeyFilterRegexStr)
				}
			}

			scanBehaviours := scan.Behaviours{
				NumMessages:    scanNumMessages,
				KeyFilterRegex: keyFilterRegex,
				SaveMessages:   scanSaveMessages,
				Decode:         scanDecode,
				BatchSize:      scanBatchSize,
			}

			if *debug {
				fmt.Printf(`%s
  output directory        %s

%s

%s
`,
					config.Display(),
					*outputDir,
					scanBehaviours.Display(),
					consumeBehaviours.Display(),
				)

				return nil
			}

			var awsConfig *aws.Config
			if config.Authentication == t.AWSMSKIAM {
				awsCfg, err := a.GetAWSConfig(context.Background())
				if err != nil {
					return err
				}

				awsConfig = &awsCfg
			}

			client, err := k.GetKafkaClient(
				config.Authentication,
				config.Brokers,
				config.Topic,
				*consumeBehaviours,
				awsConfig,
			)
			if err != nil {
				return err
			}

			defer client.Close()

			scanner := scan.New(client, *config, scanBehaviours, *outputDir)

			return scanner.Execute()
		},
	}

	cmd.Flags().StringVarP(fromOffset, "from-offset", "o", "", "scan messages from this offset; provide a single offset for all partitions (eg. 1000) or specify offsets per partition (e.g., '0:1000,2:1500')")
	cmd.Flags().StringVarP(fromTimestamp, "from-timestamp", "t", "", "scan messages from this timestamp (in RFC3339 format, e.g., 2006-01-02T15:04:05Z07:00)")
	cmd.Flags().StringVarP(&scanKeyFilterRegexStr, "key-regex", "k", "", "regex to filter message keys by")
	cmd.Flags().UintVarP(&scanNumMessages, "num-records", "n", scan.ScanNumRecordsDefault, "maximum number of messages to scan")
	cmd.Flags().BoolVarP(&scanSaveMessages, "save-messages", "s", false, "whether to save kafka messages to the local filesystem")
	cmd.Flags().BoolVarP(&scanDecode, "decode", "d", true, "whether to decode message values (false is equivalent to 'encodingFormat: raw' in kplay's config)")
	cmd.Flags().UintVarP(&scanBatchSize, "batch-size", "b", 100, "number of messages to fetch per batch (must be greater than 0)")
	cmd.Flags().StringVarP(outputDir, "output-dir", "O", defaultOutputDir, "directory to save scan results in")

	return cmd
}
