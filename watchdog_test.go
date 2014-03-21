package watchdog

import (
	"testing"
	"time"
)

func drainExecutions() {
	
}

func TestRun(t *testing.T) {
	count := 0
	w := Watch(
		&Task{10 * time.Millisecond,
			func(time.Time) { count += 1 },
			time.Hour})
	<- time.After(25 * time.Millisecond)
	w.Stop()
	if count != 2 {
		t.Errorf("expected task to trigger 2 times; got %d", count)
	}
	<- time.After(20 * time.Millisecond)
	if count > 2 {
		t.Errorf("expected task to stop when watchdog stopped")
	}
}

func TestRunMultiple(t *testing.T) {
	aCount, bCount := 0, 0
	w := Watch(
		&Task{6 * time.Millisecond,
			func(time.Time) { aCount += 1 },
			time.Hour},
		&Task{7 * time.Millisecond,
			func(time.Time) { bCount += 1 },
			time.Hour})
	<- time.After(20 * time.Millisecond)
	w.Stop()
	if aCount != 3 {
		t.Errorf("expected task A to trigger 3 times; got %d", aCount)
	}
	if bCount != 2 {
		t.Errorf("expected task B to trigger 2 times; got %d", bCount)
	}
	<- time.After(20 * time.Millisecond)
	if aCount > 3 {
		t.Errorf("expected task A to stop when watchdog stopped")
	}
	if bCount > 2 {
		t.Errorf("expected task B to stop when watchdog stopped")
	}
}

func TestStall(t *testing.T) {
	count := 0
	w := Watch(
		&Task{10 * time.Millisecond,
			func(time.Time) { <- time.After(2 * time.Millisecond); count += 1 },
			1 * time.Millisecond})
	stalls := make([]*TaskEvent, 0, 2)
	go func() {
		events := w.Events()
		for e := range events {
			stalls = append(stalls, e)
		}
	}()
	<- time.After(25 * time.Millisecond)
	w.Stop()
	if count != 2 {
		t.Errorf("expected stalled task to trigger 2 times; got %d", count)
	}
	if len(stalls) != 2 {
		t.Errorf("expected 2 stalls; got %d", len(stalls))
	}
	<- time.After(20 * time.Millisecond)
	if count > 2 {
		t.Errorf("expected stalled task to stop when watchdog stopped")
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

