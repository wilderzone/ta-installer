package main

import "sync"

// parallel runs fn for each item, with at most `limit` invocations running
// concurrently, and blocks until every item has been processed. fn must be safe
// to call from multiple goroutines.
func parallel[T any](items []T, limit int, fn func(T)) {
	if limit < 1 {
		limit = 1
	}
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, limit)
	for _, item := range items {
		semaphore <- struct{}{} // Blocks once `limit` are in flight.
		wg.Add(1)
		go func(item T) {
			defer wg.Done()
			defer func() { <-semaphore }()
			fn(item)
		}(item)
	}
	wg.Wait()
}
