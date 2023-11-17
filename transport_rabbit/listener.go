package transportrabbit

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type ParserFunc[T any] func([]byte) (*T, error)

type HandlerFunc[T any] func(context.Context, *T)

type Listener[T any] struct {
	channel  *amqp091.Channel
	logger   *logrus.Logger
	messages <-chan amqp091.Delivery
	parser   ParserFunc[T]
	handler  HandlerFunc[T]
}

func NewListener[T any](ctx context.Context, factory AmqpFactory, gName string, parser ParserFunc[T], handler HandlerFunc[T]) *Listener[T] {
	rmqChannel, err := factory.getRmqChannel()
	if err != nil {
		return nil
	}
	messages, err := rmqChannel.ConsumeWithContext(ctx, gName, "", false, false, false, false, nil)
	if err != nil {
		factory.logger.Errorln("Failed create consumer to queue: ", gName, " message: ", err.Error())
		return nil
	}
	return &Listener[T]{
		channel:  rmqChannel,
		logger:   factory.logger,
		messages: messages,
		parser:   parser,
		handler:  handler,
	}
}

func (l *Listener[T]) Run(ctx context.Context) {
	for {
		select {
		case msg, ok := <-l.messages:
			{
				if !ok {
					l.logger.Errorln("Fail consume message! ")
					continue
				}
				body, err := l.parser(msg.Body)
				if err != nil {
					l.logger.Errorln("Invalid message format! ", err.Error())
					msg.Nack(false, false)
					continue
				}
				msg.Ack(false)
				go l.handler(ctx, body)
			}
		}
	}
}
