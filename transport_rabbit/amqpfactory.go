package transportrabbit

import (
	"context"
	"errors"
	"order-processing/statics"
	"time"

	"github.com/rabbitmq/amqp091-go"
	logger "github.com/sirupsen/logrus"
)

type AmqpFactory struct {
	connection *amqp091.Connection
}

func NewFactory(connectionString string) AmqpFactory {
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
	return AmqpFactory{connection: con}
}

func (f *AmqpFactory) InitRmq() {
	ch, _ := f.getRmqChannel()
	defer ch.Close()

	ch.ExchangeDeclare(statics.ExNameOrders, "topic", true, false, false, false, nil)
	ch.QueueDeclare(statics.QueueNameCreateOrderRequest, true, false, false, false, nil)
	ch.QueueDeclare(statics.QueueNameCreateOrderResponse, true, false, false, false, nil)
	ch.QueueBind(statics.QueueNameCreateOrderRequest, statics.RkCreateOrderRequest, statics.ExNameOrders, false, nil)
	ch.QueueBind(statics.QueueNameCreateOrderResponse, statics.RkCreateOrderResponse, statics.ExNameOrders, false, nil)

}

func (f *AmqpFactory) getRmqChannel() (*amqp091.Channel, error) {
	ch, err := f.connection.Channel()
	if err != nil {
		logger.Errorln("Channel doesn`t created, message: ", err.Error())
		return nil, errors.New(statics.InternalError)
	}
	return ch, nil
}

func (f *AmqpFactory) NewSender(ctx context.Context, ex string, rk string) (*AmqpSender, error) {
	ch, err := f.getRmqChannel()
	if err != nil {
		return nil, errors.New(statics.InternalError)
	}
	return &AmqpSender{channel: ch, ex: ex, rk: rk}, nil
}
