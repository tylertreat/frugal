// Package frugal provides the library APIs used by the Frugal code generator.
package frugal

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

var (
	packageLogger logrus.Logger = *logrus.StandardLogger()
	loggerMu      sync.RWMutex
)

// SetLogger sets the Logger used by Frugal.
func SetLogger(logger *logrus.Logger) {
	loggerMu.Lock()
	packageLogger = *logger
	loggerMu.Unlock()
}

// logger returns the global Logger. Do not mutate the pointer returned, use
// SetLogger instead.
func logger() *logrus.Logger {
	loggerMu.RLock()
	logger := &packageLogger
	loggerMu.RUnlock()
	return logger
}
