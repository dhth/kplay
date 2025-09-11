package scan

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/dhth/kplay/internal/fs"
	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"
	"github.com/dhth/kplay/internal/utils"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	errCouldntWriteRecordToFile = errors.New("couldn't write record to file")
	errCouldntInterpretRecord   = errors.New("couldn't interpret kafka record")
)

type Scanner struct {
	client     *kgo.Client
	config     t.Config
	behaviours Behaviours
	homeDir    string
	progress   scanProgress
}

type messageWriter struct {
	file      *os.File
	writer    *bufio.Writer
	csvWriter *csv.Writer
}

type RecordData struct {
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
	Timestamp int64  `json:"timestamp"`
	Key       string `json:"key"`
	Tombstone bool   `json:"tombstone"`
}

type scanProgress struct {
	numRecordsConsumed uint
	numRecordsMatched  uint
	numBytesConsumed   uint64
	lastOffsetSeen     int64
	fsErrors           []fsError
}

func New(client *kgo.Client, config t.Config, behaviours Behaviours, homeDir string) Scanner {
	scanner := Scanner{
		client:     client,
		config:     config,
		behaviours: behaviours,
		homeDir:    homeDir,
	}

	return scanner
}

type fsError struct {
	offset int64
	key    string
	err    error
}

func (s *Scanner) Execute() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	scanErrChan := make(chan error)

	go func(errChan chan<- error) {
		err := s.scan(ctx)
		errChan <- err
	}(scanErrChan)

	select {
	case <-sigChan:
		cancel()
		select {
		case err := <-scanErrChan:
			return err
			// on a second signal
		case <-sigChan:
			return nil
			// timeout after first signal
		case <-time.After(8 * time.Second):
			return t.ErrCouldntShutDownGracefully
		}
	case err := <-scanErrChan:
		return err
	}
}

func (s *Scanner) scan(ctx context.Context) error {
	var recordWriter *messageWriter

	now := time.Now().Unix()
	scanOutputDir := filepath.Join(s.homeDir, ".kplay", "messages", s.config.Topic)

	err := os.MkdirAll(scanOutputDir, 0o755)
	if err != nil {
		return fmt.Errorf("%w: %s", t.ErrCouldntCreateDir, err.Error())
	}

	scanOutputFilePath := filepath.Join(scanOutputDir, fmt.Sprintf("scan-%d.csv", now))

	rw, err := newMessageWriter(scanOutputFilePath)
	if err != nil {
		return err
	}

	defer func() {
		_ = rw.close()
	}()

	recordWriter = rw

	progressChan := make(chan scanProgress, 1)

	go showSpinner(ctx, progressChan)

	for s.progress.numRecordsConsumed < s.behaviours.NumMessages {
		select {
		case <-ctx.Done():
			return s.reportResults(scanOutputDir, scanOutputFilePath)
		default:
		}

		var toFetch uint
		batchSize := s.behaviours.BatchSize
		if s.behaviours.NumMessages < batchSize {
			toFetch = s.behaviours.NumMessages
		} else if s.progress.numRecordsConsumed <= s.behaviours.NumMessages-batchSize {
			toFetch = batchSize
		} else {
			toFetch = s.behaviours.NumMessages - s.progress.numRecordsConsumed
		}

		fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		records := k.FetchRecords(fetchCtx, s.client, toFetch)
		cancel()

		if len(records) == 0 {
			continue
		}

		lastRecord := records[len(records)-1]

		decode := s.behaviours.SaveMessages && s.behaviours.Decode
		for _, record := range records {
			msg := t.GetMessageFromRecord(record, s.config, decode)
			if msg.Err != nil {
				return fmt.Errorf("%w: %s", errCouldntInterpretRecord, msg.Err.Error())
			}

			keyMatches := s.behaviours.KeyFilterRegex != nil && s.behaviours.KeyFilterRegex.MatchString(msg.Key)
			if keyMatches {
				s.progress.numRecordsMatched++
			}

			if recordWriter != nil && (s.behaviours.KeyFilterRegex == nil || keyMatches) {
				err := recordWriter.writeMsg(msg)
				if err != nil {
					return fmt.Errorf("%w: %s", errCouldntWriteRecordToFile, err.Error())
				}
			}

			if s.behaviours.SaveMessages {
				filePath := filepath.Join(
					scanOutputDir,
					fmt.Sprintf("partition-%d", msg.Partition),
					fmt.Sprintf("offset-%d.txt", msg.Offset),
				)

				err := fs.SaveMessageToFileSystem(msg, filePath)
				if err != nil {
					s.progress.fsErrors = append(s.progress.fsErrors, fsError{offset: msg.Offset, key: msg.Key, err: err})
				}
			}

			s.progress.numBytesConsumed += uint64(len(record.Value))
			s.progress.lastOffsetSeen = lastRecord.Offset
		}

		s.progress.numRecordsConsumed += uint(len(records))

		progressChan <- s.progress

	}

	return s.reportResults(scanOutputDir, scanOutputFilePath)
}

func (s *Scanner) reportResults(scanOutputDir, scanOutputFilePath string) error {
	fmt.Fprint(os.Stderr, "\r\033[K")

	if s.progress.numRecordsConsumed == 0 {
		return nil
	}

	fmt.Printf(`Summary:

Scan Results File:             %s
Number of messages scanned:    %d
Value bytes consumed:          %s
`, scanOutputFilePath, s.progress.numRecordsConsumed, utils.HumanReadableBytes(s.progress.numBytesConsumed))

	if s.behaviours.KeyFilterRegex != nil {
		fmt.Printf("Number of matches:             %d\n", s.progress.numRecordsMatched)
	}

	if s.behaviours.SaveMessages && len(s.progress.fsErrors) < int(s.progress.numRecordsConsumed) {
		fmt.Printf("Messages saved in:             %s\n", scanOutputDir)
	}

	if len(s.progress.fsErrors) > 0 {
		errStrs := make([]string, len(s.progress.fsErrors))
		for i, err := range s.progress.fsErrors {
			errStrs[i] = fmt.Sprintf("- offset: %d, key: %s, error: %s", err.offset, err.key, err.err.Error())
		}

		return fmt.Errorf("encountered the following errors while saving values to the local filesystem:\n%s", strings.Join(errStrs, "\n"))
	}

	return nil
}

func newMessageWriter(filePath string) (*messageWriter, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}

	rw := &messageWriter{
		file:   file,
		writer: bufio.NewWriter(file),
	}

	rw.csvWriter = csv.NewWriter(rw.writer)
	err = rw.csvWriter.Write([]string{"partition", "offset", "timestamp", "key", "tombstone"})
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	return rw, nil
}

func (rw *messageWriter) writeMsg(msg t.Message) error {
	return rw.writeCSV(msg)
}

func (rw *messageWriter) writeCSV(msg t.Message) error {
	tombstone := "false"
	if msg.Value == nil {
		tombstone = "true"
	}

	return rw.csvWriter.Write([]string{
		fmt.Sprintf("%d", msg.Partition),
		fmt.Sprintf("%d", msg.Offset),
		msg.Timestamp,
		msg.Key,
		tombstone,
	})
}

func (rw *messageWriter) close() error {
	if rw.csvWriter != nil {
		rw.csvWriter.Flush()
	}
	if rw.writer != nil {
		rw.writer.Flush()
	}
	if rw.file != nil {
		return rw.file.Close()
	}
	return nil
}

func showSpinner(ctx context.Context, progressChan chan scanProgress) {
	var progress scanProgress
	spinnerRunes := []rune{'⣷', '⣯', '⣟', '⡿', '⢿', '⣻', '⣽', '⣾'}
	spinnerIndex := 0

	for {
		select {
		case <-ctx.Done():
			return
		case p := <-progressChan:
			progress = p
		default:
			if spinnerIndex >= len(spinnerRunes)-1 {
				spinnerIndex = 0
			}

			spinnerRune := spinnerRunes[spinnerIndex]
			if progress.numRecordsConsumed == 0 {
				fmt.Fprintf(os.Stderr, "\r\033[K%c scanning...", spinnerRune)
			} else {
				bytesConsumed := utils.HumanReadableBytes(progress.numBytesConsumed)
				if progress.numRecordsMatched > 0 {
					fmt.Fprintf(os.Stderr, "\r\033[K%c %d messages scanned; %d match filter (last offset: %d, value bytes consumed: %s)",
						spinnerRune,
						progress.numRecordsConsumed,
						progress.numRecordsMatched,
						progress.lastOffsetSeen,
						bytesConsumed,
					)
				} else {
					fmt.Fprintf(os.Stderr, "\r\033[K%c %d messages scanned (last offset: %d, value bytes consumed: %s)",
						spinnerRune,
						progress.numRecordsConsumed,
						progress.lastOffsetSeen,
						bytesConsumed,
					)
				}
			}

			spinnerIndex++

			select {
			case <-ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
}
