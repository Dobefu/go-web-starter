package logger

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func init() {
	color.NoColor = false
}

func createTempLogger(t *testing.T, level Level) (*Logger, func() string) {
	tmpfile, err := os.CreateTemp("", "logger_test")
	assert.NoError(t, err)

	cleanup := func() string {
		err := tmpfile.Sync()
		assert.NoError(t, err)

		content, err := os.ReadFile(tmpfile.Name())
		assert.NoError(t, err)

		err = tmpfile.Close()
		assert.NoError(t, err)

		err = os.Remove(tmpfile.Name())
		assert.NoError(t, err)

		return string(content)
	}

	return New(level, tmpfile), cleanup
}

func TestBasicLogging(t *testing.T) {
	t.Parallel()

	logger, cleanup := createTempLogger(t, DebugLevel)

	logger.Debug("Debug message", Fields{"key": "value"})
	logger.Info("Info message", Fields{"key": "value"})
	logger.Warn("Warning message", Fields{"key": "value"})
	logger.Error("Error message", Fields{"key": "value"})

	loggerWithID := logger.WithRequestID("test-request-id")
	loggerWithID.Info("Message with request ID", Fields{"key": "value"})

	output := cleanup()
	lines := strings.Split(output, "\n")

	for _, line := range lines[:len(lines)-1] {
		assert.Contains(t, line, time.Now().UTC().Format("01/Jan/2006"))
		assert.Contains(t, line, "key=value")
		assert.Contains(t, line, "\"")
	}

	requestIDLine := lines[len(lines)-2]
	assert.Contains(t, requestIDLine, color.New(color.FgBlue).Sprint("test-request-id"))
}

func TestLogLevelFiltering(t *testing.T) {
	t.Parallel()

	logger, cleanup := createTempLogger(t, DebugLevel)

	logger.Trace("Should not appear", nil)
	logger.Debug("Should appear", nil)
	logger.Info("Should appear", nil)
	logger.Warn("Should appear", nil)
	logger.Error("Should appear", nil)

	output := cleanup()
	assert.NotContains(t, output, "TRACE")
	assert.Contains(t, output, "DEBUG")
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "WARN")
	assert.Contains(t, output, "ERROR")
}

func TestColorOutput(t *testing.T) {
	t.Parallel()

	pr, pw, err := os.Pipe()
	assert.NoError(t, err)
	defer func() { _ = pr.Close() }()

	logger := New(TraceLevel, pw)
	logger.Trace("Test message", nil)
	err = pw.Close()
	assert.NoError(t, err)

	var output strings.Builder
	buf := make([]byte, 1024)

	for {
		n, err := pr.Read(buf)
		if err != nil {
			break
		}
		output.Write(buf[:n])
	}

	content := output.String()

	assert.Contains(t, content, "\033[")
	assert.Contains(t, content, "\033[90m")
	assert.Contains(t, content, "\033[35m")
	assert.Contains(t, content, "\033[37m")
}

func TestUnknownLevel(t *testing.T) {
	t.Parallel()

	unknownLevel := Level(9001)
	assert.Equal(t, "UNKNOWN", unknownLevel.String())

	logger, cleanup := createTempLogger(t, DebugLevel)
	logger.log(unknownLevel, "Test message", nil)
	output := cleanup()

	assert.Contains(t, output, color.New(color.FgWhite).Sprint("UNKNOWN"))
}

func TestLoggerWithNilOutput_DoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		logger := New(DebugLevel, nil)
		logger.Info("This should not panic", Fields{"foo": "bar"})
	})
}

func TestLoggerLogWithNilOutputField_DoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		logger := &Logger{level: DebugLevel, output: nil}
		logger.Info("This should not panic", Fields{"should": "not appear"})
	})
}
