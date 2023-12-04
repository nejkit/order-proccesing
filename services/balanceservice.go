package services

import (
	"context"
	"order-processing/external/balances"
	"order-processing/external/orders"
	transportrabbit "order-processing/transport_rabbit"
	"order-processing/util"

	"github.com/google/uuid"
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

func (s *BalanceService) LockBalance(ctx context.Context, request *orders.CreateOrderRequest) {
	amount := request.GetInitVolume()

	if request.GetDirection() == orders.Direction_DIRECTION_TYPE_BUY {
		multiplier, _ := koeficientMap[request.GetOrderType()]
		amount = (request.GetInitPrice() * request.GetInitVolume() * multiplier)
	}

	id := uuid.NewString()
	cur := util.ParseCurrencyFromDirection(request.GetCurrencyPair(), int(request.GetDirection()))
	event := &balances.LockBalanceRequest{
		Id:       id,
		Address:  request.GetExchangeWallet(),
		Currency: cur,
		Amount:   amount,
	}
	s.lockSender.SendMessage(ctx, event)
}

func (s *BalanceService) CreateTransfer(ctx context.Context, request *balances.CreateTransferRequest) {
	s.transferSender.SendMessage(ctx, request)
}
