package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	a "github.com/dhth/kplay/internal/awsweb"
	k "github.com/dhth/kplay/internal/kafka"
	"github.com/dhth/kplay/internal/server"
	t "github.com/dhth/kplay/internal/types"
	"github.com/spf13/cobra"
)

func newServeCmd(
	preRunE func(cmd *cobra.Command, args []string) error,
	config *t.Config,
	consumeBehaviours *t.ConsumeBehaviours,
	fromOffset *string,
	fromTimestamp *string,
	debug *bool,
) *cobra.Command {
	var selectOnHover bool
	var webOpen bool

	cmd := &cobra.Command{
		Use:               "serve <PROFILE>",
		Short:             "open kplay's web interface",
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		PersistentPreRunE: preRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			behaviours := server.Behaviours{
				SelectOnHover: selectOnHover,
			}
			if *debug {
				fmt.Printf(`%s

%s
`,
					config.Display(),
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

			return server.Serve(cl, *config, behaviours, webOpen)
		},
	}

	cmd.Flags().StringVarP(fromOffset, "from-offset", "o", "", "start consuming messages from this offset; provide a single offset for all partitions (eg. 1000) or specify offsets per partition (e.g., '0:1000,2:1500')")
	cmd.Flags().StringVarP(fromTimestamp, "from-timestamp", "t", "", "start consuming messages from this timestamp (in RFC3339 format, e.g., 2006-01-02T15:04:05Z07:00)")
	cmd.Flags().BoolVarP(&selectOnHover, "select-on-hover", "S", false, "whether to start the web interface with the setting \"select on hover\" ON")
	cmd.Flags().BoolVarP(&webOpen, "open", "O", false, "whether to open web interface in browser automatically")

	return cmd
}
