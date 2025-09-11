package fs

import (
	"fmt"
	"os"
	"path/filepath"

	t "github.com/dhth/kplay/internal/types"
)

func SaveMessageToFileSystem(msg t.Message, path string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return fmt.Errorf("%w: %w", t.ErrCouldntCreateDir, err)
	}
	details := msg.GetDetails()

	err = os.WriteFile(path, []byte(details), 0o644)
	if err != nil {
		return fmt.Errorf("%w: %w", t.ErrCouldntWriteToFile, err)
	}

	return nil
}
