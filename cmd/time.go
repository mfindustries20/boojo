package cmd

import (
	"fmt"
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
		startTime := time.Now()
		fmt.Printf("Time tracking started at: %s\n", startTime.Format("15:04:05"))

		var totalPaused time.Duration
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
					fmt.Printf("\n%sTime tracking stopped%s at: %s", RED, RESET, time.Now().Format("15:04:05"))
					os.Exit(0)
				case 'p', 'P':
					if !paused {
						pauseStart = time.Now()
						paused = true
						fmt.Printf("\n%sTracking paused%s at: %s", YELLOW, RESET, pauseStart.Format("15:04:05"))
					}
				case 'r', 'R':
					if paused {
						paused = false
						pausedDuration := time.Since(pauseStart)
						totalPaused += pausedDuration
						fmt.Printf("\n%sTracking resumed%s at: %s\n", GREEN, RESET, time.Now().Format("15:04:05"))
					}
				}

				if key == keyboard.KeyEsc {
					fmt.Printf("\n%sTime tracking stopped%s at: %s", RED, RESET, time.Now().Format("15:04:05"))
					os.Exit(0)
				}
			}
		}()

		for {
			select {
			case <-ticker.C:
				if !paused {
					elapsed := time.Since(startTime) - totalPaused
					fmt.Printf("\rProgress: %02d:%02d", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
				}
			case <-quit:
				fmt.Println("\nTime tracking interrupted.")
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(timeCmd)
}
