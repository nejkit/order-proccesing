package services

import (
	"context"
	"order-processing/external/balances"
	"order-processing/external/orders"
	transportrabbit "order-processing/transport_rabbit"
)

var (
	koeficientMap = map[orders.OrderType]float32{
		orders.OrderType_ORDER_TYPE_LIMIT:  1,
		orders.OrderType_ORDER_TYPE_MARKET: 1.15,
	}
)

type BalanceService struct {
	lockSender     transportrabbit.AmqpSender
	transferSender transportrabbit.AmqpSender
}

func NewBalanceService(
	lockSender transportrabbit.AmqpSender,
	transferSender transportrabbit.AmqpSender) BalanceService {
	return BalanceService{lockSender: lockSender, transferSender: transferSender}
}

func (s *BalanceService) LockBalance(ctx context.Context, request *balances.LockBalanceRequest) {
	s.lockSender.SendMessage(ctx, request)
}

func (s *BalanceService) CreateTransfer(ctx context.Context, request *balances.CreateTransferRequest) {
	s.transferSender.SendMessage(ctx, request)
}
