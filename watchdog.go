package watchdog

import (
	"sync"
	"time"
)

// Basic scheduling unit
type Task struct {
	// How frequently the task should execute
	Schedule time.Duration
	// Function to invoke: each execution will be passed the time
	// it was originally scheduled for (which may be behind
	// wall-clock time in case of stalls).
	Command func(time.Time) error
	// How long to wait before considering an execution stalled
	Timeout time.Duration
}

// Information about each execution
type Execution struct {
	// Task being executed
	Task *Task
	// Time the Task was originally scheduled for
	StartedAt time.Time
	// Time the Task completed
	FinishedAt time.Time
	// Error returned by the Task Command
	Error error
}

// Information about each stall
type Stall struct {
	// Task which stalled
	Task *Task
	// Time the Task was originally scheduled for
	StartedAt time.Time
	// Time the task was considered stalled
	StalledAt time.Time
}

// Execution monitor
type Watchdog struct {
	tasks []*Task

	done chan bool
	sync sync.WaitGroup

	executions chan *Execution
	stalls     chan *Stall
}

// Create a new, running watchdog with the given task(s).
func Watch(tasks ...*Task) *Watchdog {
	w := &Watchdog{
		tasks:      tasks,
		done:       make(chan bool),
		executions: make(chan *Execution, 10),
		stalls:     make(chan *Stall, 10),
	}
	go w.run()
	return w
}

// Channel of executions for a given Watchdog. Note that the channel
// must be drained promptly while the Watchdog is running: the channel
// has a small buffer, but failure to keep up will apply backpressure
// to the system and delay scheduling.
func (w *Watchdog) Executions() <-chan *Execution {
	return w.executions
}

// Channel of stalls for a given Watchdog. As above, the channel must
// be drained while the Watchdog is running.
func (w *Watchdog) Stalls() <-chan *Stall {
	return w.stalls
}

func (w *Watchdog) run() {
	for _, task := range w.tasks {
		go w.runTask(task)
	}
	w.sync.Add(len(w.tasks))
}

func (w *Watchdog) runTask(task *Task) {
	ticker := time.NewTicker(task.Schedule)
	schedule := make(chan time.Time, 1)
	stallTimer := new(time.Timer)
	taskDone := make(chan bool, 1)
	go func() {
		for startedAt := range schedule {
			stallTimer = time.NewTimer(task.Timeout)
			err := task.Command(startedAt)
			stallTimer.Reset(task.Schedule)
			finishedAt := time.Now()
			w.executions <- &Execution{task, startedAt, finishedAt, err}
		}
		taskDone <- true
	}()
	var startedAt time.Time
monitor:
	for {
		select {
		case <-w.done:
			ticker.Stop()
			close(schedule)
			break monitor
		case stalledAt := <-stallTimer.C:
			w.stalls <- &Stall{task, startedAt, stalledAt}
		case startedAt = <-ticker.C:
			select {
			case schedule <- startedAt:
			default:
			}
		}
	}
cleanup:
	for {
		select {
		case stalledAt := <-stallTimer.C:
			w.stalls <- &Stall{task, startedAt, stalledAt}
		case <-taskDone:
			stallTimer.Stop()
			break cleanup
		}
	}
	w.sync.Done()
}

// Stop a running Watchdog. Waits for any currently-executing tasks to
// complete, then closes the Executions and Stalls channels and
// returns.
func (w *Watchdog) Stop() {
	close(w.done)
	w.sync.Wait()
	close(w.executions)
	close(w.stalls)
}
