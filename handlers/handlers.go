package handlers

import (
	"context"
	"order-processing/api"
	"order-processing/external/orders"
)

type Handler struct {
	createapi api.OrderApi
}

func NewHandler(api api.OrderApi) Handler {
	return Handler{createapi: api}
}

func (h *Handler) GetHandlerForCreateOrder() func(context.Context, *orders.CreateOrderRequest) {
	return func(ctx context.Context, cor *orders.CreateOrderRequest) {
		h.createapi.CreateOrder(ctx, cor)
	}
}

func (h *Handler) GetHandlerForGetOrder() func(context.Context, *orders.GetOrderRequest) {
	return func(ctx context.Context, gor *orders.GetOrderRequest) {
		h.createapi.GetOrder(ctx, gor)
	}
}
