# Taskly

Taskly is a small command-line todo app written in Go. It stores tasks locally in `.data/tasks.json`, assigns each task an ID, and displays tasks in a table.

> [!NOTE]
> This is a learning project for practicing Go fundamentals; it is not intended to be production-ready.

## Requirements

- Go 1.27.1 or later, as declared in `go.mod`.

## Build and run

From the project directory, build the executable:

```sh
go build -o taskly .
```

Run it with:

```sh
./taskly ls
```

The `.data` directory and `tasks.json` are created automatically when a task is first saved.

## Commands

| Command | Description |
| --- | --- |
| `taskly add "<title>"` | Create a new task. |
| `taskly ls` | List all tasks. |
| `taskly edit <id> "<new title>"` | Change a task title. |
| `taskly delete <id>` | Delete a task. |
| `taskly check <id>` | Mark a task as complete. |
| `taskly uncheck <id>` | Mark a task as incomplete. |