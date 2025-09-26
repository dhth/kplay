package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	t "github.com/dhth/kplay/internal/types"
	"github.com/dhth/kplay/internal/utils"
	"github.com/spf13/cobra"
)

const (
	configFileName   = "kplay/kplay.yml"
	envVarConfigPath = "KPLAY_CONFIG_PATH"
)

func Execute(version string) error {
	rootCmd, err := NewRootCommand(version)
	if err != nil {
		return err
	}

	return rootCmd.Execute()
}

func NewRootCommand(version string) (*cobra.Command, error) {
	var (
		configPath        string
		configPathFull    string
		homeDir           string
		outputDir         string
		fromOffset        string
		fromTimestamp     string
		debug             bool
		config            t.Config
		consumeBehaviours t.ConsumeBehaviours
	)

	preRunE := func(cmd *cobra.Command, args []string) error {
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
	}

	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntGetUserHomeDir, err.Error())
	}

	defaultOutputDir := filepath.Join(homeDir, ".kplay")

	rootCmd := &cobra.Command{
		Use:   "kplay",
		Short: "kplay lets you inspect messages in a Kafka topic in a simple and deliberate manner.",
		Long: `kplay ("kafka playground") lets you inspect messages in a Kafka topic in a simple and deliberate manner.

kplay relies on a configuration file that contains profiles for various Kafka topics, each with its own details related
to brokers, message encoding, authentication, etc.
`,
		SilenceErrors: true,
		Version:       version,
	}

	tuiCmd := newTuiCmd(
		preRunE,
		&config,
		&consumeBehaviours,
		&fromOffset,
		&fromTimestamp,
		&outputDir,
		&debug,
		defaultOutputDir,
	)

	serveCmd := newServeCmd(
		preRunE,
		&config,
		&consumeBehaviours,
		&fromOffset,
		&fromTimestamp,
		&debug,
	)

	scanCmd := newScanCmd(
		preRunE,
		&config,
		&consumeBehaviours,
		&fromOffset,
		&fromTimestamp,
		&outputDir,
		&debug,
		defaultOutputDir,
	)

	forwardCmd := newForwardCmd(&configPath, homeDir, &debug, version)

	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntGetUserConfigDir, err.Error())
	}

	defaultConfigPath := filepath.Join(configDir, configFileName)

	rootCmd.PersistentFlags().StringVarP(&configPath, "config-path", "c", defaultConfigPath, fmt.Sprintf("location of kplay's config file (can also be provided via $%s)", envVarConfigPath))
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "whether to only display config picked up by kplay without running it")

	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(forwardCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	return rootCmd, nil
}
