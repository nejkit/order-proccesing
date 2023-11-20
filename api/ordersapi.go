package api

import (
	"context"
	"order-processing/external/orders"
	"order-processing/services"
	transportrabbit "order-processing/transport_rabbit"

	"github.com/sirupsen/logrus"
)

type OrderApi struct {
	logger     *logrus.Logger
	ordersServ services.OrderService
	rmqpSender transportrabbit.AmqpSender
}

func NewOrderApi(logger *logrus.Logger, oserv services.OrderService, sender transportrabbit.AmqpSender) OrderApi {
	return OrderApi{logger: logger, ordersServ: oserv, rmqpSender: sender}
}

func (api *OrderApi) CreateOrder(ctx context.Context, request *orders.CreateOrderRequest) {
	response := api.ordersServ.CreateOrder(ctx, request)
	api.rmqpSender.SendMessage(ctx, response)
}
