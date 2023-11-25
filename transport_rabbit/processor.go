package transportrabbit

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	logger "github.com/sirupsen/logrus"
)

type AmqpProcessor[T any] struct {
	parser  ParserFunc[T]
	handler HandlerFunc[T]
}

func NewAmqpProcessor[T any](handler HandlerFunc[T], parser ParserFunc[T]) AmqpProcessor[T] {
	return AmqpProcessor[T]{
		parser:  parser,
		handler: handler,
	}
}

func (p *AmqpProcessor[T]) ProcessMessage(ctx context.Context, msg amqp091.Delivery) {
	body, err := p.parser(msg.Body)
	if err != nil {
		logger.Warningln("Error parsing message, skipping...")
		msg.Nack(false, false)
		return
	}
	msg.Ack(false)
	p.handler(ctx, body)
}
