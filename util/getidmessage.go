package util

import (
	"order-processing/external/balances"
)

func GetIdFromLockBalanceResponse() func(*balances.LockBalanceResponse) string {
	return func(cor *balances.LockBalanceResponse) string {
		return cor.GetId()
	}
}

func GetIdFromTransferResponse() func(*balances.Transfer) string {
	return func(t *balances.Transfer) string {
		return t.GetId()
	}
}
