package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Struct for lines with priority
type task struct {
	_line       string
	id          int
	description string
	priority    int
	status      taskStatusType
	layout      layoutType
	projects    []string
	contexts    []string
	tags        []string
	createdAt   *time.Time
	completedAt *time.Time
	dueAt       *time.Time
	recurrence  recurrence
	effort      float32
}

type statistics struct {
	fileName            string
	logType             string
	totalFileLines      int // Total lines of file
	totalFileEntries    int // Total entries in file (w/o blank and not parsable lines)
	totalEntries        int // Total (filtered) entries
	totalTasks          int
	totalTasksCompleted int
	totalTasksCancelled int
	totalNotes          int
	projectTags         map[string]int
	contextTags         map[string]int
	filters             map[string]int
	totalEffort         float32
}

var tasks []task
var stats statistics
var filterAll bool
var displayMeta bool
var dueRegex = regexp.MustCompile(` due:(\d{4}-\d{2}-\d{2})`)
var effortRegex = regexp.MustCompile(` ph:([0-9]\.\d{1,3})`)
var recurrenceRegex = regexp.MustCompile(` rec:(\+)?(\d+)(d|b|w|m|y)`)

var listCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all entries from the specified log",
	Run: func(cmd *cobra.Command, args []string) {
		fileName, err := getLogFilePath()
		entries, err := readAndProcessFile(fileName)
		if err != nil {
			fmt.Println("Error reading file:", err)
		}

		stats.filters = map[string]int{}
		for _, key := range args {
			stats.filters[key] = 0
		}
		if len(stats.filters) > 0 {
			entries = filterEntries(entries, stats.filters)
		}
		sortEntries(entries)
		summarizeEntries(entries)
		printEntries(entries)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&logType, "log", "l", "", "Log type (daily, monthly, future)")
	listCmd.Flags().BoolVarP(&filterAll, "all", "a", false, "Display all entries")
	listCmd.Flags().BoolVarP(&displayMeta, "meta", "m", false, "Display extra line with meta infos (with key/value pairs, creation date, completion date etc.)")
}

func readAndProcessFile(fileName string) ([]task, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	taskRegEx := regexp.MustCompile(`^\s*(?P<status>[-·.x/]) (?P<completed>\d{4}-\d{2}-\d{2}\s)?(?P<priority>\([ABC]\)\s)?(?P<created>\d{4}-\d{2}-\d{2}\s)?`)
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++
		if taskRegEx.MatchString(line) {
			matches := taskRegEx.FindStringSubmatch(line)
			status := OPEN
			layout := TASK
			if len(matches) > 1 {
				if matches[1] == "x" {
					status = COMPLETED
				} else if matches[1] == "/" {
					status = CANCELLED
				} else if matches[1] == "-" {
					layout = NOTE
				}
			}
			priority := 9
			if matches[3] == "(A) " {
				priority = 1
			} else if matches[3] == "(B) " {
				priority = 2
			} else if matches[3] == "(C) " {
				priority = 3
			}
			var creationDate, completionDate *time.Time
			if matches[2] != "" {
				if matches[4] == "" {
					parsedCreationDate, err := time.Parse("2006-01-02", strings.TrimSpace(matches[2]))
					if err == nil {
						creationDate = &parsedCreationDate
					}
				} else {
					parsedCreationDate, err := time.Parse("2006-01-02", strings.TrimSpace(matches[2]))
					if err == nil {
						completionDate = &parsedCreationDate
					}
					parsedCompletionDate, err := time.Parse("2006-01-02", strings.TrimSpace(matches[4]))
					if err == nil {
						creationDate = &parsedCompletionDate
					}
				}
			}
			var dueDate *time.Time
			if dueMatch := dueRegex.FindStringSubmatch(line); dueMatch != nil {
				parsedDate, err := time.Parse("2006-01-02", dueMatch[1])
				if err == nil {
					dueDate = &parsedDate
				}
			}
			var rec recurrence
			if recurrenceMatch := recurrenceRegex.FindStringSubmatch(line); recurrenceMatch != nil {
				number, err := strconv.Atoi(recurrenceMatch[2])
				if err != nil {
				}
				rec = recurrence{
					_expr:    strings.TrimSpace(recurrenceMatch[0]),
					number:   int(number),
					unit:     recurrenceUnit(recurrenceMatch[3]),
					isStrict: recurrenceMatch[1] == "+",
				}
			}
			var effort float32
			if effortMatch := effortRegex.FindStringSubmatch(line); effortMatch != nil {
				parsedFloat, err := strconv.ParseFloat(effortMatch[1], 32)
				if err == nil {
					effort = float32(parsedFloat)
				}
			}

			// 'Clean' description string
			description := taskRegEx.ReplaceAllString(line, "")
			description = dueRegex.ReplaceAllString(description, "")

			t := task{_line: line, id: lineNumber, description: description, priority: priority, status: status, layout: layout, createdAt: creationDate, completedAt: completionDate, dueAt: dueDate, recurrence: rec, effort: effort}
			tasks = append(tasks, t)
			stats.totalFileEntries++
		}
	}
	stats.totalFileLines = lineNumber

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func filterEntries(entries []task, filters map[string]int) []task {
	filtered := []task{}
	for _, task := range entries {
		if !filterAll && task.status == COMPLETED {
			continue
		}
		line := strings.ToLower(task._line)
		matchesAll := true
		for filter, _ := range filters {
			if !strings.Contains(line, strings.ToLower(filter)) {
				matchesAll = false
				break
			}
		}
		if matchesAll {
			filtered = append(filtered, task)
		}
	}
	stats.totalEntries = len(filtered)
	return filtered
}

func sortEntries(entries []task) {
	sort.Slice(entries, func(i, j int) bool {
		// 1. Priority (ascending)
		if entries[i].priority != entries[j].priority {
			return entries[i].priority < entries[j].priority
		}

		// 2. layoutType: task < event < note
		layoutOrder := map[layoutType]int{
			TASK:  0,
			EVENT: 1,
			NOTE:  2,
		}
		if layoutOrder[entries[i].layout] != layoutOrder[entries[j].layout] {
			return layoutOrder[entries[i].layout] < layoutOrder[entries[j].layout]
		}

		// 3. taskStatusType: open < completed < cancelled
		statusOrder := map[taskStatusType]int{
			OPEN:      0,
			COMPLETED: 1,
			CANCELLED: 2,
		}
		if statusOrder[entries[i].status] != statusOrder[entries[j].status] {
			return statusOrder[entries[i].status] < statusOrder[entries[j].status]
		}

		// 4. dueAt (descending)
		// Sort nil-values first (at the end)
		if entries[i].dueAt != nil && entries[j].dueAt != nil {
			if !entries[i].dueAt.Equal(*entries[j].dueAt) {
				return entries[i].dueAt.After(*entries[j].dueAt)
			}
		} else if entries[i].dueAt != nil {
			return true // i hat Wert, j nicht → i kommt zuerst
		} else if entries[j].dueAt != nil {
			return false // j hat Wert, i nicht → j kommt zuerst
		}

		// 5. id (ascending)
		return entries[i].id < entries[j].id
	})
}

func summarizeEntries(entries []task) {
	for _, task := range entries {
		if task.effort > 0 {
			stats.totalEffort += task.effort
		}
		if task.status == COMPLETED {
			stats.totalTasksCompleted++
		} else if task.status == CANCELLED {
			stats.totalTasksCancelled++
		} else if task.layout == NOTE {
			stats.totalNotes++
		}
	}
}

func printEntries(entries []task) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayExpr := today.Format("2006-01-02")

	// Colors
	reset := "\033[0m"
	red := "\033[31m"
	gray := "\033[37m"
	green := "\033[32m"
	yellow := "\033[33m"
	blue := "\033[34m"
	magenta := "\033[35m"
	cyan := "\033[36m"
	//white := "\033[37m"

	//bold := "\033[1m"
	//underline := "\033[4m"
	//strike := "\033[9m"
	//italic := "\033[3m"

	taskCompletedRegex := regexp.MustCompile(`^[x]`)
	taskCancelledRegex := regexp.MustCompile(`^[/]`)
	taskOpenRegex := regexp.MustCompile(`^[\.]`)
	noteRegex := regexp.MustCompile(`^[-]`)
	priorityRegex := regexp.MustCompile(`\([ABC]\)`)
	contextRegex := regexp.MustCompile(`@[A-Za-z0-9ÄÖÜäöüß\-_]+`)
	projectRegex := regexp.MustCompile(`\+[A-Za-z0-9ÄÖÜäöüß\-_]+`)
	counterRegex := regexp.MustCompile(`\s#[A-Za-z0-9ÄÖÜäöüß\-_]+`)

	projectTags := map[string]int{}
	contextTags := map[string]int{}

	// @todo Fix maxLineNumberLen for filtered entries
	maxLineNumberLen := len(strconv.Itoa(len(entries)))

	for _, task := range entries {
		line := task._line

		// Remove expressions from line
		line = dueRegex.ReplaceAllString(line, "")
		dateRegex := regexp.MustCompile(`\d{4}-\d{2}-\d{2}\s`)
		line = dateRegex.ReplaceAllString(line, "")
		//line = counterRegex.ReplaceAllString(line, "")
		line = recurrenceRegex.ReplaceAllString(line, "")
		line = effortRegex.ReplaceAllString(line, "")

		line = taskCompletedRegex.ReplaceAllStringFunc(line, func(status string) string {
			return green + status + reset
		})
		line = taskCancelledRegex.ReplaceAllStringFunc(line, func(status string) string {
			return gray + status + reset
		})
		line = taskOpenRegex.ReplaceAllStringFunc(line, func(status string) string {
			return red + status + reset
		})
		line = noteRegex.ReplaceAllStringFunc(line, func(status string) string {
			return blue + status + reset
		})
		if priorityRegex.MatchString(line) {
			switch task.priority {
			case 1:
				line = priorityRegex.ReplaceAllStringFunc(line, func(task string) string {
					return red + task + reset
				})
			case 2:
				line = priorityRegex.ReplaceAllStringFunc(line, func(task string) string {
					return yellow + task + reset
				})
			case 3:
				line = priorityRegex.ReplaceAllStringFunc(line, func(task string) string {
					return cyan + task + reset
				})
			}
		}
		line = contextRegex.ReplaceAllStringFunc(line, func(context string) string {
			contextTags[context]++
			return blue + context + reset
		})
		line = projectRegex.ReplaceAllStringFunc(line, func(project string) string {
			projectTags[project]++
			return magenta + project + reset
		})
		line = counterRegex.ReplaceAllStringFunc(line, func(counter string) string {
			return gray + counter + reset
		})

		// Display infinity sign at the end of a line to mark recurring task
		if task.recurrence._expr != "" {
			line += " " + cyan + "∞" + reset
		}

		// Meta info line
		if displayMeta {
			meta := ""
			//meta = fmt.Sprintf("%s prio:%d", meta, task.priority)
			if task.effort != 0 {
				meta = fmt.Sprintf("%s %sph:%.2f%s", meta, green, task.effort, reset)
			}
			if task.dueAt != nil {
				meta = meta + " due:" + task.dueAt.Format("2006-01-02")
			}
			if task.createdAt != nil {
				meta = meta + gray + " created:" + task.createdAt.Format("2006-01-02") + reset
			}
			if task.completedAt != nil {
				meta = meta + gray + " completed:" + task.completedAt.Format("2006-01-02") + reset
			}
			if task.recurrence._expr != "" {
				meta = meta + gray + " " + task.recurrence._expr + reset
			}
			if meta != "" {
				line = line + "\n  " + strings.Repeat(" ", maxLineNumberLen) + meta
			}
		}

		// Highlight today
		line = strings.Replace(line, todayExpr, yellow+todayExpr+reset, -1)

		// Add due date diff in days
		//dueExpr := ""
		//if task.dueAt != nil {
		//	dueInDays := task.dueAt.Sub(today).Hours() / 24
		//	if dueInDays >= 0.0 {
		//		dueExpr = fmt.Sprintf(" %s+%.0fd%s ", cyan, math.Round(dueInDays), reset)
		//	} else {
		//		dueExpr = fmt.Sprintf(" %s%.0fd%s ", cyan, math.Round(dueInDays), reset)
		//	}
		//}

		// Add line number
		lineNumberFormat := strconv.Itoa(maxLineNumberLen)
		if task.status == COMPLETED {
			fmt.Printf("%s%0"+lineNumberFormat+"d%s %s\n", green, task.id, reset, line)
		} else if task.status == CANCELLED {
			fmt.Printf("%s%0"+lineNumberFormat+"d%s %s\n", gray, task.id, reset, line)
		} else if task.layout == "note" {
			fmt.Printf("%s%0"+lineNumberFormat+"d%s %s\n", blue, task.id, reset, line)
		} else {
			fmt.Printf("%s%0"+lineNumberFormat+"d%s %s\n", red, task.id, reset, line)
		}
	}
	totalTasks := stats.totalEntries - stats.totalNotes
	totalOpen := stats.totalEntries - stats.totalTasksCompleted - stats.totalTasksCancelled - stats.totalNotes
	fmt.Printf("\n%s log | %d/%d parsed line(s)\n",
		logType,
		stats.totalFileEntries,
		stats.totalFileLines,
	)
	if len(stats.filters) > 0 {
		fmt.Printf("%s%d filter(s): ", yellow, len(stats.filters))
		for key, _ := range stats.filters {
			fmt.Printf("\"%s\" ", key)
		}
		fmt.Printf("%s| ", reset)
	}
	fmt.Printf("%d task(s) | %s%d completed%s | %s%d open%s | %s%d cancelled%s | %s%d note(s)%s\n",
		totalTasks,
		green, stats.totalTasksCompleted, reset,
		red, totalOpen, reset,
		gray, stats.totalTasksCancelled, reset,
		blue, stats.totalNotes, reset,
	)
	fmt.Printf("%d project(s) %s%s%s\n", len(projectTags), magenta, printSortedTags(projectTags), reset)
	fmt.Printf("%d context(s) %s%s%s\n", len(contextTags), blue, printSortedTags(contextTags), reset)
	fmt.Printf("ph %s%.2f%s\n", green, stats.totalEffort, reset)
}

func printSortedTags(tags map[string]int) string {
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := ""
	for _, k := range keys {
		result += fmt.Sprintf("%s (%d) ", k, tags[k])
	}

	return result
}
