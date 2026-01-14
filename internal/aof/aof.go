package aof

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

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

type Operation struct {
	Type      string `json:"op"`
	Key       string `json:"key"`
	ValueType string `json:"valueType"`
	Value     string `json:"value"`
}

type AOFHeader struct {
	Format   string `json:"format"`
	Version  string `json:"version"`
	Encoding string `json:"encoding"`
}

const defaultSnapshotEntryThreshold = 100

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
	if a.file != nil {
		_ = a.writer.Flush()
		_ = a.file.Sync()

		count, err := countAOFEntries(a.file, defaultSnapshotEntryThreshold)

		if err != nil {
			logger.Error("failed to count AOF entries", "error", err)
		} else if count < defaultSnapshotEntryThreshold {
			logger.Info("Skipping snapshot. AOF entires are below threshold", "entries", count)
			return nil
		}
	}

	// Create a new temp file
	originalPath := a.file.Name()
	tempPath := originalPath + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp AOF: %w", err)
	}
	defer tempFile.Close()

	tempWriter := bufio.NewWriter(tempFile)

	// Version control of aof snapshot
	hdr := AOFHeader{
		Format:   "aof",
		Version:  time.Now().UTC().Format(time.RFC3339Nano),
		Encoding: "json-lines",
	}
	hd, err := json.Marshal(hdr)
	if err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to marshal AOF header: %w", err)
	}
	if _, err := tempWriter.Write(append(hd, '\n')); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to write AOF header: %w", err)
	}

	snapshot := store.GetAll()
	for key, value := range snapshot {
		op := Operation{
			Type:      "SET",
			Key:       key,
			ValueType: string(value.Type()),
			Value:     string(value.Serialize()),
		}
		b, err := json.Marshal(op)
		if err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return fmt.Errorf("failed to marshal snapshot op: %w", err)
		}

		if _, err := tempWriter.Write(append(b, '\n')); err != nil {
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

func countAOFEntries(f *os.File, limit int) (int, error) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	scanner := bufio.NewScanner(f)
	count := 0
	firstLine := true
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if firstLine {
			firstLine = false
			if strings.HasPrefix(line, "{") {
				var hdr AOFHeader
				if err := json.Unmarshal([]byte(line), &hdr); err != nil {
					continue
				}
			}
		}
		count++
		if limit > 0 && count > limit {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return count, err
	}
	return count, nil
}

func (a *AOF) Write(operation, key, valueType, value string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	op := Operation{
		Type:      operation,
		Key:       key,
		ValueType: valueType,
		Value:     value,
	}
	b, err := json.Marshal(op)
	if err != nil {
		logger.Error("failed to marshal AOF operation", "error", err)
		return fmt.Errorf("failed to marshal AOF operation: %w", err)
	}

	logger.Debug("writing to AOF", "operation", operation, "key", key, "valueType", valueType)

	if _, err := a.writer.Write(append(b, '\n')); err != nil {
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
	firstLine := true
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if strings.TrimSpace(line) == "" {
			continue
		}

		if firstLine {
			firstLine = false
			trim := strings.TrimSpace(line)
			if strings.HasPrefix(trim, "{") {
				var hdr AOFHeader
				if err := json.Unmarshal([]byte(trim), &hdr); err == nil && hdr.Format == "aof" {
					logger.Info("AOF header detected", "version", hdr.Version)
					continue
				}
			}
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
	trimmed := strings.TrimSpace(line)
	logger.Debug("parsing operation", "line", line)

	// Try JSON (new format) first
	if strings.HasPrefix(trimmed, "{") {
		var op Operation
		if err := json.Unmarshal([]byte(trimmed), &op); err == nil {
			if op.Type == "" || op.Key == "" {
				return Operation{}, fmt.Errorf("invalid JSON operation: %s", line)
			}
			logger.Debug("operation parsed (json)", "type", op.Type, "key", op.Key, "valueType", op.ValueType, "valueLength", len(op.Value))
			return op, nil
		}
		logger.Debug("failed to unmarshal JSON AOF op, falling back to legacy parser")
	}

	// legacy format: split by spaces into max 4 parts
	parts := strings.SplitN(line, " ", 4)
	logger.Debug("parsing operation (legacy)", "line", line, "partsCount", len(parts))

	if len(parts) < 2 {
		return Operation{}, fmt.Errorf("invalid operation format %s", line)
	}

	var valueType, value string
	if len(parts) >= 3 {
		valueType = parts[2]
	}
	if len(parts) == 4 {
		value = parts[3]
	}

	op := Operation{
		Type:      parts[0],
		Key:       parts[1],
		ValueType: valueType,
		Value:     value,
	}

	if op.Type == "SET" {
		if parts[2] == "" || parts[3] == "" {
			return Operation{}, fmt.Errorf("SET operation missing value type or value: %s", line)
		}
	}

	logger.Debug("operation parsed (legacy)", "type", op.Type, "key", op.Key, "valueType", op.ValueType, "valueLength", len(op.Value))
	return op, nil
}
