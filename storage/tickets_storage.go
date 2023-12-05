package storage

import (
	"context"
	"encoding/json"
	"errors"
	"order-processing/external/tickets"
	"order-processing/tickets/dto"

	"github.com/google/uuid"
	logger "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	TicketPrefix = "ticket:"
)

type TicketStorage struct {
	client     RedisClient
	instanceId string
}

func NewTicketStorage(client RedisClient) TicketStorage {
	return TicketStorage{client: client, instanceId: uuid.NewString()}
}

func (s *TicketStorage) SaveTicketForOperation(ctx context.Context, ticketType tickets.OperationType, msgBody protoreflect.ProtoMessage) error {
	ticketId := uuid.NewString()
	data, err := proto.Marshal(msgBody)
	if err != nil {
		return err
	}
	ticket := &tickets.Ticket{
		TicketId:      ticketId,
		State:         tickets.TicketState_TICKET_STATE_NEW,
		OperationType: ticketType,
		Data:          data,
	}
	ticketModel := dto.MapToModel(ticket)
	bytes, err := json.Marshal(ticketModel)
	result, err := s.client.SetNXKey(ctx, TicketPrefix+ticketId, string(bytes))
	if result {
		logger.Infoln("Success create ticket with id: ", ticketId)
		return nil
	}
	if err != nil {
		logger.Infoln("Failed creation ticket with id: ", ticketId, " Reazon: ", err.Error())
		return err
	}
	return nil
}

func (s *TicketStorage) GetTicketById(ctx context.Context, id string) (*tickets.Ticket, error) {
	data, err := s.client.GetKey(ctx, id)
	if err != nil {
		return nil, err
	}
	var ticketModel dto.TicketModel
	if err = json.Unmarshal([]byte(data), &ticketModel); err != nil {
		return nil, err
	}
	ticket := dto.MapToProto(ticketModel)

	return ticket, nil
}

func (s *TicketStorage) GetTickets(ctx context.Context) ([]string, error) {
	return s.client.GetKeysByPattern(ctx, TicketPrefix+"*")
}

func (s *TicketStorage) UpdateTicket(ctx context.Context, ticketInfo *tickets.Ticket) error {
	err := s.tryLockTicket(ctx, ticketInfo.TicketId)
	if err != nil {
		return err
	}
	ticketModel := dto.MapToModel(ticketInfo)
	data, err := json.Marshal(ticketModel)
	if err != nil {
		return err
	}
	if err = s.client.SetKey(ctx, TicketPrefix+ticketInfo.TicketId, string(data)); err != nil {
		return err
	}
	s.tryUnlockTicket(ctx, ticketInfo.TicketId)
	return nil

}

func (s *TicketStorage) tryLockTicket(ctx context.Context, id string) error {
	exists, err := s.client.SetNXKey(ctx, "lock_"+TicketPrefix+id, s.instanceId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("ResourceIsBlocked")
	}
	return nil
}

func (s *TicketStorage) tryUnlockTicket(ctx context.Context, id string) error {
	if err := s.client.DelKeyWithValue(ctx, "lock_"+TicketPrefix+id, s.instanceId); err != nil {
		return err
	}
	return nil
}

func (s *TicketStorage) DeleteTicket(ctx context.Context, id string) error {

	if err := s.tryLockTicket(ctx, id); err != nil {
		return err
	}
	s.client.DelKey(ctx, TicketPrefix+id)
	if err := s.tryUnlockTicket(ctx, id); err != nil {
		return nil
	}

	return nil
}
