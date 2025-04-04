package cmd

import (
	"fmt"
	"path/filepath"
)

type cliColor string

// S. https://gist.github.com/vratiu/9780109
const (
	RESET cliColor = "\033[0m"
	BLACK cliColor = "\033[0;30m"
	RED   cliColor = "\033[0;31m"

	GREEN  cliColor = "\033[0;32m"
	YELLOW cliColor = "\033[0;33m"
	BLUE   cliColor = "\033[0;34m"
	PURPLE cliColor = "\033[0;35m"
	CYAN   cliColor = "\033[0;36m"
	GRAY   cliColor = "\033[0;37m"

	ON_BLACK  cliColor = "\033[40m"
	ON_RED    cliColor = "\033[41m"
	ON_GREEN  cliColor = "\033[42m"
	ON_YELLOW cliColor = "\033[43m"
	ON_BLUE   cliColor = "\033[44m"
	ON_PURPLE cliColor = "\033[45m"
	ON_CYAN   cliColor = "\033[46m"
	ON_GRAY   cliColor = "\033[47m"
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
