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

	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"
	"github.com/twmb/franz-go/pkg/kgo"
)

var errCouldntWriteRecordToFile = errors.New("couldn't write record to file")

const (
	ScanFormatTXT            = "txt"
	ScanReportFmtTable       = "table"
	ScanNumRecordsDefault    = 100
	ScanNumRecordsUpperBound = 10000
)

type scanFormat uint8

const (
	ScanFormatCSV scanFormat = iota
	ScanFormatJSONL
	ScanFormatTxt
)

type Scanner struct {
	client     *kgo.Client
	config     t.Config
	behaviours Behaviours
}

type RecordWriter struct {
	file      *os.File
	writer    *bufio.Writer
	csvWriter *csv.Writer
	format    scanFormat
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

func (s *Scanner) Execute() error {
	var recordWriter *RecordWriter

	rw, err := newRecordWriter(s.behaviours.OutPathFull)
	if err != nil {
		return err
	}

	defer rw.close()

	recordWriter = rw

	spinnerDone := make(chan struct{})
	progressChan := make(chan scanProgress)
	var numConsumed uint

	go showSpinner(spinnerDone, progressChan)

	var numMatched uint

	for numConsumed < s.behaviours.NumRecords {

		ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
		defer cancel()

		var toFetch uint
		batchSize := s.behaviours.BatchSize
		if s.behaviours.NumRecords < batchSize {
			toFetch = s.behaviours.NumRecords
		} else if numConsumed <= s.behaviours.NumRecords-batchSize {
			toFetch = batchSize
		} else {
			toFetch = s.behaviours.NumRecords - numConsumed
		}

		records := k.FetchRecords(ctx, s.client, toFetch)

		if len(records) == 0 {
			continue
		}

		lastRecord := records[len(records)-1]

		for _, record := range records {
			keyMatches := s.behaviours.KeyFilterRegex != nil && s.behaviours.KeyFilterRegex.Match(record.Key)
			if keyMatches {
				numMatched++
			}

			if recordWriter != nil && (s.behaviours.KeyFilterRegex == nil || keyMatches) {
				err := recordWriter.writeRecord(record)
				if err != nil {
					spinnerDone <- struct{}{}
					return fmt.Errorf("%w: %s", errCouldntWriteRecordToFile, err.Error())
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
			if numConsumed > 0 {
				fmt.Printf("%d records matching key filter written to %s\n", numMatched, s.behaviours.OutPathFull)
			} else {
				fmt.Println("no records matched key filter")
			}
		} else {
			fmt.Printf("%d records written to %s\n", numConsumed, s.behaviours.OutPathFull)
		}
	}

	return nil
}

func inferFormatFromPath(filePath string) (scanFormat, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".csv":
		return ScanFormatCSV, nil
	case ".jsonl":
		return ScanFormatJSONL, nil
	case ".txt":
		return ScanFormatTxt, nil
	default:
		return ScanFormatCSV, fmt.Errorf("unsupported file extension: %q (supported: csv, jsonl, txt)", ext)
	}
}

func newRecordWriter(filePath string) (*RecordWriter, error) {
	format, err := inferFormatFromPath(filePath)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}

	rw := &RecordWriter{
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

func (rw *RecordWriter) writeRecord(record *kgo.Record) error {
	switch rw.format {
	case ScanFormatCSV:
		return rw.writeCSVRecord(record)
	case ScanFormatJSONL:
		return rw.writeJSONLRecord(record)
	case ScanFormatTxt:
		return rw.writeTXTRecord(record)
	default:
		return rw.writeTXTRecord(record)
	}
}

func (rw *RecordWriter) writeCSVRecord(record *kgo.Record) error {
	tombstone := "false"
	if record.Value == nil {
		tombstone = "true"
	}

	return rw.csvWriter.Write([]string{
		fmt.Sprintf("%d", record.Partition),
		fmt.Sprintf("%d", record.Offset),
		record.Timestamp.Format(time.RFC3339),
		string(record.Key),
		tombstone,
	})
}

func (rw *RecordWriter) writeJSONLRecord(record *kgo.Record) error {
	recordData := RecordData{
		Partition: record.Partition,
		Offset:    record.Offset,
		Timestamp: record.Timestamp.UnixMilli(),
		Key:       string(record.Key),
		Tombstone: record.Value == nil,
	}

	encoder := json.NewEncoder(rw.writer)
	return encoder.Encode(recordData)
}

func (rw *RecordWriter) writeTXTRecord(record *kgo.Record) error {
	tombstone := "false"
	if record.Value == nil {
		tombstone = "true"
	}

	line := fmt.Sprintf("partition=%d offset=%d timestamp=%s key=%s tombstone=%s",
		record.Partition,
		record.Offset,
		record.Timestamp.Format(time.RFC3339),
		string(record.Key),
		tombstone,
	)

	_, err := fmt.Fprintln(rw.writer, line)
	return err
}

func (rw *RecordWriter) close() error {
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
				fmt.Fprintf(os.Stderr, "\r\033[K%c fetching records...", spinnerRune)
			} else {
				if progress.numRecordsMatched > 0 {
					fmt.Fprintf(os.Stderr, "\r\033[K%c %d records fetched; %d match filter (last offset: %d)",
						spinnerRune,
						progress.numRecordsConsumed,
						progress.numRecordsMatched,
						progress.lastOffsetSeen,
					)
				} else {
					fmt.Fprintf(os.Stderr, "\r\033[K%c %d records fetched (last offset: %d)",
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
