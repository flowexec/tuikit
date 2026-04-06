package io_test

import (
	"testing"
	"time"

	"github.com/flowexec/tuikit/io"
	"github.com/flowexec/tuikit/themes"
)

func TestTaskStatus_String(t *testing.T) {
	tests := []struct {
		status io.TaskStatus
		want   string
	}{
		{io.TaskRunning, "running"},
		{io.TaskSuccess, "success"},
		{io.TaskFailed, "failed"},
		{io.TaskSkipped, "skipped"},
	}
	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("TaskStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestTaskStatus_Icon(t *testing.T) {
	tests := []struct {
		status io.TaskStatus
		want   string
	}{
		{io.TaskRunning, "..."},
		{io.TaskSuccess, "ok"},
		{io.TaskFailed, "FAIL"},
		{io.TaskSkipped, "skip"},
	}
	for _, tt := range tests {
		if got := tt.status.Icon(); got != tt.want {
			t.Errorf("TaskStatus(%d).Icon() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestTaskTracker_StartAndComplete(t *testing.T) {
	tracker := io.NewTaskTracker()

	task1 := tracker.StartTask("build")
	task2 := tracker.StartTask("test")

	if task1.Name != "build" {
		t.Errorf("task1.Name = %q, want %q", task1.Name, "build")
	}
	if task2.Name != "test" {
		t.Errorf("task2.Name = %q, want %q", task2.Name, "test")
	}
	if task1.ColorIdx != 0 {
		t.Errorf("task1.ColorIdx = %d, want 0", task1.ColorIdx)
	}
	if task2.ColorIdx != 1 {
		t.Errorf("task2.ColorIdx = %d, want 1", task2.ColorIdx)
	}
	if task1.Status != io.TaskRunning {
		t.Errorf("task1.Status = %v, want TaskRunning", task1.Status)
	}

	tracker.CompleteTask(task1, io.TaskSuccess, nil)
	if task1.Status != io.TaskSuccess {
		t.Errorf("after complete, task1.Status = %v, want TaskSuccess", task1.Status)
	}
	if task1.EndTime.IsZero() {
		t.Error("after complete, task1.EndTime should not be zero")
	}

	tasks := tracker.Tasks()
	if len(tasks) != 2 {
		t.Fatalf("Tasks() returned %d tasks, want 2", len(tasks))
	}
}

func TestTaskContext_Duration(t *testing.T) {
	now := time.Now()
	tc := &io.TaskContext{
		Name:      "test",
		StartTime: now.Add(-500 * time.Millisecond),
		EndTime:   now,
	}
	d := tc.Duration()
	if d < 400*time.Millisecond || d > 600*time.Millisecond {
		t.Errorf("Duration() = %v, want ~500ms", d)
	}
}

func TestTaskContext_DurationRunning(t *testing.T) {
	tc := &io.TaskContext{
		Name:      "test",
		StartTime: time.Now().Add(-100 * time.Millisecond),
	}
	d := tc.Duration()
	if d < 50*time.Millisecond {
		t.Errorf("Duration() for running task = %v, want >= 50ms", d)
	}
}

func TestPrefixColors(t *testing.T) {
	palette := themes.EverforestTheme().ColorPalette()
	colors := io.PrefixColors(palette)
	if len(colors) != 6 {
		t.Fatalf("PrefixColors() returned %d colors, want 6", len(colors))
	}
	for i, c := range colors {
		if c == "" {
			t.Errorf("PrefixColors()[%d] is empty", i)
		}
	}
}
