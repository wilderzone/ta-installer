package main

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// PROGRESS_INTERVAL is how often a measured download reports its progress. Often
// enough for an interface to animate smoothly, rarely enough that the reporting never
// becomes the thing being measured.
const PROGRESS_INTERVAL = 250 * time.Millisecond

// RATE_SMOOTHING is the weight given to the newest sample when averaging the transfer
// rate. Low enough that a momentary stall doesn't wipe out the estimate, high enough that
// the rate still follows a connection that genuinely changes speed.
const RATE_SMOOTHING = 0.15

// Progress tracks the bytes received across a batch of concurrent downloads, and reports
// the running total at a fixed cadence. A nil *Progress is a valid, inert progress: every
// method tolerates it, so a caller with nothing to measure never needs to check.
type Progress struct {
	counter  *Counter
	size     uint64        // Total bytes expected.
	received atomic.Uint64 // Bytes received so far.
	rate     atomic.Uint64 // Smoothed bytes per second.
	sampled  uint64        // The reading taken at the previous tick.
	once     sync.Once
	quit     chan struct{}
	finished chan struct{}
}

// measure attaches a byte counter to the counter's work, reporting the progress of a
// `size`-byte download until the counter is done. It returns nil - an inert progress -
// when the size isn't known, or when nothing is listening for the reports.
func (c *Counter) measure(size uint64) *Progress {
	// Besides `-p`, the spinner is the only consumer, and it is already inert in the
	// modes where redrawing would corrupt the output - so a counter without one has
	// nobody watching it closely enough to be worth measuring.
	if size == 0 || c.logger.silent || (!c.logger.progress && c.spinner == nil) {
		return nil
	}
	progress := &Progress{
		counter:  c,
		size:     size,
		quit:     make(chan struct{}),
		finished: make(chan struct{}),
	}
	c.progress = progress
	progress.report(DOWNLOAD_STARTED)
	go progress.run()
	return progress
}

// run reports the download's progress until it is stopped.
func (p *Progress) run() {
	defer close(p.finished)
	ticker := time.NewTicker(PROGRESS_INTERVAL)
	defer ticker.Stop()
	for {
		select {
			case <-p.quit:
				return

			case <-ticker.C:
				p.sample()
				p.report(DOWNLOAD_PROGRESS)
		}
	}
}

// sample folds the bytes received since the previous tick into the smoothed transfer rate.
func (p *Progress) sample() {
	received := p.received.Load()
	instant := float64(received-p.sampled) / PROGRESS_INTERVAL.Seconds()
	p.sampled = received
	rate := float64(p.rate.Load())
	if rate == 0 {
		rate = instant
	} else {
		rate = rate*(1-RATE_SMOOTHING) + instant*RATE_SMOOTHING
	}
	p.rate.Store(uint64(rate))
}

// report sends the download's current state to whoever is watching: the interface that
// asked for progress lines, or the spinner in front of a person.
func (p *Progress) report(template LogLine) {
	if p.counter.logger.progress {
		line := template
		line.Bytes = p.received.Load()
		line.Size = p.size
		line.Rate = p.rate.Load()
		p.counter.logger.logProgress(line)
		return
	}
	p.counter.spinner.update(p.counter.status())
}

// stop ends the reporting and sends a final line, so a watching interface knows the
// measured download is over and can go back to counting the work its own way. This
// is safe to call more than once.
func (p *Progress) stop() {
	if p == nil {
		return
	}
	p.once.Do(func() {
		close(p.quit)
		<-p.finished // Wait for the last report, so nothing lands after the final line.
		p.rate.Store(0)
		p.report(DOWNLOAD_COMPLETE)
	})
}

// summary describes the download's progress in plain English.
func (p *Progress) summary() string {
	if p == nil {
		return ""
	}
	received := min(p.received.Load(), p.size)
	summary := fmt.Sprintf("%d%% of %s", received*100/p.size, byteCountSI(p.size))
	if remaining := p.remaining(); remaining != "" {
		summary += " - " + remaining
	}
	return summary
}

// remaining estimates how much longer the download will take. The estimate is worth
// nothing until the rate has had a moment to settle, and nothing at all once every byte
// has arrived, so it is left out in both cases.
func (p *Progress) remaining() string {
	rate := p.rate.Load()
	received := p.received.Load()
	if rate == 0 || received >= p.size {
		return ""
	}
	return formatDuration((p.size-received)/rate) + " left"
}

// formatDuration renders a number of seconds as a rough estimate in plain English.
// A second-accurate countdown on a large download would be false precision,
// and would spend the whole download flickering.
func formatDuration(seconds uint64) string {
	if seconds < 10 {
		return "a few seconds"
	}
	if rounded := (seconds + 2) / 5 * 5; seconds < 60 && rounded < 60 {
		return fmt.Sprintf("%d seconds", rounded)
	}
	// Rounding can carry a duration up into the next unit, so each threshold is checked
	// against the rounded figure rather than the raw one - "60 minutes" is nobody's idea
	// of an estimate.
	minutes := (seconds + 30) / 60
	if minutes < 2 {
		return "1 minute"
	}
	if minutes < 60 {
		return fmt.Sprintf("%d minutes", minutes)
	}
	hours := float64(seconds) / (60 * 60)
	if hours == 1 {
		return fmt.Sprintf("%.1f hour", hours)
	}
	if hours < 10 {
		return fmt.Sprintf("%.1f hours", hours)
	}
	return fmt.Sprintf("%.0f hours", hours)
}

// reader wraps r so that everything read through it is counted towards the progress.
func (p *Progress) reader(r io.Reader) io.Reader {
	if p == nil {
		return r
	}
	return &meter{reader: r, progress: p}
}

// meter counts what is read through it, so a download can report itself as it arrives
// rather than only once it has finished.
type meter struct {
	reader   io.Reader
	progress *Progress
}

func (m *meter) Read(p []byte) (int, error) {
	n, err := m.reader.Read(p)
	m.progress.received.Add(uint64(n))
	return n, err
}
