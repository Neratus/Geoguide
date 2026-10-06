package mocks

import (
	"context"
	"sync"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
)

type MockTaskQueue struct {
	sync.RWMutex
	PublishedTasks []interfaces.Task
}

func NewMockTaskQueue() *MockTaskQueue {
	return &MockTaskQueue{PublishedTasks: []interfaces.Task{}}
}

func (q *MockTaskQueue) Publish(ctx context.Context, task interfaces.Task) error {
	q.Lock()
	defer q.Unlock()
	q.PublishedTasks = append(q.PublishedTasks, task)
	return nil
}

func (q *MockTaskQueue) Subscribe(taskName string, handler interfaces.TaskHandler) error { return nil }
func (q *MockTaskQueue) Run(ctx context.Context) error                                   { return nil }
func (q *MockTaskQueue) Close() error                                                    { return nil }
