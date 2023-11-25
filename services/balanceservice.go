package services

import (
	"context"
	"errors"
	"order-processing/external/balances"
	"order-processing/external/orders"
	transportrabbit "order-processing/transport_rabbit"
	"strings"

	"github.com/google/uuid"
)

var (
	koeficientMap = map[orders.OrderType]float32{
		orders.OrderType_ORDER_TYPE_LIMIT:  1,
		orders.OrderType_ORDER_TYPE_MARKET: 1.15,
	}
)

type BalanceService struct {
	lockSender transportrabbit.AmqpSender
	storage    transportrabbit.AmqpStorage[balances.LockBalanceResponse]
}

func NewBalanceService(lockSender transportrabbit.AmqpSender, lockStorage transportrabbit.AmqpStorage[balances.LockBalanceResponse]) BalanceService {
	return BalanceService{lockSender: lockSender, storage: lockStorage}
}

func (s *BalanceService) LockBalance(ctx context.Context, request *orders.CreateOrderRequest) error {
	var amount float32
	switch request.GetDirection() {
	case orders.Direction_DIRECTION_TYPE_SELL:
		{
			amount = request.GetInitVolume()
		}
	case orders.Direction_DIRECTION_TYPE_BUY:
		{
			multiplier, _ := koeficientMap[request.GetOrderType()]
			amount = (request.GetInitPrice() * request.GetInitVolume() * multiplier)
		}
	}

	id := uuid.NewString()
	cur := strings.Split(request.GetCurrencyPair(), "/")[request.GetDirection().Number()-1]
	event := &balances.LockBalanceRequest{
		Id:       id,
		Address:  request.GetExchangeWallet(),
		Currency: cur,
		Amount:   amount,
	}
	s.lockSender.SendMessage(ctx, event)
	response, err := s.storage.GetMessage(id)
	if err != nil {
		return err
	}
	if response.GetState() == balances.LockBalanceStatus_DONE {
		return nil
	}
	return errors.New(response.GetErrorMessage())

}
