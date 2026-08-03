package main

import (
	"context"
	"sync"
)

// parallel runs fn for each item, with at most `limit` invocations running
// concurrently, and blocks until every item has been processed. fn must be safe
// to call from multiple goroutines.
//
// Items that haven't started by the time ctx is cancelled are skipped, so a
// cancelled run stops rather than working through the rest of its queue. The
// items already in flight are functions to abandon.
func parallel[T any](ctx context.Context, items []T, limit int, fn func(T)) {
	if limit < 1 {
		limit = 1
	}
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, limit)
	for _, item := range items {
		if ctx.Err() != nil {
			break
		}
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
