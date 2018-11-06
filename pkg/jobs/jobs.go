package jobs

import (
	"sync"
)

var wg sync.WaitGroup

// Async starts a new async function.
func Async(f func()) {
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()

		f()
	}()
}

// Shutdown waits for all jobs to finish.
func Shutdown() {
	wg.Wait()
}
