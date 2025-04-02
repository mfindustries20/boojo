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

const workDuration = 25 * time.Minute
const breakDuration = 5 * time.Minute

var subject string

var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "Start time tracking for a task",
	Run: func(cmd *cobra.Command, args []string) {
		logFile, err := os.OpenFile("timelog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error opening log file:", err)
			return
		}
		defer logFile.Close()

		logger := log.New(logFile, "", log.LstdFlags)

		startTime := time.Now()
		isWorkPeriod := true
		cycleStart := time.Now()

		if subject != "" {
			logger.Printf("Time tracking started for [%s]\n", subject)
		} else {
			logger.Printf("Time tracking started\n")
		}
		if subject != "" {
			fmt.Printf("[%s] %sTime tracking started%s for %s", startTime.Format("15:04:05"), GREEN, RESET, subject)
		} else {
			fmt.Printf("[%s] %sTime tracking started%s", startTime.Format("15:04:05"), GREEN, RESET)
		}
		fmt.Println()

		var totalPaused time.Duration
		var elapsed time.Duration
		var pauseStart time.Time
		paused := false
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
					if subject != "" {
						logger.Printf("Time tracking stopped for [%s]\n", subject)
					} else {
						logger.Printf("Time tracking stopped\n")
					}
					logger.Printf("Total paused: %02d:%02d\n", int(totalPaused.Minutes()), int(totalPaused.Seconds())%60)
					logger.Printf("Total progress: %02d:%02d\n\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
					if subject != "" {
						fmt.Printf("\n[%s] %sTime tracking stopped%s for %s", time.Now().Format("15:04:05"), RED, RESET, subject)
					} else {
						fmt.Printf("\n[%s] %sTime tracking stopped%s", time.Now().Format("15:04:05"), RED, RESET)
					}
					fmt.Printf("\nTotal paused: %02d:%02d", int(totalPaused.Minutes()), int(totalPaused.Seconds())%60)
					fmt.Printf("\nTotal progress: %02d:%02d\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
					os.Exit(0)
				case 'p', 'P':
					if !paused {
						pauseStart = time.Now()
						paused = true
						logger.Printf("Tracking paused\n")
						fmt.Printf("\n[%s] %sTracking paused%s", time.Now().Format("15:04:05"), YELLOW, RESET)
					}
				case 'r', 'R':
					if paused {
						paused = false
						pausedDuration := time.Since(pauseStart)
						totalPaused += pausedDuration
						logger.Printf("[Pause]: %02d:%02d", int(pausedDuration.Minutes()), int(pausedDuration.Seconds())%60)
						fmt.Printf("\n[Pause]: %02d:%02d", int(pausedDuration.Minutes()), int(pausedDuration.Seconds())%60)
						logger.Printf("Tracking resumed\n")
						fmt.Printf("\n[%s] %sTracking resumed%s\n", time.Now().Format("15:04:05"), GREEN, RESET)
					}
				}

				if key == keyboard.KeyEsc {
					if subject != "" {
						logger.Printf("Time tracking stopped for [%s]\n", subject)
					} else {
						logger.Printf("Time tracking stopped\n")
					}
					logger.Printf("Total paused: %02d:%02d\n", int(totalPaused.Minutes()), int(totalPaused.Seconds())%60)
					logger.Printf("Total progress: %02d:%02d\n\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
					if subject != "" {
						fmt.Printf("\n[%s] %sTime tracking stopped%s for %s", time.Now().Format("15:04:05"), RED, RESET, subject)
					} else {
						fmt.Printf("\n[%s] %sTime tracking stopped%s", time.Now().Format("15:04:05"), RED, RESET)
					}
					fmt.Printf("\nTotal paused: %02d:%02d", int(totalPaused.Minutes()), int(totalPaused.Seconds())%60)
					fmt.Printf("\nTotal progress: %02d:%02d\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
					os.Exit(0)
				}
			}
		}()

		for {
			select {
			case <-ticker.C:
				if !paused {
					elapsed = time.Since(cycleStart)
					if isWorkPeriod && elapsed >= workDuration {
						logger.Printf("Work period completed. Starting break.")
						fmt.Printf("\n[%s] ✅ Work period done. Time for a break!\n", time.Now().Format("15:04:05"))
						isWorkPeriod = false
						cycleStart = time.Now()
					} else if !isWorkPeriod && elapsed >= breakDuration {
						logger.Printf("Break completed. Starting new work period.")
						fmt.Printf("\n[%s] ☕ Break over. Back to work!\n", time.Now().Format("15:04:05"))
						isWorkPeriod = true
						cycleStart = time.Now()
					} else {
						minutes := int(elapsed.Minutes())
						seconds := int(elapsed.Seconds()) % 60
						mode := "Work"
						if !isWorkPeriod {
							mode = "Break"
						}
						fmt.Printf("\r[%s] %02d:%02d", mode, minutes, seconds)
					}
				}
			case <-quit:
				if subject != "" {
					logger.Printf("Time tracking interrupted for [%s]\n", subject)
				} else {
					logger.Printf("Time tracking interrupted\n")
				}
				logger.Printf("Total paused: %02d:%02d\n", int(totalPaused.Minutes()), int(totalPaused.Seconds())%60)
				logger.Printf("Total progress: %02d:%02d\n\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
				fmt.Printf("\n[%s] Time tracking interrupted", time.Now().Format("15:04:05"))
				fmt.Printf("\nTotal paused: %02d:%02d", int(totalPaused.Minutes()), int(totalPaused.Seconds())%60)
				fmt.Printf("\nTotal progress: %02d:%02d\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
				return
			}
		}
	},
}

func init() {
	timeCmd.Flags().StringVarP(&subject, "subject", "s", "", "Subject for this time tracking session")
	rootCmd.AddCommand(timeCmd)
}
