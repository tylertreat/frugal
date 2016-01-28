package frugal

import (
	"fmt"
	"time"
)

// FTransportMonitor watches and heals an FTransport.
type FTransportMonitor struct {
	// ClosedCleanly is called when the transport is closed cleanly by a call to Close()
	ClosedCleanly func()

	// ClosedUncleanly is called when the transport is closed for a reason *other* than a call to Close().
	// Returns whether to try reopening the transport and, if so, how long to wait before making the attempt.
	ClosedUncleanly func() (reopen bool, wait time.Duration)

	// ReopenFailed is called when an attempt to reopen the transport fails.
	// Given the number of previous attempts to re-open the transport and the length of the previous wait,
	// Returns whether to attempt to re-open the transport, and how long to wait before making the attempt.
	ReopenFailed func(prevAttempts uint, prevWait time.Duration) (reopen bool, wait time.Duration)

	// ReopenSucceeded is called after the transport has been successfully re-opened.
	ReopenSucceeded func()

	// InitialReopenWait is the initial delay before the first attempt to re-open the transport.
	InitialReopenWait time.Duration

	// MaxReopenWait is the maximum delay between reopen attempts.
	MaxReopenWait time.Duration
}

// NewFTransportMonitor returns a configuration for a transport monitor that logs events,
// and attempts to re-open closed transport with exponential backoff behavior.
func NewFTransportMonitor(maxReopenAttempts uint, initialWait, maxWait time.Duration) *FTransportMonitor {
	return &FTransportMonitor{
		ClosedUncleanly: func() (bool, time.Duration) {
			return maxReopenAttempts > 0, initialWait
		},
		ReopenFailed: func(prevAttempts uint, prevWait time.Duration) (bool, time.Duration) {
			if prevAttempts >= maxReopenAttempts {
				return false, 0
			}

			nextWait := prevWait * 2
			if nextWait > maxWait {
				nextWait = maxWait
			}
			return true, nextWait
		},
		InitialReopenWait: initialWait,
		MaxReopenWait:     maxWait,
	}
}

// Asynchronously starts a monitor with the given configuration, returning a channel to be used
// as a stop signal.
func (m *FTransportMonitor) monitor(transport FTransport, closedCleanly, closedUncleanly <-chan struct{}) {
MonitoringLoop:
	for {
		select {
		case <-closedCleanly:
			m.handleCleanClose()
			break MonitoringLoop
		case <-closedUncleanly:
			if shouldContinue := m.handleUncleanClose(transport); shouldContinue {
				continue MonitoringLoop
			}
		}
		break
	}
}

// Handle a clean close of the transport.
func (m *FTransportMonitor) handleCleanClose() {
	fmt.Println("FTransport Monitor: FTransport was closed cleanly. Terminating...")
	if m.ClosedCleanly != nil {
		m.ClosedCleanly()
	}
}

// Handle an unclean close of the transport.
func (m *FTransportMonitor) handleUncleanClose(transport FTransport) bool {
	fmt.Println("FTransport Monitor: FTransport was closed uncleanly!")

	if m.ClosedUncleanly == nil {
		fmt.Println("FTransport Monitor: ClosedUncleanly callback not defined. Terminating...")
		return false
	}

	var initialWait time.Duration
	var reopen bool
	if reopen, initialWait = m.ClosedUncleanly(); !reopen {
		fmt.Println("FTransport Monitor: ClosedUncleanly callback instructed not to reopen. Terminating...")
		return false
	}

	return m.attemptReopen(initialWait, transport)
}

// Attempt to reopen the uncleanly closed transport.
func (m *FTransportMonitor) attemptReopen(wait time.Duration, transport FTransport) bool {
	reopen := true
	prevAttempts := uint(0)

	for reopen {
		fmt.Printf("FTransport Monitor: Attempting to reopen after %v\n", wait)
		time.Sleep(wait)

		if err := transport.Open(); err != nil {
			fmt.Printf("FTransport Monitor: Failed to re-open transport due to: %v\n", err)
			prevAttempts++
			if m.ReopenFailed == nil {
				fmt.Println("FTransport Monitor: ReopenFailed callback not defined. Terminating...")
				return false
			}

			reopen, wait = m.ReopenFailed(prevAttempts, wait)
			continue
		}
		fmt.Printf("FTransport Monitor: Successfully re-opened!")
		if m.ReopenSucceeded != nil {
			m.ReopenSucceeded()
		}
		return true
	}

	fmt.Println("FTransport Monitor: ReopenFailed callback instructed not to reopen. Terminating...")
	return false
}
