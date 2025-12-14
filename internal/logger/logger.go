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
	Debug Level = iota
	Info
	Warn
	Error
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

func (l *stdLogger) Debug(msg string, kv ...any) { l.log(Debug, msg, kv...) }
func (l *stdLogger) Info(msg string, kv ...any)  { l.log(Info, msg, kv...) }
func (l *stdLogger) Warn(msg string, kv ...any)  { l.log(Warn, msg, kv...) }
func (l *stdLogger) Error(msg string, kv ...any) { l.log(Error, msg, kv...) }

func (l *stdLogger) log(level Level, msg string, kv ...any) {
	if level < l.level {
		return
	}

	var b strings.Builder
	b.WriteString(time.Now().Format(time.RFC3339))
	b.WriteString(" ")

	b.WriteString("[")
	b.WriteString(levelString(level))
	b.WriteString("]")

	b.WriteString(msg)

	if len(kv) > 0 {
		b.WriteString(" | ")
		b.WriteString(formatKV(kv))
	}

	l.out.Println(b.String())
}

func parseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return Debug
	case "warn", "warning":
		return Warn
	case "error":
		return Error
	default:
		return Info
	}
}

func levelString(level Level) string {
	switch level {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
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
		b.WriteString("=")
		b.WriteString(val)

		if i+2 < len(kv) {
			b.WriteString(", ")
		}
	}

	return b.String()
}
