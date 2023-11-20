package transportrabbit

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type ParserFunc[T any] func([]byte) (*T, error)
type GetMessageId[T any] func(*T) string
type HandlerFunc[T any] func(context.Context, *T)

type Listener[T any] struct {
	channel   *amqp091.Channel
	logger    *logrus.Logger
	messages  <-chan amqp091.Delivery
	parser    ParserFunc[T]
	handler   HandlerFunc[T]
	messageId GetMessageId[T]
}

func NewListener[T any](ctx context.Context, factory AmqpFactory, gName string, parser ParserFunc[T], handler HandlerFunc[T], idfunc GetMessageId[T]) *Listener[T] {
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

func (l *Listener[T]) ConsumeById(ctx context.Context, id string) *T {
	for {
		for msg := range l.messages {
			body, err := l.parser(msg.Body)
			if err != nil {
				l.logger.Warningln("Error while parse a message! message: ", err.Error())
				msg.Nack(false, false)
				continue
			}
			if l.messageId(body) == id {
				msg.Ack(false)
				return body
			}
			select {
			case <-time.After(time.Minute * 1):
				l.logger.Errorln("Fail consume response, timeout error")
				return nil
			default:
				time.Sleep(time.Millisecond * 1)
			}
		}
	}
}
