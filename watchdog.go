package watchdog

import (
	"sync"
	"time"
)

type Task struct {
	Schedule time.Duration
	Command func(time.Time) error
	Timeout time.Duration
}

type Execution struct {
	Task *Task
	StartedAt time.Time
	FinishedAt time.Time
	Error error
}

type Stall struct {
	Task *Task
	StartedAt time.Time
	StalledAt time.Time
}

type Watchdog struct {
	tasks []*Task

	done chan bool
	sync sync.WaitGroup

	executions chan *Execution
	stalls chan *Stall
}

func Watch(tasks ...*Task) *Watchdog {
	w := &Watchdog{
		tasks: tasks,
		done: make(chan bool),
		executions: make(chan *Execution, 10),
		stalls: make(chan *Stall, 10),
	}
	go w.run()
	return w
}

func (w *Watchdog) Executions() <- chan *Execution {
	return w.executions
}

func (w *Watchdog) Stalls() <- chan *Stall {
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
	stallTimer := time.NewTimer(task.Timeout)
	go func () {
		for startedAt := range schedule {
			stallTimer.Reset(task.Timeout)
			err := task.Command(startedAt)
			finishedAt := time.Now()
			if onTime := stallTimer.Reset(task.Schedule); onTime {
				w.executions <- &Execution{task, startedAt, finishedAt, err}
			}
		}
	}()
	loop: for {
		var startedAt time.Time
		select {
		case <- w.done:
			ticker.Stop()
			stallTimer.Stop()
			break loop
		case startedAt = <- ticker.C:
			select {
			case schedule <- startedAt:
			default:
			}
		case stalledAt := <- stallTimer.C:
			w.stalls <- &Stall{task, startedAt, stalledAt}
		}
	}
	close(schedule)
	stallTimer.Stop()
	w.sync.Done()
}

func (w *Watchdog) Stop() {
	close(w.done)
	go func() {
		w.sync.Wait()
		close(w.executions)
		close(w.stalls)
	}()
}
