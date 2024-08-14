package main

// log.go contains the implementation for how logs are used and implemented.
// Contains implementation for Log, LogStore, and LogType.

import (
	"fmt"
	"time"
)

// Log entries are the operations of the state machine.
// Logs are replicated to all raft node's state machines.
type Log struct {
	// Index is the log's index in the log entry.
	Index uint64

	// Term is the term of the log entry.
	Term uint64

	// Type is the type of log entry.
	Type LogType

	// Data is the data of the log entry.
	Data []byte

	// AppendedAt is the time the leader appended the log entry to its LogStore.
	AppendedAt time.Time
}

// LogStore is an interface for storing and retrieving logs.
type LogStore interface {
	// StoreLog stores a log entry.
	StoreLog(log *Log) error

	// StoreLogs stores multiple log entries.
	StoreLogs(logs []*Log) error

	// GetLog gets a log entry at a given index.
	GetLog(index uint64, log *Log) error

	// FirstIndex returns the first index written. 0 for no entries.
	FirstIndex() (uint64, error)

	// LastIndex returns the last index written. 0 for no entries.
	LastIndex() (uint64, error)

	// DeleteEntry deletes an entrry at an index.
	DeleteEntry(index uint64) error

	// DeleteRange deletes entries from start to end index.
	DeleteEntries(start, end uint64) error
}

// LogType defines the different type of log entries.
type LogType uint8

const (
	// LogCommand is an operation that is applied to the state machines.
	LogCommand LogType = iota

	// LogLeadership demonstrates leadership change.
	LogLeadership

	// LogConfiguration shows a membership configuration change.
	LogConfiguration
)

// String is the string representation of the LogType.
func (logType LogType) String() string {
	switch logType {
	case LogCommand:
		return "LogCommand"
	case LogLeadership:
		return "LogLeadership"
	case LogConfiguration:
		return "LogConfiguration"
	default:
		return fmt.Sprintf("%d", logType)
	}
}
