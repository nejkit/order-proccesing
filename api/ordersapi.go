package api

import (
	"context"
	"order-processing/external/orders"
	"order-processing/services"
	"order-processing/statics"
	"order-processing/storage"
	transportrabbit "order-processing/transport_rabbit"

	logger "github.com/sirupsen/logrus"
)

type OrderApi struct {
	ordersServ services.OrderService
	senders    map[int]transportrabbit.AmqpSender
}

func NewOrderApi(oserv services.OrderService, senders map[int]transportrabbit.AmqpSender) OrderApi {
	return OrderApi{ordersServ: oserv, senders: senders}
}

func (api *OrderApi) CreateOrder(ctx context.Context, request *orders.CreateOrderRequest) {
	oid, err := api.ordersServ.CreateOrder(ctx, request)
	sender, ok := api.senders[statics.CreateOrderSender]
	if !ok {
		logger.Errorln("Sender not initialized! ")
		panic("Not init sender")
	}
	if err != nil {
		//add middlewareS
		sender.SendMessage(ctx, &orders.CreateOrderResponse{Id: request.GetId(), ErrorMessage: err.Error()})
		return
	}

	sender.SendMessage(ctx, &orders.CreateOrderResponse{Id: request.GetId(), OrderId: *oid})
}

func (api *OrderApi) GetOrder(ctx context.Context, request *orders.GetOrderRequest) {
	data, err := api.ordersServ.GetOrder(ctx, request)
	sender, ok := api.senders[statics.GetOrderSender]
	if !ok {
		logger.Errorln("Sender not initialized! ")
		panic("Not init sender")
	}
	if err != nil {
		response := orders.GetOrderResponse{
			Id:        request.GetId(),
			OrderData: convertGetOrderResponse(data),
		}
		sender.SendMessage(ctx, &response)
	}
	response := orders.GetOrderResponse{
		Id:        request.GetId(),
		OrderData: nil,
	}
	logger.Infoln("Send response: ", response.String())

	sender.SendMessage(ctx, &response)
}

func convertGetOrderResponse(data []storage.OrderInfo) []*orders.OrderInfo {
	var protoData []*orders.OrderInfo
	for _, orderInfo := range data {
		var matchingInfo []*orders.MatchingData
		for _, mData := range orderInfo.MatchInfo {
			matchingInfo = append(matchingInfo, &orders.MatchingData{
				FillPrice:  float32(mData.FillPrice),
				FillVolume: float32(mData.FillVolume),
				Date:       mData.Date,
			})
		}
		protoData = append(protoData, &orders.OrderInfo{
			Id:             orderInfo.Id,
			CurrencyPair:   orderInfo.CurrencyPair,
			Direction:      orders.Direction(orderInfo.Direction),
			InitPrice:      float32(orderInfo.InitPrice),
			InitVolume:     float32(orderInfo.InitVolume),
			OrderType:      orders.OrderType(orderInfo.OrderType),
			OrderState:     orders.OrderState(orderInfo.OrderState),
			MatchInfos:     matchingInfo,
			CreationDate:   orderInfo.CreationDate,
			UpdatedDate:    orderInfo.UpdatedDate,
			ExpirationDate: orderInfo.ExpirationDate,
			ExchangeWallet: orderInfo.ExchangeWallet,
		})
	}
	return protoData
}
