package util

import (
	"context"
	"order-processing/external/balances"
	transportrabbit "order-processing/transport_rabbit"
)

func GetHandlerForLockBalanceProcessor(storage transportrabbit.AmqpStorage[balances.LockBalanceResponse]) func(context.Context, *balances.LockBalanceResponse) {
	return func(ctx context.Context, lbr *balances.LockBalanceResponse) {
		storage.AddMessage(lbr)
	}
}
