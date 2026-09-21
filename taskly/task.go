package main

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	IsCompleted bool   `json:"is_completed"`
}

const tasksFile = ".data/tasks.json"

func getNextTaskID(tasks []Task) int {
	if len(tasks) == 0 {
		return 1
	}
	maxTask := slices.MaxFunc(tasks, func(a, b Task) int {
		return cmp.Compare(a.ID, b.ID)
	})
	return maxTask.ID + 1
}

func createTask(title string) (Task, error) {
	tasks, err := getTasks()
	if err != nil {
		return Task{}, err
	}
	newTask := Task{
		ID:          getNextTaskID(tasks),
		Title:       title,
		IsCompleted: false,
	}
	tasks = append(tasks, newTask)
	saveErr := saveTask(tasks)
	if saveErr != nil {
		return Task{}, saveErr
	}
	return newTask, nil
}

func saveTask(tasks []Task) error {
	if err := os.MkdirAll(".data", 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tasksFile, data, 0644)
}

func getTasks() ([]Task, error) {
	data, err := os.ReadFile(tasksFile)
	if errors.Is(err, os.ErrNotExist) {
		return []Task{}, nil // No Tasks
	}
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return []Task{}, nil
	}
	var tasks []Task
	err = json.Unmarshal(data, &tasks)
	return tasks, err
}

func getIndexOfTaskByID(tasks []Task, id int) int {
	for i, task := range tasks {
		if task.ID == id {
			return i
		}
	}
	return -1
}

func updateTask(id int, newTitle string) (Task, error) {
	tasks, getTasksErr := getTasks()
	if getTasksErr != nil {
		return Task{}, getTasksErr
	}
	taskIndex := getIndexOfTaskByID(tasks, id)
	if taskIndex == -1 {
		return Task{}, fmt.Errorf("no task with ID %d found", id)
	}
	tasks[taskIndex].Title = newTitle

	if err := saveTask(tasks); err != nil {
		return Task{}, fmt.Errorf("saving updated task: %w", err)
	}
	return tasks[taskIndex], nil
}

func deleteTask(id int) error {
	tasks, err := getTasks()
	if err != nil {
		return fmt.Errorf("getting tasks: %w", err)
	}

	taskIndex := getIndexOfTaskByID(tasks, id)

	if taskIndex == -1 {
		return fmt.Errorf("no task with ID %d", id)
	}

	tasks = append(tasks[:taskIndex], tasks[taskIndex+1:]...)

	if err := saveTask(tasks); err != nil {
		return fmt.Errorf("saving tasks after deletion: %w", err)
	}

	return nil
}

func completeTask(id int) error {
	return setTaskCompletion(id, true)
}

func incompleteTask(id int) error {
	return setTaskCompletion(id, false)
}

func setTaskCompletion(id int, isCompleted bool) error {
	tasks, err := getTasks()
	if err != nil {
		return fmt.Errorf("getting tasks: %w", err)
	}

	taskIndex := getIndexOfTaskByID(tasks, id)

	if taskIndex == -1 {
		return fmt.Errorf("no task with ID %d", id)
	}
	tasks[taskIndex].IsCompleted = isCompleted
	if err := saveTask(tasks); err != nil {
		return fmt.Errorf("saving updated task: %w", err)
	}
	return nil
}
