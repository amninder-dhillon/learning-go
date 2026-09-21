package main

import "fmt"

type Queue []string

func (q *Queue) enqueue(val string) {
	*q = append(*q, val)
}

func (q *Queue) dequeue() (string, error) {
	if len(*q) == 0 {
		err := fmt.Errorf("Queue is empty")
		return "", err
	}
	front := (*q)[0]
	*q = (*q)[1:]
	return front, nil
}
