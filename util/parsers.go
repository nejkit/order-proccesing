package util

import (
	"order-processing/external/balances"
	"order-processing/external/orders"

	"google.golang.org/protobuf/proto"
)

func GetParserForLockBalanceResponse() func([]byte) (*balances.LockBalanceResponse, error) {
	return func(b []byte) (*balances.LockBalanceResponse, error) {
		var request balances.LockBalanceResponse
		err := proto.Unmarshal(b, &request)
		if err != nil {
			return nil, err
		}
		return &request, nil
	}
}

func GetParserForCreateOrderRequest() func([]byte) (*orders.CreateOrderRequest, error) {
	return func(b []byte) (*orders.CreateOrderRequest, error) {
		var request orders.CreateOrderRequest
		err := proto.Unmarshal(b, &request)
		if err != nil {
			return nil, err
		}
		return &request, nil
	}
}
