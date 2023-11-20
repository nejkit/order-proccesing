package util

import (
	"order-processing/external/balances"
)

func GetIdFromLockBalanceResponse() func(*balances.LockBalanceResponse) string {
	return func(cor *balances.LockBalanceResponse) string {
		return cor.GetId()
	}
}
