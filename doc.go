/*

Watchdog is an in-process task scheduler and simple execution monitor.
Watchdog accepts a workload and runs it at the specified interval.

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
				fmt.Println("done!")
				break loop
			}
		}
	}

*/
