package logger

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
}

type stdLogger struct {
	out   *log.Logger
	level Level
}

func NewStdLogger(w io.Writer, prefix, level string) Logger {
	return &stdLogger{
		out:   log.New(w, prefix, 0),
		level: parseLevel(level),
	}
}

func (l *stdLogger) Debug(msg string, kv ...any) { l.log(LevelDebug, msg, kv...) }
func (l *stdLogger) Info(msg string, kv ...any)  { l.log(LevelInfo, msg, kv...) }
func (l *stdLogger) Warn(msg string, kv ...any)  { l.log(LevelWarn, msg, kv...) }
func (l *stdLogger) Error(msg string, kv ...any) { l.log(LevelError, msg, kv...) }

func (l *stdLogger) log(level Level, msg string, kv ...any) {
	if level < l.level {
		return
	}

	var b strings.Builder
	b.WriteString(time.Now().Format(time.RFC3339))
	b.WriteString(" ")

	b.WriteString("[")
	b.WriteString(levelString(level))
	b.WriteString("] ")

	b.WriteString(msg)

	if len(kv) > 0 {
		b.WriteString(" | ")
		b.WriteString(formatKV(kv))
	}

	l.out.Println(b.String())
}

var L Logger = NewNopLogger()

func Init(w io.Writer, prefix, level string) {
	L = NewStdLogger(w, prefix, level)
}

func Debug(msg string, kv ...any) {
	L.Debug(msg, kv...)
}

func Info(msg string, kv ...any) {
	L.Info(msg, kv...)
}

func Warn(msg string, kv ...any) {
	L.Warn(msg, kv...)
}

func Error(msg string, kv ...any) {
	L.Error(msg, kv...)
}

type nopLogger struct{}

func NewNopLogger() Logger {
	return &nopLogger{}
}

func (n *nopLogger) Debug(string, ...any) {}
func (n *nopLogger) Info(string, ...any)  {}
func (n *nopLogger) Warn(string, ...any)  {}
func (n *nopLogger) Error(string, ...any) {}

func parseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func levelString(level Level) string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func formatKV(kv []any) string {
	var b strings.Builder

	for i := 0; i < len(kv); i += 2 {
		key := fmt.Sprint(kv[i])
		val := "<missing>"

		if i+1 < len(kv) {
			val = fmt.Sprint(kv[i+1])
		}

		b.WriteString(key)
		b.WriteString(" = ")
		b.WriteString(val)

		if i+2 < len(kv) {
			b.WriteString(", ")
		}
	}

	return b.String()
}
