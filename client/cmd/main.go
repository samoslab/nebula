package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spolabs/nebula/client"
)

func NewLogger(logFilename string, debug bool) (*logrus.Logger, error) {
	log := logrus.New()
	log.Out = os.Stdout
	log.Formatter = &logrus.TextFormatter{
		FullTimestamp:    true,
		QuoteEmptyFields: true,
	}
	log.Level = logrus.InfoLevel

	if debug {
		log.Level = logrus.DebugLevel
	}

	return log, nil
}
func main() {
	log, err := NewLogger("", true)
	if err != nil {
		return
	}
	cm, err := client.NewClientManager(log)
	if err != nil {
		return
	}
	log.Infof("start client")
	cm.NodeId = []byte("test")
	cm.TempDir = "/tmp/"
}
