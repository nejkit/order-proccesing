package services

import (
	"context"
	"errors"
	"order-processing/external/balances"
	"order-processing/external/orders"
	transportrabbit "order-processing/transport_rabbit"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type BalanceService struct {
	logger       *logrus.Logger
	lockSender   transportrabbit.AmqpSender
	lockListener transportrabbit.Listener[balances.LockBalanceResponse]
}

func NewBalanceService(logger *logrus.Logger, lockSender transportrabbit.AmqpSender, lockListener transportrabbit.Listener[balances.LockBalanceResponse]) BalanceService {
	return BalanceService{logger: logger, lockSender: lockSender, lockListener: lockListener}
}

func (s *BalanceService) LockBalance(ctx context.Context, request *orders.CreateOrderRequest) error {
	var amount float32
	switch request.GetOrderType() {
	case orders.OrderType_ORDER_TYPE_MARKET:
		{
			if request.GetDirection() == orders.Direction_DIRECTION_TYPE_SELL {
				amount = request.GetInitVolume()
			} else {
				amount = (request.GetInitPrice() * request.GetInitVolume() * 1.1) //TODO: when quoutes up, fix me
			}
		}
	case orders.OrderType_ORDER_TYPE_LIMIT:
		{
			if request.GetDirection() == orders.Direction_DIRECTION_TYPE_SELL {
				amount = (request.GetInitVolume())
			} else {
				amount = (request.GetInitPrice() * request.GetInitVolume())
			}
		}
	}

	id := uuid.NewString()
	event := &balances.LockBalanceRequest{
		Id:       id,
		Address:  request.GetExchangeWallet(),
		Currency: request.GetSellCurency(),
		Amount:   amount,
	}
	s.lockSender.SendMessage(ctx, event)
	response := s.lockListener.ConsumeById(ctx, id)
	if response.GetState() == balances.LockBalanceStatus_DONE {
		return nil
	} else {
		return errors.New(response.GetErrorMessage())
	}
}
