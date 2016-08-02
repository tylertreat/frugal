package frugal

import (
	"flag"
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"
)

func init() {
	flag.Parse()
	logger := logrus.New()
	if testing.Verbose() {
		logger.Level = logrus.DebugLevel
	} else {
		logger.Out = ioutil.Discard
	}
	SetLogger(logger)
}
