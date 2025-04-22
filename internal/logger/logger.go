package logger

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/fatih/color"
)

type Level int

const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel

	timestampFormat = "01/Jan/2006:15:04:05 -0700"

	levelTrace   = "TRACE"
	levelDebug   = "DEBUG"
	levelInfo    = "INFO"
	levelWarn    = "WARN"
	levelError   = "ERROR"
	levelUnknown = "UNKNOWN"
)

type Fields map[string]any

type Logger struct {
	level     Level
	output    io.Writer
	requestID string
}

func New(level Level, output io.Writer) *Logger {
	return &Logger{
		level:  level,
		output: output,
	}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	newLogger := *l
	newLogger.requestID = requestID
	return &newLogger
}

func (l *Logger) log(level Level, msg string, fields Fields) {
	if level < l.level {
		return
	}

	timestamp := time.Now().UTC().Format(timestampFormat)
	levelStr := level.String()

	timestampPart := color.New(color.FgHiBlack).Sprintf("[%s]", timestamp)
	levelPart := l.getLevelColor(level).Sprint(levelStr)

	var requestIDPart string

	if l.requestID != "" {
		requestIDPart = fmt.Sprintf(" [%s]", color.New(color.FgBlue).Sprint(l.requestID))
	}

	messagePart := color.New(color.FgWhite).Sprintf("\"%s\"", msg)

	var fieldsPart string
	keys := make([]string, 0, len(fields))

	for k := range fields {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		fieldsPart += " " + color.New(color.FgCyan).Sprintf("%s=%v", k, fields[k])
	}

	logLine := fmt.Sprintf("%s %s%s %s%s", timestampPart, levelPart, requestIDPart, messagePart, fieldsPart)
	_, _ = fmt.Fprintln(l.output, logLine)
}

func (l *Logger) getLevelColor(level Level) *color.Color {
	switch level {
	case TraceLevel:
		return color.New(color.FgMagenta)
	case DebugLevel:
		return color.New(color.FgCyan)
	case InfoLevel:
		return color.New(color.FgGreen)
	case WarnLevel:
		return color.New(color.FgYellow)
	case ErrorLevel:
		return color.New(color.FgRed)
	default:
		return color.New(color.FgWhite)
	}
}

func (l Level) String() string {
	switch l {
	case TraceLevel:
		return levelTrace
	case DebugLevel:
		return levelDebug
	case InfoLevel:
		return levelInfo
	case WarnLevel:
		return levelWarn
	case ErrorLevel:
		return levelError
	default:
		return levelUnknown
	}
}

func (l *Logger) Trace(msg string, fields Fields) {
	l.log(TraceLevel, msg, fields)
}

func (l *Logger) Debug(msg string, fields Fields) {
	l.log(DebugLevel, msg, fields)
}

func (l *Logger) Info(msg string, fields Fields) {
	l.log(InfoLevel, msg, fields)
}

func (l *Logger) Warn(msg string, fields Fields) {
	l.log(WarnLevel, msg, fields)
}

func (l *Logger) Error(msg string, fields Fields) {
	l.log(ErrorLevel, msg, fields)
}
