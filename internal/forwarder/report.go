package forwarder

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"
)

var headers = []string{"topic", "partition", "offset", "timestamp", "key", "tombstone", "decode_error", "upload_error"}

type reportWriter struct {
	csvWriter *csv.Writer
	buffer    *bytes.Buffer
	numMsgs   uint16
	startTime time.Time
}

func newReportWriter() *reportWriter {
	buffer := &bytes.Buffer{}
	csvWriter := csv.NewWriter(buffer)

	writer := reportWriter{
		csvWriter: csvWriter,
		buffer:    buffer,
		startTime: time.Now(),
	}

	_ = csvWriter.Write(headers)

	return &writer
}

func (w *reportWriter) writeRow(result uploadResult) error {
	msg := result.work.msg

	tombstone := "false"
	if len(msg.Value) == 0 {
		tombstone = "true"
	}

	var decodeErr string
	if msg.DecodeErr != nil {
		decodeErr = msg.DecodeErr.Error()
	}
	var uploadErr string
	if result.err != nil {
		uploadErr = result.err.Error()
	}

	row := []string{
		msg.Topic,
		fmt.Sprintf("%d", msg.Partition),
		fmt.Sprintf("%d", msg.Offset),
		msg.Timestamp.Format(time.RFC3339),
		msg.Key,
		tombstone,
		decodeErr,
		uploadErr,
	}

	err := w.csvWriter.Write(row)
	if err != nil {
		return err
	}
	w.numMsgs++

	return nil
}

func (w *reportWriter) getContent() ([]byte, error) {
	w.csvWriter.Flush()
	flushErr := w.csvWriter.Error()
	if flushErr != nil {
		return nil, flushErr
	}

	content := w.buffer.Bytes()
	w.reset()

	return content, nil
}

func (w *reportWriter) reset() {
	w.buffer.Reset()
	w.csvWriter = csv.NewWriter(w.buffer)
	w.numMsgs = 0
	w.startTime = time.Now()

	_ = w.csvWriter.Write(headers)
}
