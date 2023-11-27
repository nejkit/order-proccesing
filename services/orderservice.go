package services

import (
	"context"
	"order-processing/external/orders"
	"order-processing/storage"
	"time"

	"github.com/google/uuid"
	logger "github.com/sirupsen/logrus"
)

type OrderService struct {
	store *storage.OrderManager
	bs    BalanceService
}

func NewMarketOrderService(store *storage.OrderManager, balServ BalanceService) OrderService {
	return OrderService{store: store, bs: balServ}
}

func (s *OrderService) CreateOrder(ctx context.Context, request *orders.CreateOrderRequest) (*string, error) {
	logger.Infoln("Received request: ", request.String())

	lockResponse := s.bs.LockBalance(ctx, request)

	if lockResponse != nil {
		logger.Warningln("Balance not locked. error code: ", lockResponse.Error())
		return nil, lockResponse
	}
	oid := uuid.NewString()
	orderData := storage.OrderInfo{
		Id:             oid,
		CurrencyPair:   request.CurrencyPair,
		Direction:      int(request.GetDirection()),
		InitPrice:      float64(request.GetInitPrice()),
		InitVolume:     float64(request.GetInitVolume()),
		ExchangeWallet: request.GetExchangeWallet(),
		CreationDate:   uint64(time.Now().UTC().UnixMilli()),
		ExpirationDate: uint64(time.Now().Add(24 * time.Hour).UTC().UnixMilli()),
		OrderState:     int(orders.OrderState_ORDER_STATE_NEW),
		OrderType:      int(request.GetOrderType()),
	}

	err := s.store.InsertNewOrder(ctx, orderData)
	if err != nil {
		return nil, err
	}

	return &oid, nil
}

func (s *OrderService) GetOrder(ctx context.Context, request *orders.GetOrderRequest) ([]storage.OrderInfo, error) {
	data, err := s.store.GetOrderById(ctx, request.GetOrderId())
	return []storage.OrderInfo{*data}, err

}
