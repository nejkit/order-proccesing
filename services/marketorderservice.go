package services

import (
	"context"
	"order-processing/external/orders"
	"order-processing/storage"

	"github.com/sirupsen/logrus"
)

type MarketOrderService struct {
	logger *logrus.Logger
	store  storage.OrderManager
}

func NewMarketOrderService(logger *logrus.Logger, store storage.OrderManager) MarketOrderService {
	return MarketOrderService{logger: logger, store: store}
}

func (s *MarketOrderService) CreateOrder(ctx context.Context, request *orders.CreateOrderRequest) *orders.CreateOrderResponse {

	return nil
}
