package aof

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/mrpurushotam/mini_database/internal/logger"
)

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

func (a *AOF) Write(operation, key, value string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	line := fmt.Sprintf("%s %s %s\n", operation, key, value)
	logger.Debug("writing to AOF", "operation", operation, "key", key)

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
	Type  string
	Key   string
	Value string
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
	parts := strings.SplitN(line, " ", 3)
	logger.Debug("parsing operation", "line", line, "partsCount", len(parts))

	if len(parts) < 2 {
		return Operation{}, fmt.Errorf("invalid operation format %s", line)
	}

	op := Operation{
		Type: parts[0],
		Key:  parts[1],
	}

	if op.Type == "SET" {
		if len(parts) != 3 {
			return Operation{}, fmt.Errorf("SET operation missing value: %s", line)
		}
		op.Value = parts[2]
	}
	logger.Debug("operation parsed", "type", op.Type, "key", op.Key, "valueLength", len(op.Value))
	return op, nil
}
