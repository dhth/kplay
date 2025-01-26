package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	c "github.com/dhth/kplay/internal/config"
	k "github.com/dhth/kplay/internal/kafka"
	"github.com/dhth/kplay/internal/tui"
	"github.com/dhth/kplay/internal/utils"
	"github.com/spf13/cobra"
)

const (
	configFileName    = "kplay/kplay.yml"
	consumerGroupFlag = "consumer-group"
)

var (
	errCouldntCreateKafkaClient = errors.New("couldn't create kafka client")
	errCouldntPingBrokers       = errors.New("couldn't ping the brokers")
	errCouldntGetUserHomeDir    = errors.New("couldn't get your home directory")
	errCouldntGetUserConfigDir  = errors.New("couldn't get your config directory")
	ErrCouldntReadConfigFile    = errors.New("couldn't read config file")
	ErrConfigInvalid            = errors.New("config is invalid")
)

func Execute() error {
	rootCmd, err := NewRootCommand()
	if err != nil {
		return err
	}

	return rootCmd.Execute()
}

func NewRootCommand() (*cobra.Command, error) {
	var (
		configPath        string
		configPathFull    string
		homeDir           string
		persistMessages   bool
		skipMessages      bool
		consumerGroup     string
		config            c.Config
		displayConfigOnly bool
	)

	rootCmd := &cobra.Command{
		Use:   "kplay <PROFILE>",
		Short: "kplay lets you inspect messages in a Kafka topic in a simple and deliberate manner.",
		Long: `kplay ("kafka playground") lets you inspect messages in a Kafka topic in a simple and deliberate manner.

kplay relies on a configuration file that contains profiles for various Kafka topics, each with its own details related
to brokers, message encoding, authentication, etc.
`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			configPathFull = utils.ExpandTilde(configPath, homeDir)
			configBytes, err := os.ReadFile(configPathFull)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrCouldntReadConfigFile, err)
			}

			config, err = GetProfileConfig(configBytes, args[0], homeDir)
			if errors.Is(err, errProfileNotFound) {
				return err
			} else if err != nil {
				return fmt.Errorf("%w: %w", ErrConfigInvalid, err)
			}

			if cmd.Flags().Changed(consumerGroupFlag) {
				if strings.TrimSpace(consumerGroup) == "" {
					return errConsumerGroupEmpty
				}

				config.ConsumerGroup = consumerGroup
			}

			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			behaviours := c.Behaviours{PersistMessages: persistMessages, SkipMessages: skipMessages}
			if displayConfigOnly {
				fmt.Printf(`Config:
---

- topic                   %s
- consumer group          %s
- authentication          %s
- encoding                %s
- brokers                 %v

Behaviours 
---

- persist messages        %v
- skip messages           %v
`,
					config.Topic,
					config.ConsumerGroup,
					config.AuthenticationDisplay(),
					config.EncodingDisplay(),
					config.Brokers,
					behaviours.PersistMessages,
					behaviours.SkipMessages)
				return nil
			}

			cl, err := k.GetClient(config.Authentication, config.Brokers, config.ConsumerGroup, config.Topic)
			if err != nil {
				return fmt.Errorf("%w: %s", errCouldntCreateKafkaClient, err.Error())
			}

			defer cl.Close()

			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			if err := cl.Ping(ctx); err != nil {
				return fmt.Errorf("%w: %s", errCouldntPingBrokers, err.Error())
			}

			return tui.Render(cl, config, behaviours)
		},
	}

	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntGetUserHomeDir, err.Error())
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntGetUserConfigDir, err.Error())
	}

	defaultConfigPath := filepath.Join(configDir, configFileName)

	rootCmd.Flags().StringVarP(&configPath, "config-path", "c", defaultConfigPath, "location of kplay's config file")
	rootCmd.Flags().BoolVarP(&persistMessages, "persist-messages", "p", false, "whether to start the TUI with the \"persist messages\" setting ON")
	rootCmd.Flags().BoolVarP(&skipMessages, "skip-messages", "s", false, "whether to start the TUI with the \"skip messages\" setting ON")
	rootCmd.Flags().StringVarP(&consumerGroup, consumerGroupFlag, "g", "", "consumer group to use (overrides the one in kplay's config file)")
	rootCmd.Flags().BoolVar(&displayConfigOnly, "display-config-only", false, "whether to only display config picked up by kplay")

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	return rootCmd, nil
}
