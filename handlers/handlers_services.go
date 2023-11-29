package handlers

import (
	"context"
	"order-processing/external/balances"
	"order-processing/services"
	transportrabbit "order-processing/transport_rabbit"
)

func GetHandlerForLockBalanceProcessor(storage transportrabbit.AmqpStorage[balances.LockBalanceResponse]) func(context.Context, *balances.LockBalanceResponse) {
	return func(ctx context.Context, lbr *balances.LockBalanceResponse) {
		storage.AddMessage(lbr)
	}
}

func GetHandlerForTransferProcessor(matcherService *services.MatcherService) func(context.Context, *balances.Transfer) {
	return func(ctx context.Context, t *balances.Transfer) {
		matcherService.HandleTransfersResponse(ctx, t)
	}
}
