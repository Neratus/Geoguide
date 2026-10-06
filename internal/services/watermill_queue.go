package services

import (
	"context"
	"encoding/json"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

type WatermillQueue struct {
	publisher  message.Publisher
	subscriber message.Subscriber
	router     *message.Router
	handlers   map[string]interfaces.TaskHandler
}

func NewWatermillQueue(logger watermill.LoggerAdapter) (*WatermillQueue, error) {
	pubSub := gochannel.NewGoChannel(gochannel.Config{
		OutputChannelBuffer: 1024,
	}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	return &WatermillQueue{
		publisher:  pubSub,
		subscriber: pubSub,
		router:     router,
		handlers:   make(map[string]interfaces.TaskHandler),
	}, nil
}

func (w *WatermillQueue) Publish(ctx context.Context, task interfaces.Task) error {
	payload, err := json.Marshal(task.Payload())
	if err != nil {
		return err
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	return w.publisher.Publish(task.Name(), msg)
}

func (w *WatermillQueue) Subscribe(taskName string, handler interfaces.TaskHandler) error {
	w.handlers[taskName] = handler
	w.router.AddHandler(
		taskName+"_handler",
		taskName,
		w.subscriber,
		taskName+"_output",
		w.publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			ctx := context.Background()
			if err := handler(ctx, msg.Payload); err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
	return nil
}

func (w *WatermillQueue) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		w.Close()
	}()
	return w.router.Run(ctx)
}

func (w *WatermillQueue) Close() error {
	return w.router.Close()
}
