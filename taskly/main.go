package main

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
)

func printTasks() error {
	tasks, err := getTasks()
	if err != nil {
		return fmt.Errorf("getting tasks: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tDone\tTitle")
	fmt.Fprintln(w, "--\t----\t----")

	for _, task := range tasks {
		done := "[ ]"
		if task.IsCompleted {
			done = "[X]"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\n", task.ID, done, task.Title)
	}

	return w.Flush()
}

func parseTaskID(value string) (int, error) {
	id, err := strconv.Atoi(value)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid task ID %q", value)
	}
	return id, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Expected Usage: taskly <cmd> <parameters>")
		os.Exit(1)
	}
	cmd := os.Args[1]
	switch {
	case cmd == "add":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "Task Title is Required")
			os.Exit(1)
		}
		_, err := createTask(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error while creating a task:", err)
			os.Exit(1)
		}
		if err := printTasks(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case cmd == "ls":
		if err := printTasks(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case cmd == "edit":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "Expected usage: taskly edit <id> <value>")
			os.Exit(1)
		}
		id, err := parseTaskID(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		newTitle := os.Args[3]

		_, err = updateTask(id, newTitle)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := printTasks(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case cmd == "delete":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "Task ID is Required to Delete")
			os.Exit(1)
		}
		id, err := parseTaskID(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := deleteTask(id); err != nil {
			fmt.Fprintf(os.Stderr, "Error while deleting task %d: %v\n", id, err)
			os.Exit(1)
		}
		if err := printTasks(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case cmd == "check":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "Task ID is required to check a task")
			os.Exit(1)
		}
		id, err := parseTaskID(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := completeTask(id); err != nil {
			fmt.Fprintf(os.Stderr, "Error while checking the task %d: %v\n", id, err)
			os.Exit(1)
		}
		if err := printTasks(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case cmd == "uncheck":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "Task ID is required to uncheck a task")
			os.Exit(1)
		}
		id, err := parseTaskID(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := incompleteTask(id); err != nil {
			fmt.Fprintf(os.Stderr, "Error while unchecking the task %d: %v\n", id, err)
			os.Exit(1)
		}
		if err := printTasks(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}

}
