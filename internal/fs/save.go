package fs

import (
	"fmt"
	"os"
	"path/filepath"

	t "github.com/dhth/kplay/internal/types"
)

func SaveMessageToFileSystem(msg t.Message, topic string) error {
	filePath := filepath.Join(".kplay", "messages",
		topic,
		fmt.Sprintf("partition-%d", msg.Partition),
		fmt.Sprintf("offset-%d.txt", msg.Offset),
	)

	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return fmt.Errorf("%w: %w", t.ErrCouldntCreateDir, err)
	}
	details := msg.GetDetails()

	err = os.WriteFile(filePath, []byte(details), 0o644)
	if err != nil {
		return fmt.Errorf("%w: %w", t.ErrCouldntWriteToFile, err)
	}

	return nil
}
