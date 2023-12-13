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
	unlockSender   transportrabbit.AmqpSender
	transferSender transportrabbit.AmqpSender
}

func NewBalanceService(
	lockSender transportrabbit.AmqpSender,
	transferSender transportrabbit.AmqpSender,
	unlockSender transportrabbit.AmqpSender) BalanceService {
	return BalanceService{lockSender: lockSender, transferSender: transferSender, unlockSender: unlockSender}
}

func (s *BalanceService) LockBalance(ctx context.Context, request *balances.LockBalanceRequest) error {
	if err := s.lockSender.SendMessage(ctx, request); err != nil {
		return err
	}
	return nil
}

func (s *BalanceService) CreateTransfer(ctx context.Context, request *balances.CreateTransferRequest) error {
	if err := s.transferSender.SendMessage(ctx, request); err != nil {
		return err
	}
	return nil
}

func (s *BalanceService) UnLockBalance(ctx context.Context, request *balances.UnLockBalanceRequest) error {
	if err := s.unlockSender.SendMessage(ctx, request); err != nil {
		return err
	}
	return nil
}
