package api

import (
	"context"
	"order-processing/external/orders"
	"order-processing/services"
	transportrabbit "order-processing/transport_rabbit"
	"order-processing/util"

	logger "github.com/sirupsen/logrus"
)

type OrderApi struct {
	ordersServ        services.OrderService
	matcherServ       *services.MatcherService
	createOrderSender transportrabbit.AmqpSender
	getOrderSender    transportrabbit.AmqpSender
}

func NewOrderApi(oserv services.OrderService, matcherserv *services.MatcherService, cos transportrabbit.AmqpSender, gos transportrabbit.AmqpSender) OrderApi {
	return OrderApi{ordersServ: oserv, createOrderSender: cos, getOrderSender: gos, matcherServ: matcherserv}
}

func (api *OrderApi) GetOrder(ctx context.Context, request *orders.GetOrderRequest) {
	data, err := api.ordersServ.GetOrder(ctx, request)

	if err != nil {
		response := orders.GetOrderResponse{
			Id:        request.GetId(),
			OrderData: nil,
		}
		api.getOrderSender.SendMessage(ctx, &response)
	}
	response := orders.GetOrderResponse{
		Id:        request.GetId(),
		OrderData: util.ConvertOrderModelToProto(data),
	}
	logger.Infoln("Send response: ", response.String())

	api.getOrderSender.SendMessage(ctx, &response)
}
