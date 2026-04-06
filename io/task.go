package io

import (
	"fmt"
	"sync"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/flowexec/tuikit/themes"
)

// TaskStatus represents the outcome of a task execution.
type TaskStatus int

const (
	TaskRunning TaskStatus = iota
	TaskSuccess
	TaskFailed
	TaskSkipped
)

func (s TaskStatus) String() string {
	switch s {
	case TaskRunning:
		return "running"
	case TaskSuccess:
		return "success"
	case TaskFailed:
		return "failed"
	case TaskSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// Icon returns a short status indicator for display in summaries.
func (s TaskStatus) Icon() string {
	switch s {
	case TaskRunning:
		return "..."
	case TaskSuccess:
		return "ok"
	case TaskFailed:
		return "FAIL"
	case TaskSkipped:
		return "skip"
	default:
		return "?"
	}
}

// TaskContext holds metadata for a single task/step in a parallel or serial execution.
type TaskContext struct {
	Name      string
	ColorIdx  int
	StartTime time.Time
	EndTime   time.Time
	Status    TaskStatus
	Error     error
}

// Duration returns the elapsed time for the task.
func (tc *TaskContext) Duration() time.Duration {
	if tc.EndTime.IsZero() {
		return time.Since(tc.StartTime)
	}
	return tc.EndTime.Sub(tc.StartTime)
}

// TaskTracker manages task contexts for a group of parallel/serial tasks.
// It is goroutine-safe.
type TaskTracker struct {
	mu    sync.Mutex
	tasks []*TaskContext
	idx   int
}

// NewTaskTracker creates a new TaskTracker.
func NewTaskTracker() *TaskTracker {
	return &TaskTracker{}
}

// StartTask registers a new task and returns its context.
func (tt *TaskTracker) StartTask(name string) *TaskContext {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	tc := &TaskContext{
		Name:      name,
		ColorIdx:  tt.idx,
		StartTime: time.Now(),
		Status:    TaskRunning,
	}
	tt.idx++
	tt.tasks = append(tt.tasks, tc)
	return tc
}

// CompleteTask marks a task as finished with the given status.
func (tt *TaskTracker) CompleteTask(tc *TaskContext, status TaskStatus, err error) {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	tc.EndTime = time.Now()
	tc.Status = status
	tc.Error = err
}

// Tasks returns a snapshot of all tracked tasks.
func (tt *TaskTracker) Tasks() []*TaskContext {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	out := make([]*TaskContext, len(tt.tasks))
	copy(out, tt.tasks)
	return out
}

// TaskAwareLogger is implemented by loggers that support task-prefixed output.
// This is a separate interface from Logger to avoid breaking existing implementations.
type TaskAwareLogger interface {
	PrintWithTask(task *TaskContext, line string)
	PrintErrWithTask(task *TaskContext, line string)
	PrintTaskSummary(tasks []*TaskContext)
	BeginGroup(name string)
	EndGroup()
}

// PrefixColors returns a set of colors suitable for task prefixes,
// derived from the theme's color palette.
func PrefixColors(palette *themes.ColorPalette) []string {
	return []string{
		palette.Primary,
		palette.Secondary,
		palette.Tertiary,
		palette.Info,
		palette.Emphasis,
		palette.Warning,
	}
}

// taskPrefix renders a colored [taskName] prefix for terminal output.
func taskPrefix(task *TaskContext, palette *themes.ColorPalette) string {
	colors := PrefixColors(palette)
	color := colors[task.ColorIdx%len(colors)]
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true)
	return style.Render(fmt.Sprintf("[%s]", task.Name))
}
