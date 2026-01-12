package aof

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/mrpurushotam/mini_db/internal/domain"
	"github.com/mrpurushotam/mini_db/internal/logger"
)

// StoreReader defines what AOF needs from the store (no import!)
type StoreReader interface {
	GetAll() map[string]domain.Value
}

type AOF struct {
	file   *os.File
	writer *bufio.Writer
	mu     sync.Mutex
}

// NewAOF Creates / open filepath
func NewAOF(filepath string) (*AOF, error) {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("failed to open AOF file", "filepath", filepath, "error", err)
		return nil, fmt.Errorf("failed to open AOF file: %w", err)
	}

	aof := &AOF{
		file:   file,
		writer: bufio.NewWriter(file),
	}

	logger.Info("AOF initialized", "filepath", filepath)
	return aof, nil
}

// Snapshot create a NewAof file with current state
func (a *AOF) Snapshot(store StoreReader) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	logger.Info("Building AOF Snapshot...")

	// Create a new temp file
	originalPath := a.file.Name()
	tempPath := originalPath + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp AOF: %w", err)
	}
	defer tempFile.Close()

	tempWriter := bufio.NewWriter(tempFile)
	snapshot := store.GetAll()

	for key, value := range snapshot {
		line := fmt.Sprintf("SET %s %s %s\n",
			key,
			value.Type(),
			string(value.Serialize()))

		if _, err := tempWriter.WriteString(line); err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return fmt.Errorf("failed to write snapshot: %w", err)
		}
	}

	if err := tempWriter.Flush(); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return err
	}

	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return err
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to close temp AOF: %w", err)
	}

	if a.file != nil {
		_ = a.writer.Flush()
		_ = a.file.Close()
	}

	// remove original if it exists (allow rename to succeed on Windows)
	if err := os.Remove(originalPath); err != nil && !os.IsNotExist(err) {
		os.Remove(tempPath)
		return fmt.Errorf("failed to remove original AOF: %w", err)
	}

	if err := os.Rename(tempPath, originalPath); (err) != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp AOF: %w", err)
	}

	a.file, err = os.OpenFile(originalPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to reopen AOF: %w", err)
	}
	a.writer = bufio.NewWriter(a.file)

	logger.Info("AOF snapshot completed successfully")
	return nil
}

func (a *AOF) Write(operation, key, valueType, value string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	line := fmt.Sprintf("%s %s %s %s\n", operation, key, valueType, value)
	logger.Debug("writing to AOF", "operation", operation, "key", key, "valueType", valueType)

	if _, err := a.writer.WriteString(line); err != nil {
		logger.Error("failed to write to AOF", "error", err)
		return fmt.Errorf("failed to write to AOF: %w", err)
	}

	if err := a.writer.Flush(); err != nil {
		logger.Error("failed to flush to AOF", "error", err)
		return fmt.Errorf("failed to flush to AOF: %w", err)
	}

	if err := a.file.Sync(); err != nil {
		logger.Error("failed to sync AOF", "error", err)
		return fmt.Errorf("failed to sync AOF: %w", err)
	}

	return nil
}

func (a *AOF) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.writer.Flush(); err != nil {
		return err
	}

	return a.file.Close()
}

type Operation struct {
	Type      string
	Key       string
	ValueType string
	Value     string
}

func (a *AOF) Read(filepath string) ([]Operation, error) {
	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("AOF file not found, starting fresh", "filepath", filepath)
			return []Operation{}, nil
		}
		logger.Info("AOF file not found, starting fresh", "filepath", filepath)
		return nil, fmt.Errorf("failed to open AOF for reading: %w", err)
	}
	defer file.Close()

	var operations []Operation
	scanner := bufio.NewScanner(file)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if strings.TrimSpace(line) == "" {
			continue
		}

		op, err := parseOperation(line)
		if err != nil {
			logger.Error("failed to parse AOF line", "line", lineNum, "error", err, "content", line)
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}
		operations = append(operations, op)
	}

	if err := scanner.Err(); err != nil {
		logger.Error("error reading AOF", "error", err)
		return nil, fmt.Errorf("error reading AOF: %w", err)
	}
	logger.Info("AOF loaded successfully", "operationsCount", len(operations))
	return operations, nil
}

func parseOperation(line string) (Operation, error) {
	parts := strings.SplitN(line, " ", 4)
	logger.Debug("parsing operation", "line", line, "partsCount", len(parts))

	if len(parts) < 2 {
		return Operation{}, fmt.Errorf("invalid operation format %s", line)
	}

	op := Operation{
		Type:      parts[0],
		Key:       parts[1],
		ValueType: parts[2],
		Value:     parts[3],
	}

	if op.Type == "SET" {
		if len(parts) != 4 {
			return Operation{}, fmt.Errorf("SET operation missing value type or value: %s", line)
		}
	}
	logger.Debug("operation parsed", "type", op.Type, "key", op.Key, "valueType", op.ValueType, "valueLength", len(op.Value))
	return op, nil
}
