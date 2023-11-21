package services

import (
	"context"
	"order-processing/external/orders"
	"order-processing/storage"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type OrderService struct {
	logger *logrus.Logger
	store  storage.OrderManager
	bs     BalanceService
}

func NewMarketOrderService(logger *logrus.Logger, store storage.OrderManager, balServ BalanceService) OrderService {
	return OrderService{logger: logger, store: store, bs: balServ}
}

func (s *OrderService) CreateOrder(ctx context.Context, request *orders.CreateOrderRequest) *orders.CreateOrderResponse {
	s.logger.Infoln("Received request: ", request.String())

	lockResponse := s.bs.LockBalance(ctx, request)

	if lockResponse != nil {
		s.logger.Warningln("Balance not locked. Reazon: ", lockResponse.Error())
		return &orders.CreateOrderResponse{
			Id:           request.GetId(),
			ErrorMessage: lockResponse.Error(),
		}
	}
	oid := uuid.NewString()
	orderData := storage.OrderInfo{
		Id:             oid,
		SellCurrency:   request.GetSellCurency(),
		BuyCurrency:    request.GetBuyCurrency(),
		Direction:      int(request.GetDirection()),
		InitPrice:      float64(request.GetInitPrice()),
		InitVolume:     float64(request.GetInitVolume()),
		ExchangeWallet: request.GetExchangeWallet(),
		CreationDate:   uint64(time.Now().UTC().UnixMilli()),
		ExpirationDate: uint64(time.Now().Add(24 * time.Hour).UTC().UnixMilli()),
		OrderState:     int(orders.OrderState_ORDER_STATE_NEW),
		OrderType:      int(request.GetOrderType()),
	}

	s.store.InsertNewOrder(ctx, orderData)

	return &orders.CreateOrderResponse{
		Id:      request.GetId(),
		OrderId: oid,
	}
}

func (s *OrderService) GetOrder(ctx context.Context, request *orders.GetOrderRequest) *orders.GetOrderResponse {
	if request.GetOrderId() != "" {
		data := s.store.GetOrderById(ctx, request.GetOrderId())
		protoData := parseDtoToProto(*data)
		return &orders.GetOrderResponse{Id: request.GetId(), OrderData: []*orders.OrderInfo{protoData}}
	} else {
		data := s.store.GetOrderByWallet(ctx, request.GetWalletId())
		var protoData []*orders.OrderInfo
		for _, oData := range data {
			convertedData := parseDtoToProto(*oData)
			protoData = append(protoData, convertedData)
		}
		return &orders.GetOrderResponse{Id: request.GetId(), OrderData: protoData}
	}

}

func parseDtoToProto(data storage.OrderInfo) *orders.OrderInfo {
	var matchInfos []*orders.MatchingData
	for _, val := range data.MatchInfo {
		matchInfos = append(matchInfos, &orders.MatchingData{
			FillPrice:  float32(val.FillPrice),
			FillVolume: float32(val.FillVolume),
			Date:       val.Date,
		})
	}
	protoData := orders.OrderInfo{
		Id:             data.Id,
		SellCurency:    data.SellCurrency,
		BuyCurrency:    data.BuyCurrency,
		Direction:      orders.Direction(data.Direction),
		InitPrice:      float32(data.InitPrice),
		InitVolume:     float32(data.InitVolume),
		OrderType:      orders.OrderType(data.OrderType),
		OrderState:     orders.OrderState(data.OrderState),
		MatchInfos:     matchInfos,
		CreationDate:   data.CreationDate,
		UpdatedDate:    data.UpdatedDate,
		ExpirationDate: data.ExpirationDate,
		ExchangeWallet: data.ExchangeWallet,
	}
	return &protoData
}
