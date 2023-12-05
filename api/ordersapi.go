package api

import (
	"context"
	"order-processing/external/orders"
	"order-processing/services"
	"order-processing/storage"
	transportrabbit "order-processing/transport_rabbit"
	"order-processing/util"
	"time"

	logger "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (api *OrderApi) CreateOrder(ctx context.Context, request *orders.CreateOrderRequest) {
	oid, err := api.ordersServ.CreateOrder(ctx, request)
	if err != nil {
		go api.createOrderSender.SendMessage(ctx, &orders.CreateOrderResponse{Id: request.GetId(), Error: &orders.OrderErrorMessage{
			ErorCode: util.MapError(err),
			Message:  err.Error(),
		}})

	}

	api.createOrderSender.SendMessage(ctx, &orders.CreateOrderResponse{Id: request.GetId(), OrderId: *oid})
	go api.matcherServ.MatchOrderById(ctx, *oid)
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
		OrderData: convertGetOrderResponse(data),
	}
	logger.Infoln("Send response: ", response.String())

	api.getOrderSender.SendMessage(ctx, &response)
}

func convertGetOrderResponse(orderInfo *storage.OrderInfo) *orders.OrderInfo {
	var protoData *orders.OrderInfo

	var matchingInfo []*orders.MatchingData
	for _, mData := range orderInfo.MatchInfo {
		matchingInfo = append(matchingInfo, &orders.MatchingData{
			FillPrice:  float32(mData.FillPrice),
			FillVolume: float32(mData.FillVolume),
			Date:       timestamppb.New(time.UnixMilli(int64(mData.Date))),
		})
	}
	protoData = &orders.OrderInfo{
		Id:             orderInfo.Id,
		CurrencyPair:   orderInfo.CurrencyPair,
		Direction:      orders.Direction(orderInfo.Direction),
		InitPrice:      float32(orderInfo.InitPrice),
		InitVolume:     float32(orderInfo.InitVolume),
		OrderType:      orders.OrderType(orderInfo.OrderType),
		OrderState:     orders.OrderState(orderInfo.OrderState),
		MatchInfos:     matchingInfo,
		CreationDate:   timestamppb.New(time.UnixMilli(int64(orderInfo.CreationDate))),
		UpdatedDate:    timestamppb.New(time.UnixMilli(int64(orderInfo.UpdatedDate))),
		ExpirationDate: timestamppb.New(time.UnixMilli(int64(orderInfo.ExpirationDate))),
		ExchangeWallet: orderInfo.ExchangeWallet,
	}
	return protoData
}
