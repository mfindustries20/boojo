package cmd

import (
	"fmt"
	"path/filepath"
)

type cliColor string

const (
	RESET   cliColor = "\033[0m"
	RED     cliColor = "\033[31m"
	GRAY    cliColor = "\033[37m"
	GREEN   cliColor = "\033[32m"
	YELLOW  cliColor = "\033[33m"
	BLUE    cliColor = "\033[34m"
	MAGENTA cliColor = "\033[35m"
	CYAN    cliColor = "\033[36m"
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
