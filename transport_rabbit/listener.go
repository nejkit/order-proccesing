package transportrabbit

import (
	"context"
	"errors"
	"order-processing/statics"

	"github.com/rabbitmq/amqp091-go"
	logger "github.com/sirupsen/logrus"
)

type ParserFunc[T any] func([]byte) (*T, error)
type GetMessageId[T any] func(*T) string
type HandlerFunc[T any] func(context.Context, *T)

type Listener[T any] struct {
	channel   *amqp091.Channel
	messages  <-chan amqp091.Delivery
	processor AmqpProcessor[T]
}

func NewListener[T any](ctx context.Context, factory AmqpFactory, gName string, processor AmqpProcessor[T]) (*Listener[T], error) {
	rmqChannel, err := factory.getRmqChannel()
	if err != nil {
		return nil, errors.New(statics.InternalError)
	}
	messages, err := rmqChannel.ConsumeWithContext(ctx, gName, "", false, false, false, false, nil)
	if err != nil {
		logger.Errorln("Failed create consumer to queue: ", gName, " message: ", err.Error())
		return nil, errors.New(statics.InternalError)
	}
	return &Listener[T]{
		channel:   rmqChannel,
		messages:  messages,
		processor: processor,
	}, nil
}

func (l *Listener[T]) Run(ctx context.Context) {
	for {
		select {
		case msg, ok := <-l.messages:
			{
				if !ok {
					logger.Errorln("Fail consume message! ")
					continue
				}
				go l.processor.ProcessMessage(ctx, msg)
			}
		}
	}
}
