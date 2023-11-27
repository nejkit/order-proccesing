package util

import (
	"errors"
	"order-processing/external/balances"
	"order-processing/external/orders"
	"order-processing/statics"
)

func MapError(err error) orders.ErrorCodes {
	switch err.Error() {
	case statics.ErrorOrderNotFound:
		return orders.ErrorCodes_ERROR_CODE_ORDER_NOT_FOUND
	case statics.ErrorNotEnoughBalance:
		return orders.ErrorCodes_ERROR_CODE_NOT_ENOUGH_BALANCE
	case statics.ErrorBalanceNotExists:
		return orders.ErrorCodes_ERROR_CODE_NOT_EXISTS_CURRENCY
	default:
		return orders.ErrorCodes_ERROR_CODE_INTERNAL
	}
}

func ConvertBalanceError(err balances.ErrorCodes) error {
	switch err {
	case balances.ErrorCodes_ERROR_CODE_NOT_ENOUGH_BALANCE:
		return errors.New(statics.ErrorNotEnoughBalance)
	case balances.ErrorCodes_ERROR_CODE_NOT_EXISTS_BALANCE:
		return errors.New(statics.ErrorBalanceNotExists)
	default:
		return errors.New(statics.InternalError)
	}

}
