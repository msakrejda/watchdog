package watchdog

import (
	"testing"
	"time"
)

func drainExecutions(execs *[]*Execution, execCh <- chan *Execution, done chan <- bool) {
	var tmp []*Execution
	for e := range execCh {
		tmp = append(*execs, e)
		*execs = tmp
	}
	done <- true
}

func drainStalls(stalls *[]*Stall, stallCh <- chan *Stall, done chan <- bool) {
	var tmp []*Stall
	for s := range stallCh {
		tmp = append(*stalls, s)
		*stalls = tmp
	}
	done <- true
}

// true if b happened within d of a
func within(a, b time.Time, delta time.Duration) bool {
	return a.Before(b) && b.Sub(a) < delta
}

func TestRun(t *testing.T) {
	count := 0
	freq := 10 * time.Millisecond
	expectedRuntime := (3 * time.Millisecond) // generous padding here
	task := &Task{freq, func(time.Time) error { count += 1; return nil }, time.Hour}
	start := time.Now()
	w := Watch(task)
	var executions []*Execution
	var stalls []*Stall
	done := make(chan bool)
	go drainExecutions(&executions, w.Executions(), done)
	go drainStalls(&stalls, w.Stalls(), done)
	<- time.After(25 * time.Millisecond)
	w.Stop()
	<- done
	<- done
	if count != 2 {
		t.Errorf("expected task to trigger 2 times; got %d", count)
	}
	if execCount := len(executions); execCount != 2 {
		t.Errorf("expected task to send 2 executions; got %d", execCount)
	}
	for i, exec := range executions {
		if exec.Task != task {
			t.Errorf("expected execution %d task to be %p; got %p", i, task, exec.Task)
		}
		padding := 1 * time.Millisecond
		stepDelay := time.Duration(i + 1) * freq
		if expected := start.Add(stepDelay); !within(expected, exec.StartedAt, padding) {
			t.Errorf("expected execution %d start to be within %v of schedule; got within %v",
				i, padding, exec.StartedAt.Sub(expected))
		}
		if exec.FinishedAt.Before(exec.StartedAt) {
			t.Errorf("expected execution %d end not to shear time; started at %v, finished at %v",
				i, exec.StartedAt, exec.FinishedAt)
		}
		if exec.FinishedAt.Sub(exec.StartedAt) > expectedRuntime {
			t.Errorf("expected execution %d end to run for no more than %v; got %v",
				i, expectedRuntime, exec.FinishedAt.Sub(exec.StartedAt))
		}
		if err := exec.Error; err != nil {
			t.Errorf("expected execution %d success; got err %v", i, err)
		}
	}
	if stallCount := len(stalls); stallCount != 0 {
		t.Errorf("expected task to send 0 stalls; got %d", stallCount)
	}
	<- time.After(20 * time.Millisecond)
	if count > 2 {
		t.Errorf("expected task to stop when watchdog stopped")
	}
}

// - doesn't send duplicate stall notifications
// - sn

// check
//  + simple execution
//  + multiple tasks
//  - stalls
//    - no duplicate stalls ?
//  - recovery
//  - stalls with multiple tasks
//  - recovery with multiple tasks
//  - multiple stalls with multiple tasks
//  - multiple recoveries with multiple tasks

