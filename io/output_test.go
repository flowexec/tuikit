package io_test

import (
	"os"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/flowexec/tuikit/io"
	"github.com/flowexec/tuikit/io/mocks"
)

func TestStdOutWriter_WriteText(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(ctrl)
	writer := io.StdOutWriter{
		Logger: mockLogger,
	}

	input := []byte("line 1\nline 2\nline 3\n")
	mockLogger.EXPECT().LogMode().Return(io.Text).AnyTimes()
	mockLogger.EXPECT().Print("line 1\nline 2\nline 3\n")

	_, err := writer.Write(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStdOutWriter_WriteLogFmt(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(ctrl)
	fields := []interface{}{"key1", "value1", "key2", "value2"}
	writer := io.StdOutWriter{
		Logger:    mockLogger,
		LogFields: fields,
	}

	input := []byte("line 1\nline 2\nline 3\nline 4")
	mockLogger.EXPECT().LogMode().Return(io.Logfmt).AnyTimes()
	mockLogger.EXPECT().Info("line 1", fields...)
	mockLogger.EXPECT().Info("line 2", fields...)
	mockLogger.EXPECT().Info("line 3", fields...)
	mockLogger.EXPECT().Info("line 4", fields...)

	_, err := writer.Write(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStdOutWriter_WriteHidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(ctrl)
	writer := io.StdOutWriter{
		Logger: mockLogger,
	}

	input := []byte("line 1\nline 2\nline 3\n")
	mockLogger.EXPECT().LogMode().Return(io.Hidden).AnyTimes()

	_, err := writer.Write(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStdErrWriter_WriteText(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(ctrl)
	writer := io.StdErrWriter{
		Logger: mockLogger,
	}

	input := []byte("line 1\nline 2\nline 3\n")
	mockLogger.EXPECT().LogMode().Return(io.Text).AnyTimes()
	mockLogger.EXPECT().Print("line 1\nline 2\nline 3\n")

	_, err := writer.Write(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStdErrWriter_WriteLogFmt(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(ctrl)
	fields := []interface{}{"key1", "value1", "key2", "value2"}
	writer := io.StdErrWriter{
		Logger:    mockLogger,
		LogFields: fields,
	}

	input := []byte("line 1\nline 2\nline 3\nline 4")
	mockLogger.EXPECT().LogMode().Return(io.Logfmt).AnyTimes()
	mockLogger.EXPECT().Notice("line 1", fields...)
	mockLogger.EXPECT().Notice("line 2", fields...)
	mockLogger.EXPECT().Notice("line 3", fields...)
	mockLogger.EXPECT().Notice("line 4", fields...)

	_, err := writer.Write(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStdOutWriter_WriteLogFmtWithTaskFallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(ctrl)
	fields := []interface{}{"key1", "value1"}
	task := &io.TaskContext{Name: "build"}
	writer := io.StdOutWriter{
		Logger:    mockLogger,
		LogFields: fields,
		Task:      task,
	}

	input := []byte("hello\n")
	mockLogger.EXPECT().LogMode().Return(io.Logfmt).AnyTimes()
	// MockLogger does not implement TaskAwareLogger, so it falls back to Info
	mockLogger.EXPECT().Info("hello", fields...)

	_, err := writer.Write(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStdErrWriter_WriteLogFmtWithTaskFallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(ctrl)
	fields := []interface{}{"key1", "value1"}
	task := &io.TaskContext{Name: "build"}
	writer := io.StdErrWriter{
		Logger:    mockLogger,
		LogFields: fields,
		Task:      task,
	}

	input := []byte("error line\n")
	mockLogger.EXPECT().LogMode().Return(io.Logfmt).AnyTimes()
	// MockLogger does not implement TaskAwareLogger, so it falls back to Notice
	mockLogger.EXPECT().Notice("error line", fields...)

	_, err := writer.Write(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestStdOutWriter_WriteLogFmtWithTaskAwareLogger(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "logger-test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	logger := io.NewLogger(
		io.WithMode(io.Logfmt),
		io.WithOutput(tmpFile),
	)
	task := &io.TaskContext{Name: "build", ColorIdx: 0}
	writer := io.StdOutWriter{
		Logger: logger,
		Task:   task,
	}

	input := []byte("compiling...\n")
	_, err = writer.Write(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Read back output and verify it contains the task prefix
	_, _ = tmpFile.Seek(0, 0)
	buf := make([]byte, 1024)
	n, _ := tmpFile.Read(buf)
	output := string(buf[:n])
	if !strings.Contains(output, "[build]") {
		t.Errorf("Expected output to contain [build] prefix, got: %q", output)
	}
	if !strings.Contains(output, "compiling...") {
		t.Errorf("Expected output to contain message, got: %q", output)
	}
}

func TestStdErrWriter_WriteHidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(ctrl)
	writer := io.StdErrWriter{
		Logger: mockLogger,
	}

	input := []byte("line 1\nline 2\nline 3\n")
	mockLogger.EXPECT().LogMode().Return(io.Hidden).AnyTimes()

	_, err := writer.Write(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
