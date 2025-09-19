package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	a "github.com/dhth/kplay/internal/awsweb"
	k "github.com/dhth/kplay/internal/kafka"
	"github.com/dhth/kplay/internal/tui"
	t "github.com/dhth/kplay/internal/types"
	"github.com/spf13/cobra"
)

func newTuiCmd(
	preRunE func(cmd *cobra.Command, args []string) error,
	config *t.Config,
	consumeBehaviours *t.ConsumeBehaviours,
	fromOffset *string,
	fromTimestamp *string,
	outputDir *string,
	debug *bool,
	defaultOutputDir string,
) *cobra.Command {
	var persistMessages bool
	var skipMessages bool

	cmd := &cobra.Command{
		Use:   "tui <PROFILE>",
		Short: "Browse messages in a kafka topic via a TUI",
		Long: `

This will start a TUI which will let you browse messages on demand. You can then
browse the message metadata and value in a pager. By default, kplay will consume
messages from the earliest possible offset, but you can modify this behaviour by
either providing an offset or a timestamp to start consuming messages from.
`,
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		PersistentPreRunE: preRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			behaviours := tui.Behaviours{
				PersistMessages: persistMessages,
				SkipMessages:    skipMessages,
			}

			if *debug {
				fmt.Printf(`%s
  output directory        %s

%s
`,
					config.Display(),
					*outputDir,
					behaviours.Display(),
				)

				return nil
			}

			var awsConfig *aws.Config
			if config.Authentication == t.AWSMSKIAM {
				awsCfg, err := a.GetAWSConfig(cmd.Context())
				if err != nil {
					return err
				}

				awsConfig = &awsCfg
			}

			cl, err := k.GetKafkaClient(
				config.Authentication,
				config.Brokers,
				config.Topic,
				*consumeBehaviours,
				awsConfig,
			)
			if err != nil {
				return fmt.Errorf("%w: %s", errCouldntCreateKafkaClient, err.Error())
			}

			defer cl.Close()

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			if err := cl.Ping(ctx); err != nil {
				return fmt.Errorf("%w: %s", errCouldntPingBrokers, err.Error())
			}

			return tui.Render(cl, *config, behaviours, *outputDir)
		},
	}

	cmd.Flags().BoolVarP(&persistMessages, "persist-messages", "p", false, "whether to start the TUI with the setting \"persist messages\" ON")
	cmd.Flags().BoolVarP(&skipMessages, "skip-messages", "s", false, "whether to start the TUI with the setting \"skip messages\" ON")
	cmd.Flags().StringVarP(fromOffset, "from-offset", "o", "", "start consuming messages from this offset; provide a single offset for all partitions (eg. 1000) or specify offsets per partition (e.g., '0:1000,2:1500')")
	cmd.Flags().StringVarP(fromTimestamp, "from-timestamp", "t", "", "start consuming messages from this timestamp (in RFC3339 format, e.g., 2006-01-02T15:04:05Z07:00)")
	cmd.Flags().StringVarP(outputDir, "output-dir", "O", defaultOutputDir, "directory to persist messages in")

	return cmd
}
