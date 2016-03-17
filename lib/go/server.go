package frugal

import "time"

// FServer is a Frugal service server.
type FServer interface {
	// Serve starts the server.
	Serve() error

	// Stop the server. This is optional on a per-implementation basis. Not all
	// servers are required to be cleanly stoppable.
	Stop() error

	// SetLoggingWatermark sets the miniumum amount of time a frame may await
	// processing before triggering a warning log.
	SetLoggingWatermark(watermark time.Duration)
}
