package cmd

import (
	"fmt"
	"path/filepath"
)

type layoutType string

const (
	TASK  layoutType = "task"
	EVENT layoutType = "event"
	NOTE  layoutType = "note"
)

type taskStatusType string

const (
	OPEN      taskStatusType = "open"
	COMPLETED taskStatusType = "completed"
	CANCELLED taskStatusType = "cancelled"
)

// Global variable for log type (used in multiple commands)
var logType string

// Helper function to get the log file path based on the log type
func getLogFilePath() (string, error) {
	if logType == "" {
		logType = "daily"
	}
	switch logType {
	case "daily":
		return filepath.Join("data", "daily.txt"), nil
	case "monthly":
		return filepath.Join("data", "monthly.txt"), nil
	case "future":
		return filepath.Join("data", "future.txt"), nil
	default:
		return "", fmt.Errorf("Unknown log type. Use --log <daily|monthly|future>")
	}
}

type recurrenceUnit string

const (
	WEEKDAY recurrenceUnit = "d"
	WORKDAY recurrenceUnit = "b"
	WEEK    recurrenceUnit = "w"
	MONTH   recurrenceUnit = "m"
	YEAR    recurrenceUnit = "y"
)

type recurrence struct {
	_expr    string
	number   int
	unit     recurrenceUnit
	isStrict bool
}
