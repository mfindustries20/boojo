package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/spf13/cobra"
)

var subject string
var startTime time.Time
var totalElapsed time.Duration
var totalPaused time.Duration
var workCount int
var breakCount int
var pauseCount int

var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "Start time tracking for a task",
	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now()
		logFilePath := fmt.Sprintf("log/%s_timelog.txt", now.Format("060102"))
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error opening log file:", err)
			return
		}
		defer logFile.Close()

		logger := log.New(logFile, "", log.LstdFlags)

		const workDuration = 25 * time.Minute
		const breakDuration = 5 * time.Minute

		startTime = time.Now()
		startOutput(logger)
		fmt.Println()

		isWorkPeriod := true
		phaseElapsed := time.Duration(0)
		totalElapsed = time.Duration(0)
		totalPaused = time.Duration(0)
		var pauseStart time.Time
		paused := false
		workCount = 1
		breakCount = 0
		pauseCount = 0

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		// Handle interrupt signal to exit gracefully
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)

		// Keyboard setup
		if err := keyboard.Open(); err != nil {
			fmt.Println("Failed to listen for key events:", err)
			return
		}
		defer keyboard.Close()

		go func() {
			for {
				char, key, err := keyboard.GetKey()
				if err != nil {
					panic(err)
				}

				switch char {
				case 'q', 'Q':
					totalElapsed += phaseElapsed
					if paused {
						totalPaused += time.Since(pauseStart)
					}
					stopOutput(logger)
					os.Exit(0)
				case 'p', 'P':
					if !paused {
						pauseStart = time.Now()
						paused = true
						pauseCount++
						logger.Printf("Tracking paused\n")
						fmt.Printf("\n[%s] %sTracking paused%s (press R to resume or Q to quit)", time.Now().Format("15:04:05"), YELLOW, RESET)
					}
				case 'r', 'R':
					if paused {
						pausedDuration := time.Since(pauseStart)
						totalPaused += pausedDuration
						paused = false
						logger.Printf("[Pause] %dh %02dm %02ds", int(pausedDuration.Hours()), int(pausedDuration.Minutes())%60, int(pausedDuration.Seconds())%60)
						fmt.Printf("\n[Pause] %dh %02dm %02ds", int(pausedDuration.Hours()), int(pausedDuration.Minutes())%60, int(pausedDuration.Seconds())%60)
						logger.Printf("Tracking resumed\n")
						fmt.Printf("\n[%s] %sTracking resumed%s (press P to pause or Q to quit)\n", time.Now().Format("15:04:05"), GREEN, RESET)
					}
				}

				if key == keyboard.KeyEsc {
					stopOutput(logger)
					os.Exit(0)
				}
			}
		}()

		for {
			select {
			case <-ticker.C:
				if !paused {
					phaseElapsed += time.Second

					if isWorkPeriod && phaseElapsed >= workDuration {
						totalElapsed += phaseElapsed
						logger.Printf("Work session completed. Starting break.")
						fmt.Printf("\n[%s] Work session complete. %sTime for a ☕ break!%s (press P to pause or Q to quit)\n", time.Now().Format("15:04:05"), YELLOW, RESET)
						fmt.Print("\a")
						isWorkPeriod = false
						breakCount++
						phaseElapsed = 0
					} else if !isWorkPeriod && phaseElapsed >= breakDuration {
						logger.Printf("Break completed. Starting new work session.")
						fmt.Printf("\n[%s] Break session complete. %sBack to 💻 work!%s (press P to pause or Q to quit)\n", time.Now().Format("15:04:05"), GREEN, RESET)
						fmt.Print("\a")
						isWorkPeriod = true
						workCount++
						phaseElapsed = 0
					} else {
						hours := int(phaseElapsed.Hours())
						minutes := int(phaseElapsed.Minutes()) % 60
						seconds := int(phaseElapsed.Seconds()) % 60
						mode := "Work"
						if !isWorkPeriod {
							mode = "Break"
						}
						fmt.Printf("\r[%s] %dh %02dm %02ds", mode, hours, minutes, seconds)
					}
				}
			case <-quit:
				interruptOutput(logger)
				return
			}
		}
	},
}

func startOutput(logger *log.Logger) {
	if subject != "" {
		logger.Printf("Time tracking started for [%s]\n", subject)
	} else {
		logger.Printf("Time tracking started\n")
	}
	if subject != "" {
		fmt.Printf("[%s] %sTime tracking started%s for %s%s%s (press P to pause or Q to quit)", startTime.Format("15:04:05"), GREEN, RESET, GREEN, subject, RESET)
	} else {
		fmt.Printf("[%s] %sTime tracking started%s (press P to pause or Q to quit)", startTime.Format("15:04:05"), GREEN, RESET)
	}
}

func stopOutput(logger *log.Logger) {
	if subject != "" {
		logger.Printf("Time tracking stopped for [%s]\n", subject)
	} else {
		logger.Printf("Time tracking stopped\n")
	}
	if subject != "" {
		fmt.Printf("\n[%s] %sTime tracking stopped%s for %s%s%s", time.Now().Format("15:04:05"), RED, RESET, GREEN, subject, RESET)
	} else {
		fmt.Printf("\n[%s] %sTime tracking stopped%s", time.Now().Format("15:04:05"), RED, RESET)
	}
	summaryOutput(logger)
}

func interruptOutput(logger *log.Logger) {
	if subject != "" {
		logger.Printf("Time tracking interrupted for [%s]\n", subject)
	} else {
		logger.Printf("Time tracking interrupted\n")
	}
	if subject != "" {
		fmt.Printf("\n[%s] %sTime tracking interrupted%s for %s%s%s", time.Now().Format("15:04:05"), RED, RESET, GREEN, subject, RESET)
	} else {
		fmt.Printf("\n[%s] %sTime tracking interrupted%s", time.Now().Format("15:04:05"), RED, RESET)
	}
	summaryOutput(logger)
}

func summaryOutput(logger *log.Logger) {
	logger.Printf("Total worked: %dh %02dm %02ds in %d work session(s) with %d break(s)\n", int(totalPaused.Hours()), int(totalElapsed.Minutes())%60, int(totalElapsed.Seconds())%60, workCount, breakCount)
	logger.Printf("Total paused: %dh %02dm %02ds in %d manual pause(s)\n\n", int(totalPaused.Hours()), int(totalPaused.Minutes())%60, int(totalPaused.Seconds())%60, pauseCount)
	fmt.Printf("\nTotal worked: %dh %02dm %02ds in %d 💻 work session(s) with %d ☕ break(s)", int(totalPaused.Hours()), int(totalElapsed.Minutes())%60, int(totalElapsed.Seconds())%60, workCount, breakCount)
	fmt.Printf("\nTotal paused: %dh %02dm %02ds in %d manual pause(s)\n", int(totalPaused.Hours()), int(totalPaused.Minutes())%60, int(totalPaused.Seconds())%60, pauseCount)
}

func init() {
	timeCmd.Flags().StringVarP(&subject, "subject", "s", "", "Subject for this time tracking session")
	rootCmd.AddCommand(timeCmd)
}
