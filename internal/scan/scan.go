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

	"github.com/charmbracelet/lipgloss"
	"github.com/dhth/kplay/internal/fs"
	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"
	"github.com/dhth/kplay/internal/utils"
	"github.com/twmb/franz-go/pkg/kgo"
)

var errCouldntWriteRecordToFile = errors.New("couldn't write record to file")

type Scanner struct {
	client     *kgo.Client
	config     t.Config
	behaviours Behaviours
	outputDir  string
	progress   scanProgress
}

type messageWriter struct {
	file      *os.File
	writer    *bufio.Writer
	csvWriter *csv.Writer
}

type scanProgress struct {
	numRecordsConsumed uint
	numRecordsMatched  uint
	numBytesConsumed   uint64
	lastOffsetDetails  string
	lastTimeStampSeen  time.Time
	numDecodeErrors    uint
	fsErrors           []fsError
}

func New(client *kgo.Client, config t.Config, behaviours Behaviours, outputDir string) Scanner {
	scanner := Scanner{
		client:     client,
		config:     config,
		behaviours: behaviours,
		outputDir:  outputDir,
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
		case <-time.After(5 * time.Second):
			return t.ErrCouldntShutDownGracefully
		}
	case err := <-scanErrChan:
		return err
	}
}

func (s *Scanner) scan(ctx context.Context) error {
	var recordWriter *messageWriter

	now := time.Now().Unix()
	scanOutputDir := filepath.Join(s.outputDir, "messages", s.config.Topic)

	err := os.MkdirAll(scanOutputDir, 0o755)
	if err != nil {
		return fmt.Errorf("%w: %s", t.ErrCouldntCreateDir, err.Error())
	}

	scanOutputFilePath := filepath.Join(scanOutputDir, fmt.Sprintf("scan-%d.csv", now))

	decode := s.behaviours.SaveMessages && s.behaviours.Decode

	rw, err := newMessageWriter(scanOutputFilePath, decode)
	if err != nil {
		return err
	}

	defer func() {
		_ = rw.close()
	}()

	recordWriter = rw

	progressChan := make(chan scanProgress, 1)
	spinnerDone := make(chan struct{})

	go showSpinner(spinnerDone, progressChan, s.behaviours)

	defer func() {
		spinnerDone <- struct{}{}
		close(spinnerDone)
		close(progressChan)
		s.reportResults(scanOutputDir, scanOutputFilePath)
	}()

	for s.progress.numRecordsConsumed < s.behaviours.NumMessages {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		toFetch := min(s.behaviours.NumMessages-s.progress.numRecordsConsumed, s.behaviours.BatchSize)

		fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		records, err := k.FetchRecords(fetchCtx, s.client, toFetch)
		cancel()

		if err != nil {
			return err
		}

		if len(records) == 0 {
			continue
		}

		lastRecord := records[len(records)-1]

		for _, record := range records {
			if record == nil {
				continue
			}

			msg := t.GetMessageFromRecord(*record, s.config, decode)
			if msg.DecodeErr != nil {
				s.progress.numDecodeErrors++
			}

			keyMatches := s.behaviours.KeyFilterRegex != nil && s.behaviours.KeyFilterRegex.MatchString(msg.Key)
			if keyMatches {
				s.progress.numRecordsMatched++
			}

			saveMsg := s.behaviours.KeyFilterRegex == nil || keyMatches

			if recordWriter != nil && saveMsg {
				err := recordWriter.writeMsg(msg, decode)
				if err != nil {
					return fmt.Errorf("%w: %s", errCouldntWriteRecordToFile, err.Error())
				}
			}

			if s.behaviours.SaveMessages && saveMsg {
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
			s.progress.lastOffsetDetails = fmt.Sprintf("%d:%d", lastRecord.Partition, lastRecord.Offset)
			s.progress.lastTimeStampSeen = lastRecord.Timestamp
		}

		s.progress.numRecordsConsumed += uint(len(records))

		progressChan <- s.progress

	}

	return nil
}

func (s *Scanner) reportResults(scanOutputDir, scanOutputFilePath string) {
	fmt.Fprint(os.Stderr, "\r\033[K")

	if s.progress.numRecordsConsumed == 0 {
		return
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

	if s.progress.numDecodeErrors > 0 {
		fmt.Printf("Decode errors:                 %d\n", s.progress.numDecodeErrors)
	}

	if len(s.progress.fsErrors) > 0 {
		errStrs := make([]string, len(s.progress.fsErrors))
		for i, err := range s.progress.fsErrors {
			errStrs[i] = fmt.Sprintf("- offset: %d, key: %s, error: %s", err.offset, err.key, err.err.Error())
		}

		fmt.Printf("\nEncountered the following errors while saving values to the local filesystem:\n%s", strings.Join(errStrs, "\n"))
	}
}

func newMessageWriter(filePath string, decode bool) (*messageWriter, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}

	rw := &messageWriter{
		file:   file,
		writer: bufio.NewWriter(file),
	}

	rw.csvWriter = csv.NewWriter(rw.writer)
	headers := []string{"partition", "offset", "timestamp", "key", "tombstone"}
	if decode {
		headers = append(headers, "decode_success")
	}

	err = rw.csvWriter.Write(headers)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	return rw, nil
}

func (rw *messageWriter) writeMsg(msg t.Message, decode bool) error {
	return rw.writeCSV(msg, decode)
}

func (rw *messageWriter) writeCSV(msg t.Message, decode bool) error {
	tombstone := "false"
	if msg.Value == nil {
		tombstone = "true"
	}

	row := []string{
		fmt.Sprintf("%d", msg.Partition),
		fmt.Sprintf("%d", msg.Offset),
		msg.Timestamp,
		msg.Key,
		tombstone,
	}

	if decode {
		if msg.DecodeErr == nil {
			row = append(row, "true")
		} else {
			row = append(row, "false")
		}
	}

	return rw.csvWriter.Write(row)
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

func showSpinner(doneChan chan struct{}, progressChan chan scanProgress, behaviours Behaviours) {
	var progress scanProgress
	spinnerRunes := []rune{'⣷', '⣯', '⣟', '⡿', '⢿', '⣻', '⣽', '⣾'}
	spinnerIndex := 0

	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#282828"))
	numRecordsStyle := highlightStyle.Background(lipgloss.Color("#fabd2f"))
	offsetStyle := highlightStyle.Background(lipgloss.Color("#83a598"))
	timestampStyle := highlightStyle.Background(lipgloss.Color("#d3869b"))
	bytesStyle := highlightStyle.Background(lipgloss.Color("#8ec07c"))
	errorStyle := highlightStyle.Background(lipgloss.Color("#fb4934"))

	progressLine := "scanning..."

	for {
		select {
		case <-doneChan:
			fmt.Fprint(os.Stderr, "\r\033[K")
			return
		case p := <-progressChan:
			progress = p

			bytesConsumed := utils.HumanReadableBytes(progress.numBytesConsumed)
			var matchInfo string
			if behaviours.KeyFilterRegex != nil {
				matchInfo = fmt.Sprintf("; %d matches", progress.numRecordsMatched)
			}

			var decodeErrorsSection string
			if progress.numDecodeErrors > 0 {
				decodeErrorsSection = fmt.Sprintf(", decode errors: %s", errorStyle.Render(fmt.Sprintf("%d", progress.numDecodeErrors)))
			}
			progressLine = fmt.Sprintf("%s messages scanned%s (offset: %s, timestamp: %s, bytes consumed: %s%s)",
				numRecordsStyle.Render(fmt.Sprintf("%d", progress.numRecordsConsumed)),
				matchInfo,
				offsetStyle.Render(progress.lastOffsetDetails),
				timestampStyle.Render(progress.lastTimeStampSeen.Format(time.RFC3339)),
				bytesStyle.Render(bytesConsumed),
				decodeErrorsSection,
			)
		default:
			if spinnerIndex >= len(spinnerRunes)-1 {
				spinnerIndex = 0
			}

			spinnerRune := spinnerRunes[spinnerIndex]
			fmt.Fprintf(os.Stderr, "\r\033[K%c %s", spinnerRune, progressLine)

			spinnerIndex++

			select {
			case <-doneChan:
				fmt.Fprint(os.Stderr, "\r\033[K")
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
}
