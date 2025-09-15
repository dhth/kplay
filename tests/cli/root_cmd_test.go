package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skipIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("Skipping integration tests")
	}
}

func TestCLI(t *testing.T) {
	skipIntegration(t)

	tempDir, err := os.MkdirTemp("", "")
	require.NoErrorf(t, err, "error creating temporary directory: %s", err)

	binPath := filepath.Join(tempDir, "kplay")
	buildArgs := []string{"build", "-o", binPath, "../.."}

	c := exec.Command("go", buildArgs...)
	err = c.Run()
	require.NoErrorf(t, err, "error building binary: %s", err)

	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			fmt.Printf("couldn't clean up temporary directory (%s): %s", binPath, err)
		}
	}()

	// SUCCESSES
	t.Run("Help", func(t *testing.T) {
		// GIVEN
		// WHEN
		c := exec.Command(binPath, "-h")
		b, err := c.CombinedOutput()

		// THEN
		assert.NoError(t, err, "output:\n%s", b)
	})

	t.Run("Parsing correct config works", func(t *testing.T) {
		// GIVEN
		// WHEN
		configPath := "assets/config-correct.yml"
		c := exec.Command(binPath, "tui", "local", "-c", configPath, "--debug")
		o, err := c.CombinedOutput()
		// THEN
		if err != nil {
			fmt.Printf("output:\n%s", o)
		}
		assert.NoError(t, err, "output:\n%s", o)
	})

	t.Run("Parsing profile with raw encoding works", func(t *testing.T) {
		// GIVEN
		// WHEN
		configPath := "assets/config-raw-encoding.yml"
		c := exec.Command(binPath, "tui", "local", "-c", configPath, "--debug")
		o, err := c.CombinedOutput()
		// THEN
		if err != nil {
			fmt.Printf("output:\n%s", o)
		}
		assert.NoError(t, err, "output:\n%s", o)
	})

	// FAILURES
	t.Run("Fails for absent config file", func(t *testing.T) {
		// GIVEN
		// WHEN
		configPath := "assets/non-existent-file.yml"
		c := exec.Command(binPath, "tui", "local", "-c", configPath, "--debug")
		o, err := c.CombinedOutput()

		// THEN
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode := exitError.ExitCode()
			require.Equal(t, 1, exitCode, "exit code is not correct: got %d, expected: 1; output:\n%s", exitCode, o)
			assert.Contains(t, string(o), "couldn't read config file")
		} else {
			t.Fatalf("couldn't get error code")
		}
	})

	t.Run("Parsing incorrect config fails", func(t *testing.T) {
		// GIVEN
		// WHEN
		configPath := "assets/config-incorrect-yml.yml"
		c := exec.Command(binPath, "tui", "local", "-c", configPath, "--debug")
		o, err := c.CombinedOutput()

		// THEN
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode := exitError.ExitCode()
			require.Equal(t, 1, exitCode, "exit code is not correct: got %d, expected: 1; output:\n%s", exitCode, o)
			assert.Contains(t, string(o), "couldn't parse config file")
		} else {
			t.Fatalf("couldn't get error code")
		}
	})

	t.Run("Fails if descriptor name incorrect", func(t *testing.T) {
		// GIVEN
		// WHEN
		configPath := "assets/config-protobuf-incorrect-desc-name.yml"
		c := exec.Command(binPath, "tui", "local", "-c", configPath, "--debug")
		o, err := c.CombinedOutput()

		// THEN
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode := exitError.ExitCode()
			require.Equal(t, 1, exitCode, "exit code is not correct: got %d, expected: 1; output:\n%s", exitCode, o)
			assert.Contains(t, string(o), "descriptor name is invalid")
		} else {
			t.Fatalf("couldn't get error code")
		}
	})

	t.Run("Fails if descriptor set incorrect", func(t *testing.T) {
		// GIVEN
		// WHEN
		configPath := "assets/config-protobuf-incorrect-desc-set.yml"
		c := exec.Command(binPath, "tui", "local", "-c", configPath, "--debug")
		o, err := c.CombinedOutput()

		// THEN
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode := exitError.ExitCode()
			require.Equal(t, 1, exitCode, "exit code is not correct: got %d, expected: 1; output:\n%s", exitCode, o)
			assert.Contains(t, string(o), "there's an issue with the file descriptor set")
		} else {
			t.Fatalf("couldn't get error code")
		}
	})
}
