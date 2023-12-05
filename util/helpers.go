package util

import (
	"math"
	"order-processing/external/balances"
	"order-processing/external/orders"
	"order-processing/storage"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	koeficientMap = map[orders.OrderType]float64{
		orders.OrderType_ORDER_TYPE_LIMIT:  1,
		orders.OrderType_ORDER_TYPE_MARKET: 1.15,
	}
)

func Min(a, b float64) float64 {
	if a <= b {
		return a
	}
	return b
}

func Round(a float64, b int) float64 {
	precision := math.Pow10(b)
	return math.Round(a*precision) / precision
}

func ParseCurrencyFromDirection(pair string, dir int) string {
	if dir == int(orders.Direction_DIRECTION_TYPE_BUY) {
		return strings.Split(pair, "/")[1]
	}
	return strings.Split(pair, "/")[0]

}

func GetLockBalanceRequest(request *orders.CreateOrderRequest, id string) *balances.LockBalanceRequest {
	amount := request.GetInitVolume()

	if request.GetDirection() == orders.Direction_DIRECTION_TYPE_BUY {
		multiplier, _ := koeficientMap[request.GetOrderType()]
		amount = multiplier * request.InitPrice * request.InitVolume
	}
	cur := ParseCurrencyFromDirection(request.CurrencyPair, int(request.Direction))
	event := &balances.LockBalanceRequest{
		Id:       id,
		Address:  request.GetExchangeWallet(),
		Currency: cur,
		Amount:   Round(amount, 2),
	}
	log.Infoln("Event to balance-service: ", event.String())
	return event

}

func GetUnLockBalanceRequest(orderInfo *storage.OrderInfo) *balances.UnLockBalanceRequest {
	amount := orderInfo.InitVolume

	if orderInfo.Direction == int(orders.Direction_DIRECTION_TYPE_BUY) {
		multiplier, _ := koeficientMap[orders.OrderType(orderInfo.OrderType)]
		amount = multiplier * orderInfo.InitPrice * orderInfo.InitVolume
	}
	cur := ParseCurrencyFromDirection(orderInfo.CurrencyPair, orderInfo.Direction)
	event := &balances.UnLockBalanceRequest{
		Id:       orderInfo.Id,
		Address:  orderInfo.ExchangeWallet,
		Currency: cur,
		Amount:   Round(amount, 2),
	}
	log.Infoln("Event to balance-service: ", event.String())
	return event
}

func ConvertOrderModelToProto(orderInfo *storage.OrderInfo) *orders.OrderInfo {
	var protoData *orders.OrderInfo

	protoData = &orders.OrderInfo{
		Id:             orderInfo.Id,
		CurrencyPair:   orderInfo.CurrencyPair,
		Direction:      orders.Direction(orderInfo.Direction),
		InitPrice:      orderInfo.InitPrice,
		InitVolume:     orderInfo.InitVolume,
		OrderType:      orders.OrderType(orderInfo.OrderType),
		OrderState:     orders.OrderState(orderInfo.OrderState),
		FillPrice:      orderInfo.FillPrice,
		FillVolume:     orderInfo.FillVolume,
		Date:           timestamppb.New(time.UnixMilli(int64(orderInfo.MatchingDate))),
		TransferId:     orderInfo.TransferId,
		CreationDate:   timestamppb.New(time.UnixMilli(int64(orderInfo.CreationDate))),
		UpdatedDate:    timestamppb.New(time.UnixMilli(int64(orderInfo.UpdatedDate))),
		ExpirationDate: timestamppb.New(time.UnixMilli(int64(orderInfo.ExpirationDate))),
		ExchangeWallet: orderInfo.ExchangeWallet,
	}
	return protoData
}

func ConvertProtoOrderToModel(orderInfo *orders.OrderInfo) storage.OrderInfo {
	var orderModel storage.OrderInfo

	orderModel = storage.OrderInfo{
		Id:           orderInfo.Id,
		CurrencyPair: orderInfo.CurrencyPair,
		Direction:    int(orderInfo.Direction),
		InitPrice:    orderInfo.InitPrice,
		InitVolume:   orderInfo.InitVolume,
		OrderType:    int(orderInfo.OrderType),
		OrderState:   int(orderInfo.OrderState),
		MatchingData: storage.MatchingData{
			FillVolume:   orderInfo.FillPrice,
			FillPrice:    orderInfo.FillPrice,
			MatchingDate: uint64(orderInfo.Date.AsTime().UnixMilli()),
			TransferId:   orderInfo.TransferId,
		},
		CreationDate:   uint64(orderInfo.CreationDate.AsTime().UnixMilli()),
		UpdatedDate:    uint64(orderInfo.UpdatedDate.AsTime().UnixMilli()),
		ExpirationDate: uint64(orderInfo.ExpirationDate.AsTime().UnixMilli()),
		ExchangeWallet: orderInfo.ExchangeWallet,
	}
	return orderModel
}
