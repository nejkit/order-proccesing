package handlers

import (
	"context"
	"order-processing/api"
	"order-processing/external/balances"
	"order-processing/external/orders"
	"order-processing/external/tickets"
	"order-processing/storage"
)

type Handler struct {
	createapi   api.OrderApi
	ticketStore storage.TicketStorage
}

func NewHandler(api api.OrderApi, ticketStore storage.TicketStorage) Handler {
	return Handler{createapi: api, ticketStore: ticketStore}
}

func (h *Handler) GetHandlerForCreateOrder() func(context.Context, *orders.CreateOrderRequest) {
	return func(ctx context.Context, cor *orders.CreateOrderRequest) {
		h.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_CREATE_ORDER, cor)
	}
}

func (h *Handler) GetHandlerForGetOrder() func(context.Context, *orders.GetOrderRequest) {
	return func(ctx context.Context, gor *orders.GetOrderRequest) {
		h.createapi.GetOrder(ctx, gor)
	}
}

func (h *Handler) GetHandlerForLockBalance() func(context.Context, *balances.LockBalanceResponse) {
	return func(ctx context.Context, lbr *balances.LockBalanceResponse) {
		h.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_APPROVE_CREATION, lbr)
	}
}

func (h *Handler) GetHandlerForTransfer() func(context.Context, *balances.Transfer) {
	return func(ctx context.Context, t *balances.Transfer) {
		h.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_TRANSFER, t)
	}
}

func (h *Handler) GetHandlerForDeleteOrder() func(context.Context, *orders.DeleteOrderRequest) {
	return func(ctx context.Context, dor *orders.DeleteOrderRequest) {
		h.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_DROP_ORDER, dor)
	}
}
