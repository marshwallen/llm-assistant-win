package utils

import (
	"encoding/csv"
	"os"
	"sync"
)

type CSVWriter struct {
	file    *os.File
	writer  *csv.Writer
	mu      sync.Mutex
}

func NewCSVWriter(filename string) (*CSVWriter, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	w := &CSVWriter{
		file:   file,
		writer: csv.NewWriter(file),
	}

	return w, nil
}

func (w *CSVWriter) Write(record []string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(record)
}

func (w *CSVWriter) Flush() {
	w.writer.Flush()
}

func (w *CSVWriter) Close() {
	w.Flush()
	w.file.Close()
}