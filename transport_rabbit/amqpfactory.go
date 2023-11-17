package transportrabbit

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type AmqpFactory struct {
	logger     *logrus.Logger
	connection *amqp091.Connection
}

func NewFactory(logger *logrus.Logger, connectionString string) AmqpFactory {
	var con *amqp091.Connection
	var err error
	for {
		con, err = amqp091.Dial(connectionString)
		if err != nil {
			logger.Errorln("Connection failed, message: ", err.Error())
			time.Sleep(5 * time.Second)
		}
		break
	}
	return AmqpFactory{logger: logger, connection: con}
}

func (f *AmqpFactory) getRmqChannel() (*amqp091.Channel, error) {
	ch, err := f.connection.Channel()
	if err != nil {
		f.logger.Errorln("Channel doesn`t created, message: ", err.Error())
		return nil, err
	}
	return ch, nil
}

func (f *AmqpFactory) NewSender(ctx context.Context, ex string, rk string) (*AmqpSender, error) {
	ch, err := f.getRmqChannel()
	if err != nil {
		return nil, err
	}
	return &AmqpSender{logger: f.logger, channel: ch, ex: ex, rk: rk}, nil
}
