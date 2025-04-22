package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	k "github.com/dhth/kplay/internal/kafka"
	"github.com/dhth/kplay/internal/server"
	"github.com/dhth/kplay/internal/tui"
	t "github.com/dhth/kplay/internal/types"
	"github.com/dhth/kplay/internal/utils"
	"github.com/spf13/cobra"
)

const (
	configFileName    = "kplay/kplay.yml"
	consumerGroupFlag = "consumer-group"
)

var (
	errCouldntCreateKafkaClient = errors.New("couldn't create kafka client")
	errCouldntPingBrokers       = errors.New("couldn't ping brokers")
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
		configPath      string
		configPathFull  string
		homeDir         string
		persistMessages bool
		skipMessages    bool
		commitMessages  bool
		selectOnHover   bool
		consumerGroup   string
		config          t.Config
		debug           bool
		webOpen         bool
	)

	rootCmd := &cobra.Command{
		Use:   "kplay",
		Short: "kplay lets you inspect messages in a Kafka topic in a simple and deliberate manner.",
		Long: `kplay ("kafka playground") lets you inspect messages in a Kafka topic in a simple and deliberate manner.

kplay relies on a configuration file that contains profiles for various Kafka topics, each with its own details related
to brokers, message encoding, authentication, etc.
`,
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
	}

	tuiCmd := &cobra.Command{
		Use:          "tui <PROFILE>",
		Short:        "open kplay's TUI",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			behaviours := t.TUIBehaviours{
				CommitMessages:  commitMessages,
				PersistMessages: persistMessages,
				SkipMessages:    skipMessages,
			}

			if debug {
				fmt.Printf(`Config:
---

- topic                   %s
- consumer group          %s
- authentication          %s
- encoding                %s
- brokers                 %v

Behaviours
---
%s
`,
					config.Topic,
					config.ConsumerGroup,
					config.AuthenticationDisplay(),
					config.EncodingDisplay(),
					config.Brokers,
					behaviours.Display(),
				)
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

	serveCmd := &cobra.Command{
		Use:          "serve <PROFILE>",
		Short:        "open kplay's web interface",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			behaviours := t.WebBehaviours{
				CommitMessages: commitMessages,
				SelectOnHover:  selectOnHover,
			}
			if debug {
				fmt.Printf(`Config:
---

- topic                   %s
- consumer group          %s
- authentication          %s
- encoding                %s
- brokers                 %v

Behaviours
---
%s
`,
					config.Topic,
					config.ConsumerGroup,
					config.AuthenticationDisplay(),
					config.EncodingDisplay(),
					config.Brokers,
					behaviours.Display(),
				)
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

			return server.Serve(cl, config, behaviours, webOpen)
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

	tuiCmd.Flags().StringVarP(&configPath, "config-path", "c", defaultConfigPath, "location of kplay's config file")
	tuiCmd.Flags().BoolVarP(&persistMessages, "persist-messages", "p", false, "whether to start the TUI with the setting \"persist messages\" ON")
	tuiCmd.Flags().BoolVarP(&skipMessages, "skip-messages", "s", false, "whether to start the TUI with the setting \"skip messages\" ON")
	tuiCmd.Flags().BoolVarP(&commitMessages, "commit-messages", "C", true, "whether to start the TUI with the setting \"commit messages\" ON")
	tuiCmd.Flags().StringVarP(&consumerGroup, consumerGroupFlag, "g", "", "consumer group to use (overrides the one in kplay's config file)")
	tuiCmd.Flags().BoolVar(&debug, "debug", false, "whether to only display config picked up by kplay without running it")

	serveCmd.Flags().StringVarP(&configPath, "config-path", "c", defaultConfigPath, "location of kplay's config file")
	serveCmd.Flags().StringVarP(&consumerGroup, consumerGroupFlag, "g", "", "consumer group to use (overrides the one in kplay's config file)")
	serveCmd.Flags().BoolVarP(&commitMessages, "commit-messages", "C", true, "whether to start the web interface with the setting \"commit messages\" ON")
	serveCmd.Flags().BoolVarP(&selectOnHover, "select-on-hover", "S", false, "whether to start the web interface with the setting \"select on hover\" ON")
	serveCmd.Flags().BoolVarP(&webOpen, "open", "o", false, "whether to open web interface in browser automatically")
	serveCmd.Flags().BoolVar(&debug, "debug", false, "whether to only display config picked up by kplay without running it")

	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(serveCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	return rootCmd, nil
}
