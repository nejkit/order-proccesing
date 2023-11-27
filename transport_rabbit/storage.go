package transportrabbit

import (
	"errors"
	"order-processing/statics"
	"time"

	logger "github.com/sirupsen/logrus"
)

type AmqpStorage[T any] struct {
	messages  map[string]*T
	messageId GetMessageId[T]
}

func NewAmqpStorage[T any](idFunc GetMessageId[T]) AmqpStorage[T] {
	return AmqpStorage[T]{
		messageId: idFunc,
		messages:  make(map[string]*T),
	}
}

func (s *AmqpStorage[T]) AddMessage(msg *T) {
	id := s.messageId(msg)
	s.messages[id] = msg
}

func (s *AmqpStorage[T]) GetMessage(id string) (*T, error) {
	for i := 0; i < 60; i++ {
		msg, exists := s.messages[id]
		if exists {
			logger.Infoln("Message with id: ", id, " Was handled by service")
			delete(s.messages, id)
			return msg, nil
		}
		time.Sleep(1 * time.Second)
	}
	return nil, errors.New(statics.ErrorMessageNotFound)
}
