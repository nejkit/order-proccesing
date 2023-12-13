package services

import (
	"context"
	"order-processing/external/balances"
	"order-processing/external/orders"
	"order-processing/external/tickets"
	"order-processing/storage"
	transportrabbit "order-processing/transport_rabbit"
	"order-processing/util"
	"time"

	"github.com/google/uuid"
	logger "github.com/sirupsen/logrus"
)

type OrderService struct {
	orderStore  *storage.OrderManager
	ticketStore storage.TicketStorage
	oInfoSender transportrabbit.AmqpSender
}

func NewMarketOrderService(store *storage.OrderManager, ticketStore storage.TicketStorage) OrderService {
	return OrderService{orderStore: store, ticketStore: ticketStore}
}

func (s *OrderService) CreateOrder(ctx context.Context, request *orders.CreateOrderRequest) error {
	logger.Infoln("Received request: ", request.String())

	oid := uuid.NewString()
	orderData := storage.OrderInfo{
		Id:             oid,
		CurrencyPair:   request.CurrencyPair,
		Direction:      int(request.GetDirection()),
		InitPrice:      float64(request.GetInitPrice()),
		InitVolume:     float64(request.GetInitVolume()),
		ExchangeWallet: request.GetExchangeWallet(),
		CreationDate:   uint64(time.Now().UTC().UnixMilli()),
		ExpirationDate: uint64(time.Now().Add(24 * time.Hour).UTC().UnixMilli()),
		OrderState:     int(orders.OrderState_ORDER_STATE_NEW),
		OrderType:      int(request.GetOrderType()),
	}

	err := s.orderStore.InsertNewOrder(ctx, orderData)
	if err != nil {
		return err
	}
	response := &orders.CreateOrderResponse{
		Id:      request.Id,
		OrderId: oid,
	}
	go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_CREATE_ORDER_RESPONSE, response)
	go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_ORDER_INFO, util.ConvertOrderModelToProto(&orderData))
	go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_LOCK_BALANCE, util.GetLockBalanceRequest(request, oid))
	return nil
}

func (s *OrderService) ApproveOrder(ctx context.Context, response *balances.LockBalanceResponse) error {
	data, err := s.orderStore.GetOrderById(ctx, response.GetId())
	if err != nil {
		return err
	}
	data.OrderState = int(orders.OrderState_ORDER_STATE_IN_PROCESS)
	if response.State == balances.LockBalanceStatus_REJECTED {
		data.OrderState = int(orders.OrderState_ORDER_STATE_REJECT)
		if err = s.orderStore.UpdateOrderData(ctx, *data); err != nil {
			return err
		}
		return util.ConvertBalanceError(response.ErrorMessage.ErrorCode)
	}
	if err = s.orderStore.UpdateOrderData(ctx, *data); err != nil {
		return err
	}
	go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_ORDER_INFO, util.ConvertOrderModelToProto(data))

	event := &orders.MatchOrderRequest{Id: data.Id}
	s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_MATCH_ORDER, event)
	return nil
}

func (s *OrderService) ReCreateOrder(ctx context.Context, orderInfo storage.OrderInfo) error {
	id := uuid.NewString()
	volume := orderInfo.InitVolume - orderInfo.FillVolume
	order := storage.OrderInfo{
		Id:             id,
		CurrencyPair:   orderInfo.CurrencyPair,
		Direction:      orderInfo.Direction,
		InitPrice:      orderInfo.InitPrice,
		InitVolume:     volume,
		ExchangeWallet: orderInfo.ExchangeWallet,
		CreationDate:   uint64(time.Now().UTC().UnixMilli()),
		ExpirationDate: uint64(time.Now().Add(24 * time.Hour).UTC().UnixMilli()),
		OrderState:     int(orders.OrderState_ORDER_STATE_IN_PROCESS),
		OrderType:      int(orders.OrderType_ORDER_TYPE_LIMIT),
	}
	if err := s.orderStore.InsertNewOrder(ctx, order); err != nil {
		return err
	}

	if err := s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_MATCH_ORDER, &orders.MatchOrderRequest{Id: id}); err != nil {
		return err
	}
	return nil
}

func (s *OrderService) GetOrder(ctx context.Context, request *orders.GetOrderRequest) (*storage.OrderInfo, error) {
	data, err := s.orderStore.GetOrderById(ctx, request.GetOrderId())
	return data, err

}

func (s *OrderService) DeleteOrder(ctx context.Context, request *orders.DeleteOrderRequest) error {
	response := orders.DeleteOrderResponse{
		Id: request.Id,
	}
	orderInfo, err := s.orderStore.GetOrderById(ctx, request.OrderId)
	if err != nil {
		response.ErrorMessage = &orders.OrderErrorMessage{ErorCode: orders.OrdersErrorCodes_ORDERS_ERROR_CODE_ORDER_NOT_FOUND, Message: "NotFound"}
		go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_DROP_ORDER_RESPONSE, &response)
		return err
	}
	if orderInfo.OrderState != int(orders.OrderState_ORDER_STATE_NEW) || orderInfo.OrderState != int(orders.OrderState_ORDER_STATE_IN_PROCESS) {
		response.ErrorMessage = &orders.OrderErrorMessage{ErorCode: orders.OrdersErrorCodes_ORDERS_ERROR_CODE_ORDER_NOT_FOUND, Message: "InvalidState"}
		go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_DROP_ORDER_RESPONSE, &response)
		return err
	}
	if err = s.orderStore.TryLockOrder(ctx, request.OrderId); err != nil {
		response.ErrorMessage = &orders.OrderErrorMessage{ErorCode: orders.OrdersErrorCodes_ORDERS_ERROR_CODE_ORDER_NOT_FOUND, Message: "InvalidState"}
		go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_DROP_ORDER_RESPONSE, &response)
		return err
	}
	go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_UNLOCK_BALANCE, util.GetUnLockBalanceRequest(orderInfo))
	orderInfo.OrderState = int(orders.OrderState_ORDER_STATE_REJECT)
	s.orderStore.UpdateOrderData(ctx, *orderInfo)
	go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_ORDER_INFO, util.ConvertOrderModelToProto(orderInfo))
	go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_DROP_ORDER_RESPONSE, &response)
	return nil
}
