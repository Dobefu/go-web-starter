package logger

import (
	"fmt"
	"os"
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
)

type Fields map[string]any

type Logger struct {
	level     Level
	output    *os.File
	requestID string
}

func New(level Level, output *os.File) *Logger {
	return &Logger{
		level:  level,
		output: output,
	}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		level:     l.level,
		output:    l.output,
		requestID: requestID,
	}
}

func (l *Logger) log(level Level, msg string, fields Fields) {
	if l.level != TraceLevel && level < l.level {
		return
	}

	timestamp := time.Now().UTC().Format("01/Jan/2006:15:04:05 -0700")
	levelStr := level.String()

	timestampPart := color.New(color.FgHiBlack).Sprintf("[%s]", timestamp)
	levelPart := l.getLevelColor(level).Sprint(levelStr)

	var requestIDPart string

	if l.requestID != "" {
		requestIDPart = fmt.Sprintf(" [%s]", color.New(color.FgBlue).Sprint(l.requestID))
	}

	messagePart := color.New(color.FgWhite).Sprintf("\"%s\"", msg)

	var fieldsPart string

	for k, v := range fields {
		fieldsPart += " " + color.New(color.FgCyan).Sprintf("%s=%v", k, v)
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
		return "TRACE"
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
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
