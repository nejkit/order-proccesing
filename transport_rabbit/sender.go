package transportrabbit

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	logger "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type AmqpSender struct {
	channel *amqp091.Channel
	ex      string
	rk      string
}

func (s *AmqpSender) SendMessage(ctx context.Context, message protoreflect.ProtoMessage) {
	body, err := proto.Marshal(message)
	if err != nil {
		logger.Errorln("Parsing message error, error: ", err.Error())
	}
	err = s.channel.PublishWithContext(ctx, s.ex, s.rk, false, false, amqp091.Publishing{ContentType: "text/plain", Body: body})
	if err != nil {
		logger.Errorln("Message not publish. Reazon: ", err.Error())
		return
	}
	logger.Info("Message was succesfully published! ")

}
