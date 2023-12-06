package util

import (
	"errors"
	"order-processing/external/balances"
	"order-processing/external/orders"
	"order-processing/statics"
)

func MapError(err error) orders.OrdersErrorCodes {
	switch err.Error() {
	case statics.ErrorOrderNotFound:
		return orders.OrdersErrorCodes_ORDERS_ERROR_CODE_ORDER_NOT_FOUND
	case statics.ErrorNotEnoughBalance:
		return orders.OrdersErrorCodes_ORDERS_ERROR_CODE_NOT_ENOUGH_BALANCE
	case statics.ErrorBalanceNotExists:
		return orders.OrdersErrorCodes_ORDERS_ERROR_CODE_NOT_EXISTS_CURRENCY
	default:
		return orders.OrdersErrorCodes_ORDERS_ERROR_CODE_INTERNAL
	}
}

func ConvertBalanceError(err balances.BalancesErrorCodes) error {
	switch err {
	case balances.BalancesErrorCodes_BALANCE_ERROR_CODE_NOT_ENOUGH_BALANCE:
		return errors.New(statics.ErrorNotEnoughBalance)
	case balances.BalancesErrorCodes_BALANCE_ERROR_CODE_NOT_EXISTS_BALANCE:
		return errors.New(statics.ErrorBalanceNotExists)
	default:
		return errors.New(statics.InternalError)
	}

}
