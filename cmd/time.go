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
		fmt.Printf("[%s] %sTime tracking started%s\n", startTime.Format("15:04:05"), GREEN, RESET)
		logger.Printf("[%s] Time tracking started\n", startTime.Format("15:04:05"))

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
					fmt.Printf("\n[%s] %sTime tracking stopped%s", time.Now().Format("15:04:05"), RED, RESET)
					fmt.Printf("\nTotal paused: %02d:%02d", int(totalPaused.Minutes()), int(totalPaused.Seconds())%60)
					fmt.Printf("\nTotal progress: %02d:%02d\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
					logger.Printf("[%s] Time tracking stopped\n", time.Now().Format("15:04:05"))
					logger.Printf("Total paused: %02d:%02d\n", int(totalPaused.Minutes()), int(totalPaused.Seconds())%60)
					logger.Printf("Total progress: %02d:%02d\n\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
					os.Exit(0)
				case 'p', 'P':
					if !paused {
						pauseStart = time.Now()
						paused = true
						fmt.Printf("\n[%s] %sTracking paused%s", time.Now().Format("15:04:05"), YELLOW, RESET)
						logger.Printf("[%s] Tracking paused\n", time.Now().Format("15:04:05"))
					}
				case 'r', 'R':
					if paused {
						paused = false
						pausedDuration := time.Since(pauseStart)
						totalPaused += pausedDuration
						fmt.Printf("\nPause duration: %02d:%02d", int(pausedDuration.Minutes()), int(pausedDuration.Seconds())%60)
						fmt.Printf("\n[%s] %sTracking resumed%s\n", time.Now().Format("15:04:05"), GREEN, RESET)
						logger.Printf("[%s] Tracking resumed\n", time.Now().Format("15:04:05"))
					}
				}

				if key == keyboard.KeyEsc {
					fmt.Printf("\n[%s] %sTime tracking stopped%s", time.Now().Format("15:04:05"), RED, RESET)
					fmt.Printf("\nTotal paused: %02d:%02d", int(totalPaused.Minutes()), int(totalPaused.Seconds())%60)
					fmt.Printf("\nTotal progress: %02d:%02d\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
					logger.Printf("[%s] Time tracking stopped\n", time.Now().Format("15:04:05"))
					logger.Printf("Total paused: %02d:%02d\n", int(totalPaused.Minutes()), int(totalPaused.Seconds())%60)
					logger.Printf("Total progress: %02d:%02d\n\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
					os.Exit(0)
				}
			}
		}()

		for {
			select {
			case <-ticker.C:
				if !paused {
					elapsed = time.Since(startTime) - totalPaused
					fmt.Printf("\rProgress: %02d:%02d", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
				}
			case <-quit:
				fmt.Printf("\n[%s] Time tracking interrupted", time.Now().Format("15:04:05"))
				fmt.Printf("\nTotal paused: %02d:%02d", int(totalPaused.Minutes()), int(totalPaused.Seconds())%60)
				fmt.Printf("\nTotal progress: %02d:%02d\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
				logger.Printf("[%s] Time tracking interrupted\n", time.Now().Format("15:04:05"))
				logger.Printf("Total paused: %02d:%02d\n", int(totalPaused.Minutes()), int(totalPaused.Seconds())%60)
				logger.Printf("Total progress: %02d:%02d\n\n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(timeCmd)
}
