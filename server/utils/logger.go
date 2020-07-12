package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func InitializeLogging() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{PrettyPrint: true})
	log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)
}

type Fields log.Fields
type REntry struct {
	log.Entry
}

func (r *REntry) WithFields(fields Fields) *REntry {
	entry := r.Entry.WithFields(log.Fields(fields))
	newREntry := REntry{*entry}
	return &newREntry
}

func (r *REntry) WithField(key string, value interface{}) *REntry {
	entry := r.Entry.WithField(key, value)
	newREntry := REntry{*entry}
	return &newREntry
}

func Log() *REntry {
	return &REntry{*log.WithFields(log.Fields{})}
}
