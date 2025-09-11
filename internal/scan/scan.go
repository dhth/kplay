package scan

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dhth/kplay/internal/fs"
	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"
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
}

type messageWriter struct {
	file      *os.File
	writer    *bufio.Writer
	csvWriter *csv.Writer
	format    Format
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
	lastOffsetSeen     int64
}

func New(client *kgo.Client, config t.Config, behaviours Behaviours) Scanner {
	scanner := Scanner{
		client:     client,
		config:     config,
		behaviours: behaviours,
	}

	return scanner
}

type fsError struct {
	offset int64
	key    string
	err    error
}

func (s *Scanner) Execute() error {
	var recordWriter *messageWriter

	now := time.Now().Unix()
	scanOutputDir := filepath.Join(".kplay", "messages", s.config.Topic)

	err := os.MkdirAll(scanOutputDir, 0o755)
	if err != nil {
		return fmt.Errorf("%w: %s", t.ErrCouldntCreateDir, err.Error())
	}

	scanFilePath := filepath.Join(scanOutputDir, fmt.Sprintf("scan-%d.%s", now, s.behaviours.OutputFormat.Extension()))

	rw, err := newMessageWriter(scanFilePath, s.behaviours.OutputFormat)
	if err != nil {
		return err
	}

	defer rw.close()

	recordWriter = rw

	spinnerDone := make(chan struct{})
	progressChan := make(chan scanProgress)
	var numConsumed uint
	var fsErrors []fsError

	go showSpinner(spinnerDone, progressChan)

	var numMatched uint

	for numConsumed < s.behaviours.NumMessages {

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		var toFetch uint
		batchSize := s.behaviours.BatchSize
		if s.behaviours.NumMessages < batchSize {
			toFetch = s.behaviours.NumMessages
		} else if numConsumed <= s.behaviours.NumMessages-batchSize {
			toFetch = batchSize
		} else {
			toFetch = s.behaviours.NumMessages - numConsumed
		}

		records := k.FetchRecords(ctx, s.client, toFetch)
		cancel()

		if len(records) == 0 {
			continue
		}

		lastRecord := records[len(records)-1]

		for _, record := range records {
			msg := t.GetMessageFromRecord(record, s.config, s.behaviours.Decode)
			if msg.Err != nil {
				spinnerDone <- struct{}{}
				return fmt.Errorf("%w: %s", errCouldntInterpretRecord, msg.Err.Error())
			}

			keyMatches := s.behaviours.KeyFilterRegex != nil && s.behaviours.KeyFilterRegex.MatchString(msg.Key)
			if keyMatches {
				numMatched++
			}

			if recordWriter != nil && (s.behaviours.KeyFilterRegex == nil || keyMatches) {
				err := recordWriter.writeMsg(msg)
				if err != nil {
					spinnerDone <- struct{}{}
					return fmt.Errorf("%w: %s", errCouldntWriteRecordToFile, err.Error())
				}
			}

			if s.behaviours.SaveMessages {
				err := fs.SaveMessageToFileSystem(msg, s.config.Topic)
				if err != nil {
					fsErrors = append(fsErrors, fsError{offset: msg.Offset, key: msg.Key, err: err})
				}
			}
		}
		numConsumed += uint(len(records))

		progressChan <- scanProgress{
			numRecordsConsumed: numConsumed,
			numRecordsMatched:  numMatched,
			lastOffsetSeen:     lastRecord.Offset,
		}

	}

	spinnerDone <- struct{}{}

	if numConsumed > 0 {
		if s.behaviours.KeyFilterRegex != nil {
			if numMatched > 0 {
				fmt.Printf("%d messages matching key filter written to %s\n", numMatched, scanFilePath)
			} else {
				fmt.Println("no messages matched key filter")
			}
		} else {
			fmt.Printf("%d messages written to %s\n", numConsumed, scanFilePath)
		}
	}

	if len(fsErrors) > 0 {
		errStrs := make([]string, len(fsErrors))
		for i, err := range fsErrors {
			errStrs[i] = fmt.Sprintf("- offset: %d, key: %s, error: %s", err.offset, err.key, err.err.Error())
		}

		return fmt.Errorf("encountered the following errors while saving values to the local filesystem:\n%s", strings.Join(errStrs, "\n"))
	}

	return nil
}

func newMessageWriter(filePath string, format Format) (*messageWriter, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}

	rw := &messageWriter{
		file:   file,
		writer: bufio.NewWriter(file),
		format: format,
	}

	if format == ScanFormatCSV {
		rw.csvWriter = csv.NewWriter(rw.writer)
		err := rw.csvWriter.Write([]string{"partition", "offset", "timestamp", "key", "tombstone"})
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to write CSV header: %w", err)
		}
	}

	return rw, nil
}

func (rw *messageWriter) writeMsg(msg t.Message) error {
	switch rw.format {
	case ScanFormatCSV:
		return rw.writeCSV(msg)
	case ScanFormatJSONL:
		return rw.writeJSONL(msg)
	case ScanFormatTxt:
		return rw.writeTXT(msg)
	default:
		return rw.writeTXT(msg)
	}
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

func (rw *messageWriter) writeJSONL(msg t.Message) error {
	encoder := json.NewEncoder(rw.writer)
	return encoder.Encode(msg)
}

func (rw *messageWriter) writeTXT(msg t.Message) error {
	tombstone := "false"
	if msg.Value == nil {
		tombstone = "true"
	}

	line := fmt.Sprintf("partition=%d offset=%d timestamp=%s key=%s tombstone=%s",
		msg.Partition,
		msg.Offset,
		msg.Timestamp,
		msg.Key,
		tombstone,
	)

	_, err := fmt.Fprintln(rw.writer, line)
	return err
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

func showSpinner(done chan struct{}, progressChan chan scanProgress) {
	var progress scanProgress
	spinnerRunes := []rune{'⣷', '⣯', '⣟', '⡿', '⢿', '⣻', '⣽', '⣾'}
	spinnerIndex := 0
	for {
		select {
		case <-done:
			fmt.Fprint(os.Stderr, "\r\033[K")
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
				if progress.numRecordsMatched > 0 {
					fmt.Fprintf(os.Stderr, "\r\033[K%c %d messages scanned; %d match filter (last offset: %d)",
						spinnerRune,
						progress.numRecordsConsumed,
						progress.numRecordsMatched,
						progress.lastOffsetSeen,
					)
				} else {
					fmt.Fprintf(os.Stderr, "\r\033[K%c %d messages scanned (last offset: %d)",
						spinnerRune,
						progress.numRecordsConsumed,
						progress.lastOffsetSeen,
					)
				}
			}
			spinnerIndex++
			time.Sleep(100 * time.Millisecond)
		}
	}
}
