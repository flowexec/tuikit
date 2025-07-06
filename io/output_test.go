package io_test

import (
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/jahvon/tuikit/io"
	"github.com/jahvon/tuikit/io/mocks"
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
	mockLogger.EXPECT().Infox("line 1", fields...)
	mockLogger.EXPECT().Infox("line 2", fields...)
	mockLogger.EXPECT().Infox("line 3", fields...)
	mockLogger.EXPECT().Infox("line 4", fields...)

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
	mockLogger.EXPECT().Noticex("line 1", fields...)
	mockLogger.EXPECT().Noticex("line 2", fields...)
	mockLogger.EXPECT().Noticex("line 3", fields...)
	mockLogger.EXPECT().Noticex("line 4", fields...)

	_, err := writer.Write(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
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
