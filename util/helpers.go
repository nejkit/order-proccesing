package util

import (
	"order-processing/external/balances"
	"order-processing/external/orders"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	koeficientMap = map[orders.OrderType]float32{
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

func ParseCurrencyFromDirection(pair string, dir int) string {
	if dir == int(orders.Direction_DIRECTION_TYPE_BUY) {
		return strings.Split(pair, "/")[0]
	}
	return strings.Split(pair, "/")[1]

}

func GetLockBalanceRequest(request *orders.CreateOrderRequest, id string) *balances.LockBalanceRequest {
	amount := request.GetInitVolume()

	if request.GetDirection() == orders.Direction_DIRECTION_TYPE_BUY {
		multiplier, _ := koeficientMap[request.GetOrderType()]
		amount = (request.GetInitPrice() * request.GetInitVolume() * multiplier)
	}
	cur := ParseCurrencyFromDirection(request.CurrencyPair, int(request.Direction))
	event := &balances.LockBalanceRequest{
		Id:       id,
		Address:  request.GetExchangeWallet(),
		Currency: cur,
		Amount:   amount,
	}
	log.Infoln("Event to balance-service: ", event.String())
	return event

}
