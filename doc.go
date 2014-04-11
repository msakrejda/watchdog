/*

Watchdog is an in-process task scheduler and simple execution monitor.
Watchdog accepts a workload and runs it at the specified interval.

Watchdog exposes two channels for monitoring its workload: Executions
and Stalls. Executions are sent for every invocation of a workload
Task, and stalls are only sent if a Task invocation takes longer than
a specified timeout.

A Watchdog is created with the Watch method, and starts running its
workload immediately. Its execution semantics are very close to those
of time.Ticker: a single tick may be "queued up" at any time if the
command takes longer to execute than the scheduling period.

A Watchdog may be stopped with the Stop command. If a task is
currently executing, that task will complete before Stop returns, and
information about its execution and stall (if any) will be sent on the
standard channels.

Here is a simple but functioning example:

	import (
		"fmt"
		"github.com/deafbybeheading/watchdog"
		"time"
	)

	func main() {
		w := watchdog.Watch(&Task{
			Schedule: 1 * time.Second,
			Command: func(t time.Time) error {
				return fmt.Printf("the time is %v\n", t)
			},
			Timeout: 10 * time.Milliseconds,
		})
		loop: for {
			executions := 0
			stalls := 0
			select {
			case exec := <- w.Executions():
				executions += 1
				fmt.Printf("invoked at %v; ran %v times\n",
					exec.StartedAt, executions)
				if err := exec.Error; err != nil {
					fmt.Printf("encountered error: %v\n", err)
				}
			case stall := <- w.Stalls():
				fmt.Printf("execution %v stalled at %v\n",
					executions + 1, stall.StalledAt)
			case <- time.After(10 * time.Second):
				w.Stop()
				fmt.Println("done!")
				break loop
			}
		}
	}

*/
