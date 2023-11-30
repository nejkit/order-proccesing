package util

import (
	"order-processing/external/orders"
	"strings"
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
