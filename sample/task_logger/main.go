package main

import (
	"fmt"
	"os"
	"time"

	"github.com/flowexec/tuikit/io"
	"github.com/flowexec/tuikit/themes"
)

func main() {
	logger := io.NewLogger(
		io.WithMode(io.Logfmt),
		io.WithOutput(os.Stdout),
		io.WithTheme(themes.EverforestTheme()),
	)

	tracker := io.NewTaskTracker()

	// Simulate parallel execution of 3 tasks
	fmt.Println("=== Parallel Execution (Logfmt mode with tasks) ===")
	fmt.Println()

	taskGen := tracker.StartTask("generate tuikit/")
	taskLint := tracker.StartTask("lint tuikit/")
	taskTest := tracker.StartTask("test tuikit/tui")

	// Simulate interleaved output via StdOutWriter
	writeTask(logger, taskGen, "Generating go CLI code...")
	writeTask(logger, taskLint, "Running linter...")
	writeTask(logger, taskGen, "All go code generated successfully")
	writeTask(logger, taskLint, "0 issues.")
	writeTask(logger, taskTest, "Running Go unit tests...")
	writeTask(logger, taskTest, "ok      github.com/flowexec/tuikit    (cached)")
	writeTask(logger, taskTest, "ok      github.com/flowexec/tuikit/io    (cached)")
	writeTask(logger, taskTest, "?       github.com/flowexec/tuikit/themes    [no test files]")
	writeTask(logger, taskTest, "Unit tests completed")

	// Simulate stderr output
	writeTaskErr(logger, taskLint, "warning: unused variable in sample.go")

	tracker.CompleteTask(taskGen, io.TaskSuccess, nil)
	tracker.CompleteTask(taskLint, io.TaskSuccess, nil)
	tracker.CompleteTask(taskTest, io.TaskSuccess, nil)

	logger.PrintTaskSummary(tracker.Tasks())

	// Show what solo execution looks like (no task context)
	fmt.Println("=== Solo Execution (Logfmt mode, no task) ===")
	fmt.Println()

	soloWriter := io.StdOutWriter{Logger: logger}
	_, _ = soloWriter.Write([]byte("Building project...\n"))
	_, _ = soloWriter.Write([]byte("Build complete.\n"))

	fmt.Println()

	// Show BeginGroup/EndGroup
	fmt.Println("=== Grouped Output ===")
	logger.BeginGroup("build phase")
	writeTask(logger, taskGen, "Generating code...")
	writeTask(logger, taskGen, "Done.")
	logger.EndGroup()

	fmt.Println()

	// Show a failed task in summary
	fmt.Println("=== Summary with Failures ===")
	fmt.Println()

	tracker2 := io.NewTaskTracker()
	t1 := tracker2.StartTask("build")
	t2 := tracker2.StartTask("test")
	t3 := tracker2.StartTask("deploy")

	tracker2.CompleteTask(t1, io.TaskSuccess, nil)
	tracker2.CompleteTask(t2, io.TaskFailed, fmt.Errorf("2 tests failed"))
	tracker2.CompleteTask(t3, io.TaskSkipped, nil)

	logger.PrintTaskSummary(tracker2.Tasks())
}

func writeTask(logger *io.StandardLogger, task *io.TaskContext, msg string) {
	writer := io.StdOutWriter{
		Logger: logger,
		Task:   task,
	}
	_, _ = writer.Write([]byte(msg + "\n"))
	time.Sleep(20 * time.Millisecond) // small delay to simulate real work
}

func writeTaskErr(logger *io.StandardLogger, task *io.TaskContext, msg string) {
	writer := io.StdErrWriter{
		Logger: logger,
		Task:   task,
	}
	_, _ = writer.Write([]byte(msg + "\n"))
}
