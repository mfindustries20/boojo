# Boojo - Digital and Extended Bullet Lists

Boojo is a cli tool for maintaining digital and extended bullet lists - take care of your tasks, events and notes.

## Installation

Build and run local binary with

```shell
git clone https://github.com/zoeller-io/boojo.git
cd boojo
go build -o boojo
./boojo --help
```

## Use Cases

- Tasks
  - List all daily tasks (with optional filtering)
  - List all monthly tasks (with optional filtering)
  - Summarize effort, tags
  - Add a task
- Bookmarks
  - Add a link

## Format

`daily.txt` file format example:

```txt
. 2025-04-02 Publish the new +Blog post @business due:2025-04-02
x 2025-04-01 2025-03-31 Write a +Blog post @business due:2025-03-31 ph:2.25
- 2025-04-01 This is a note #report
```

### Task

Mandatory

- description
- (creation date is automatically set during task creation, format `YYYY-MM-DD`)

Optional

- Mark completion (with character `x`)
- Marc cancellation (with character `/`)
- Mark priority (uppercase character from A-Z enclosed in parentheses, e. g. `(A)`)
- Set completion date
- Set project tag(s) (with prefix `+`)
- Set context tag(s) (with prefix `@`)
- Set `key:value` tags to define additional metadata
- Set filter tag(s) (with prefix `#`)

### Key Value Tags

Special handling:

- Due date: `due:2024-09-10`
- Effort (in person-hours): `ph:0.5`
  - To track project efforts for later invoicing
- Recurrence: `rec:3m`
- Threshold: `t:2025-04-02`

Adapted from `todo.txt` format (s. https://github.com/todotxt/todo.txt)

#### Recurrence (planned)

Pattern examples:

- `rec:1d`: repeat this task every day; the next task’s due date will be one day after this task’s completion date
- `rec:+10b`: repeat this task every ten business days; the next task’s due date will be ten business days after this task’s due date
- `rec:+2w`: 2 weeks, strict recurrence

`SwiftoDo` format (s. https://swiftodoapp.com/todotxt-syntax/recurrence/)

### Note

- Mark note (with character `-`)

## Commands

### List tasks

```shell
./boojo ls
./boojo ls -a
./boojo ls -am
```

With filter:

```shell
go run main.go ls blog
```

### Add a task

```shell
go run main.go add ". Write a +blog article @home due:2024-10-12"
go run main.go add -l daily -p A ". Write a +blog article @home due:2024-10-12"
```

### Time tracking

- Start a task with time tracking by entering a command
- Display
  - Start time of progress
  - Progress (in minutes and seconds)
- Pause time tracking by Pressing key 'P'
- Continue time tracking by Pressing key 'R'
- Finish time tracking by Pressing 'Q' or 'Esc

Default values:

- List: `daily`
- No priority

## Useful Links

- Todo.txt specs:
  - https://github.com/todotxt/todo.txt
  - Recurrence: https://swiftodcostoapp.com/todotxt-syntax/recurrence/ 
- Colored output: https://twin.sh/articles/35/how-to-add-colors-to-your-console-terminal-output-in-go


## List of Open Points

- [ ] Get log file path by helper function
- [x] Add key-value tag to track efforts (e. g. `ph:0.5`)
- [x] Refactor statistics
- [x] Use enums for task status and entry layout
- [x] Add recurrence parsing
- [x] Modify sort of entries, display infinity sign to mark recurring task
- [x] Allow hyphen and underscore in regex for context, project and counter; add rootCmd for Cobra
- [x] Fix maxLineNumberLen for filtered entries
- [x] Set Version field in root command
- [x] Create command to track time
- [x] Log time tracking
- [x] Add Pomodoro cycles to time tracking