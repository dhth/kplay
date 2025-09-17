package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	k "github.com/dhth/kplay/internal/kafka"
	"github.com/dhth/kplay/internal/scan"
	"github.com/dhth/kplay/internal/server"
	"github.com/dhth/kplay/internal/tui"
	t "github.com/dhth/kplay/internal/types"
	"github.com/dhth/kplay/internal/utils"
	"github.com/spf13/cobra"
)

const (
	configFileName   = "kplay/kplay.yml"
	envVarConfigPath = "KPLAY_CONFIG_PATH"
)

var (
	errCouldntCreateKafkaClient = errors.New("couldn't create kafka client")
	errCouldntPingBrokers       = errors.New("couldn't ping brokers")
	errCouldntGetUserHomeDir    = errors.New("couldn't get your home directory")
	errCouldntGetUserConfigDir  = errors.New("couldn't get your config directory")
	ErrCouldntReadConfigFile    = errors.New("couldn't read config file")
	ErrConfigInvalid            = errors.New("config is invalid")
	errInvalidTimestampProvided = errors.New(`invalid value provided for "from timestamp"`)
	errInvalidOffsetProvided    = errors.New(`invalid value provided for "from offset"`)
	errInvalidRegexProvided     = errors.New("invalid regex provided")
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
		configPath     string
		configPathFull string
		homeDir        string
		outputDir      string
		fromOffset     string
		fromTimestamp  string
		debug          bool
		config         t.Config

		persistMessages bool
		skipMessages    bool
		selectOnHover   bool
		webOpen         bool

		scanKeyFilterRegexStr string
		scanNumMessages       uint
		scanSaveMessages      bool
		scanDecode            bool
		scanBatchSize         uint

		consumeBehaviours t.ConsumeBehaviours
	)

	rootCmd := &cobra.Command{
		Use:   "kplay",
		Short: "kplay lets you inspect messages in a Kafka topic in a simple and deliberate manner.",
		Long: `kplay ("kafka playground") lets you inspect messages in a Kafka topic in a simple and deliberate manner.

kplay relies on a configuration file that contains profiles for various Kafka topics, each with its own details related
to brokers, message encoding, authentication, etc.
`,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			configPathFromEnvVar := os.Getenv(envVarConfigPath)
			if configPathFromEnvVar != "" && !cmd.Flags().Changed("config-path") {
				configPath = configPathFromEnvVar
			}

			configPathFull = utils.ExpandTilde(configPath, homeDir)
			configBytes, err := os.ReadFile(configPathFull)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrCouldntReadConfigFile, err)
			}

			config, err = ParseProfileConfig(configBytes, args[0], homeDir)
			if errors.Is(err, errProfileNotFound) {
				return err
			} else if err != nil {
				return fmt.Errorf("%w: %w", ErrConfigInvalid, err)
			}

			fromTimestampChanged := cmd.Flags().Changed("from-timestamp")
			fromOffsetChanged := cmd.Flags().Changed("from-offset")
			if fromTimestampChanged && fromOffsetChanged {
				return fmt.Errorf("cannot use both --from-timestamp and --from-offset flags simultaneously")
			}

			var parsedTimestamp *time.Time
			if fromTimestampChanged {
				t, err := time.Parse(time.RFC3339, fromTimestamp)
				if err != nil {
					return fmt.Errorf("%w: %q; expected RFC3339 format (e.g., 2006-01-02T15:04:05Z07:00)",
						errInvalidTimestampProvided, fromTimestamp)
				}
				parsedTimestamp = &t
				consumeBehaviours.StartTimeStamp = parsedTimestamp
			} else if fromOffsetChanged {
				startOffset, partitionOffsets, err := parseFromOffset(fromOffset)
				if err != nil {
					return fmt.Errorf("%w: %s", errInvalidOffsetProvided, err.Error())
				}

				if startOffset != nil {
					consumeBehaviours.StartOffset = startOffset
				} else {
					consumeBehaviours.PartitionOffsets = partitionOffsets
				}
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
			behaviours := tui.Behaviours{
				PersistMessages: persistMessages,
				SkipMessages:    skipMessages,
			}

			if debug {
				fmt.Printf(`%s
  output directory        %s

%s
`,
					config.Display(),
					outputDir,
					behaviours.Display(),
				)

				return nil
			}

			cl, err := k.GetKafkaClient(
				config.Authentication,
				config.Brokers,
				config.Topic,
				consumeBehaviours,
			)
			if err != nil {
				return fmt.Errorf("%w: %s", errCouldntCreateKafkaClient, err.Error())
			}

			defer cl.Close()

			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			if err := cl.Ping(ctx); err != nil {
				return fmt.Errorf("%w: %s", errCouldntPingBrokers, err.Error())
			}

			return tui.Render(cl, config, behaviours, outputDir)
		},
	}

	serveCmd := &cobra.Command{
		Use:          "serve <PROFILE>",
		Short:        "open kplay's web interface",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			behaviours := server.Behaviours{
				SelectOnHover: selectOnHover,
			}
			if debug {
				fmt.Printf(`%s

%s
`,
					config.Display(),
					behaviours.Display(),
				)
				return nil
			}

			cl, err := k.GetKafkaClient(
				config.Authentication,
				config.Brokers,
				config.Topic,
				consumeBehaviours,
			)
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

	scanCmd := &cobra.Command{
		Use:          "scan <PROFILE>",
		Short:        "scan messages in a kafka topic and optionally write them to the local filesystem",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
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

			if debug {
				fmt.Printf(`%s
  output directory        %s

%s

%s
`,
					config.Display(),
					outputDir,
					scanBehaviours.Display(),
					consumeBehaviours.Display(),
				)

				return nil
			}

			client, err := k.GetKafkaClient(config.Authentication, config.Brokers, config.Topic, consumeBehaviours)
			if err != nil {
				return err
			}

			defer client.Close()

			scanner := scan.New(client, config, scanBehaviours, outputDir)

			return scanner.Execute()
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
	defaultOutputDir := filepath.Join(homeDir, ".kplay")

	tuiCmd.Flags().StringVarP(&configPath, "config-path", "c", defaultConfigPath, fmt.Sprintf("location of kplay's config file (can also be provided via $%s)", envVarConfigPath))
	tuiCmd.Flags().BoolVarP(&persistMessages, "persist-messages", "p", false, "whether to start the TUI with the setting \"persist messages\" ON")
	tuiCmd.Flags().BoolVarP(&skipMessages, "skip-messages", "s", false, "whether to start the TUI with the setting \"skip messages\" ON")
	tuiCmd.Flags().StringVarP(&fromOffset, "from-offset", "o", "", "start consuming messages from this offset; provide a single offset for all partitions (eg. 1000) or specify offsets per partition (e.g., '0:1000,2:1500')")
	tuiCmd.Flags().StringVarP(&fromTimestamp, "from-timestamp", "t", "", "start consuming messages from this timestamp (in RFC3339 format, e.g., 2006-01-02T15:04:05Z07:00)")
	tuiCmd.Flags().BoolVar(&debug, "debug", false, "whether to only display config picked up by kplay without running it")
	tuiCmd.Flags().StringVarP(&outputDir, "output-dir", "O", defaultOutputDir, "directory to persist messages in")

	serveCmd.Flags().StringVarP(&configPath, "config-path", "c", defaultConfigPath, fmt.Sprintf("location of kplay's config file (can also be provided via $%s)", envVarConfigPath))
	serveCmd.Flags().StringVarP(&fromOffset, "from-offset", "o", "", "start consuming messages from this offset; provide a single offset for all partitions (eg. 1000) or specify offsets per partition (e.g., '0:1000,2:1500')")
	serveCmd.Flags().StringVarP(&fromTimestamp, "from-timestamp", "t", "", "start consuming messages from this timestamp (in RFC3339 format, e.g., 2006-01-02T15:04:05Z07:00)")
	serveCmd.Flags().BoolVarP(&selectOnHover, "select-on-hover", "S", false, "whether to start the web interface with the setting \"select on hover\" ON")
	serveCmd.Flags().BoolVarP(&webOpen, "open", "O", false, "whether to open web interface in browser automatically")
	serveCmd.Flags().BoolVar(&debug, "debug", false, "whether to only display config picked up by kplay without running it")

	scanCmd.Flags().StringVarP(&configPath, "config-path", "c", defaultConfigPath, fmt.Sprintf("location of kplay's config file (can also be provided via $%s)", envVarConfigPath))
	scanCmd.Flags().BoolVar(&debug, "debug", false, "whether to only display config picked up by kplay without running it")
	scanCmd.Flags().StringVarP(&fromOffset, "from-offset", "o", "", "scan messages from this offset; provide a single offset for all partitions (eg. 1000) or specify offsets per partition (e.g., '0:1000,2:1500')")
	scanCmd.Flags().StringVarP(&fromTimestamp, "from-timestamp", "t", "", "scan messages from this timestamp (in RFC3339 format, e.g., 2006-01-02T15:04:05Z07:00)")
	scanCmd.Flags().StringVarP(&scanKeyFilterRegexStr, "key-regex", "k", "", "regex to filter message keys by")
	scanCmd.Flags().UintVarP(&scanNumMessages, "num-records", "n", scan.ScanNumRecordsDefault, "maximum number of messages to scan")
	scanCmd.Flags().BoolVarP(&scanSaveMessages, "save-messages", "s", false, "whether to save kafka messages to the local filesystem")
	scanCmd.Flags().BoolVarP(&scanDecode, "decode", "d", true, "whether to decode message values (false is equivalent to 'encodingFormat: raw' in kplay's config)")
	scanCmd.Flags().UintVarP(&scanBatchSize, "batch-size", "b", 100, "number of messages to fetch per batch (must be greater than 0)")
	scanCmd.Flags().StringVarP(&outputDir, "output-dir", "O", defaultOutputDir, "directory to save scan results in")

	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(scanCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	return rootCmd, nil
}
