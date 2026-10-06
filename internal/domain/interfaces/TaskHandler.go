package interfaces

import "context"

type Task interface {
	Name() string
	Payload() interface{}
}

type TaskHandler func(ctx context.Context, payload []byte) error

type TaskQueue interface {
	Publish(ctx context.Context, task Task) error
	Subscribe(taskName string, handler TaskHandler) error
	Run(ctx context.Context) error
	Close() error
}
