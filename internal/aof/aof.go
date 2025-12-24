package aof

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
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
		return nil, fmt.Errorf("failed to open AOF file: %w", err)
	}

	return &AOF{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (a *AOF) Write(operation, key, value string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	line := fmt.Sprintf("%s %s %s\n", operation, key, value)

	if _, err := a.writer.WriteString(line); err != nil {
		return fmt.Errorf("failed to write to AOF: %w", err)
	}

	if err := a.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush to AOF: %w", err)
	}

	if err := a.file.Sync(); err != nil {
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
			return []Operation{}, nil
		}
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
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}
		operations = append(operations, op)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading AOF: %w", err)
	}
	return operations, nil
}

func parseOperation(line string) (Operation, error) {
	parts := strings.SplitN(line, " ", 3)

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
		op.Value = parts[3]
	}
	return op, nil
}
