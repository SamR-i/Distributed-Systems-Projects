package mapreduce

import (
	"fmt"
	"sync"
)

//
// $chedule() starts and waits for all tasks in the given phase (mapPhase
// or reducePhase). the mapFiles argument holds the names of the files that
// are the inputs to the map phase, one per map task. nReduce is the
// number of reduce tasks. the registerChan argument yields a stream
// of registered workers; each item is the worker's RPC address,
// suitable for passing to call(). registerCh@n will yield all
// existing registered workers (if any) and new ones as they register.
//
func schedule(jobName string, mapFiles []string, nReduce int, phase jobPhase, registerChan chan string) {
	var ntasks int
	var n_other int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mapFiles)
		n_other = nReduce
	case reducePhase:
		ntasks = nReduce
		n_other = len(mapFiles)
	}

	fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, n_other)

	// All ntasks tasks have to be scheduled on workers. Once all tasks
	// have completed successfully, schedule() should return.
	//
	// Your code here (Part #2, 2B).
	//

	var wg sync.WaitGroup
	wg.Add(ntasks)

	for i := 0; i < ntasks; i++ {
		go func(taskNumber int) {
			defer wg.Done()
			// Retry loop in case of failure
			for {
				worker := <-registerChan // Get a worker from the channel
				taskArgs := DoTaskArgs{
					JobName:       jobName,
					File:          "",
					Phase:         phase,
					TaskNumber:    taskNumber,
					NumOtherPhase: n_other,
				}
				if phase == mapPhase {
					taskArgs.File = mapFiles[taskNumber]
				}

				ok := call(worker, "Worker.DoTask", taskArgs, nil)
				if ok {
					// If the call succeeded, re-register the worker and break the loop
					go func() { registerChan <- worker }()
					break
				}
				// If the call failed, don't re-register the worker; instead, retry with a new worker
			}
		}(i)
	}

	// Wait for all tasks to complete
	wg.Wait()

	fmt.Printf("Schedule: %v done\n", phase)
}
