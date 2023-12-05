package handlers

import (
	"context"
	"time"

	"order-processing/external/balances"
	"order-processing/external/orders"
	"order-processing/external/tickets"
	"order-processing/services"
	"order-processing/storage"
	transportrabbit "order-processing/transport_rabbit"

	logger "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type TicketHandler struct {
	ticketStore         storage.TicketStorage
	orderService        services.OrderService
	matchingService     *services.MatcherService
	balanceService      services.BalanceService
	orderCreationSender transportrabbit.AmqpSender
}

func NewTicketHandler(
	ticketStore storage.TicketStorage,
	orderService services.OrderService,
	matchingService *services.MatcherService,
	balanceService services.BalanceService,
	ocs transportrabbit.AmqpSender) TicketHandler {
	return TicketHandler{ticketStore: ticketStore, orderService: orderService, balanceService: balanceService, matchingService: matchingService, orderCreationSender: ocs}
}

func (h *TicketHandler) Handle(ctx context.Context) {
	for {
		ticketIds, err := h.ticketStore.GetTickets(ctx)
		if err != nil {
			logger.Errorln(err.Error())
		}
		for _, ticketId := range ticketIds {
			ticketInfo, err := h.ticketStore.GetTicketById(ctx, ticketId)
			if err != nil {
				logger.Errorln(err.Error())
				continue
			}
			if ticketInfo.State == tickets.TicketState_TICKET_STATE_PROCESSING || ticketInfo.State == tickets.TicketState_TICKET_STATE_DONE {
				continue
			}
			ticketInfo.State = tickets.TicketState_TICKET_STATE_PROCESSING
			if err = h.ticketStore.UpdateTicket(ctx, ticketInfo); err != nil {
				logger.Errorln(err.Error())
				continue
			}

			switch ticketInfo.OperationType {
			case tickets.OperationType_OPERATION_TYPE_CREATE_ORDER:
				request := &orders.CreateOrderRequest{}
				if err := proto.Unmarshal(ticketInfo.Data, request); err != nil {
					logger.Errorln(err.Error())
					continue
				}
				go h.orderService.CreateOrder(ctx, request)
			case tickets.OperationType_OPERATION_TYPE_LOCK_BALANCE:
				request := &balances.LockBalanceRequest{}
				if err := proto.Unmarshal(ticketInfo.Data, request); err != nil {
					logger.Errorln("Error while parse lock balance request: ", err.Error())
					continue
				}
				go h.balanceService.LockBalance(ctx, request)
			case tickets.OperationType_OPERATION_TYPE_APPROVE_CREATION:
				request := &balances.LockBalanceResponse{}
				if err := proto.Unmarshal(ticketInfo.Data, request); err != nil {
					logger.Errorln(err.Error())
					continue
				}
				go h.orderService.ApproveOrder(ctx, request)
			case tickets.OperationType_OPERATION_TYPE_MATCH_ORDER:
				request := &orders.MatchOrderRequest{}
				if err := proto.Unmarshal(ticketInfo.Data, request); err != nil {
					logger.Errorln(err.Error())
					continue
				}
				go h.matchingService.MatchOrderById(ctx, request.Id)
			case tickets.OperationType_OPERATION_TYPE_CREATE_TRANSFER:
				request := &balances.CreateTransferRequest{}
				if err := proto.Unmarshal(ticketInfo.Data, request); err != nil {
					logger.Errorln(err.Error())
					continue
				}
				go h.balanceService.CreateTransfer(ctx, request)
			case tickets.OperationType_OPERATION_TYPE_TRANSFER:
				request := &balances.Transfer{}
				if err := proto.Unmarshal(ticketInfo.Data, request); err != nil {
					logger.Errorln(err.Error())
					continue
				}
				go h.matchingService.HandleTransfersResponse(ctx, request)
			case tickets.OperationType_OPERATION_TYPE_CREATE_ORDER_RESPONSE:
				request := &orders.CreateOrderResponse{}
				if err := proto.Unmarshal(ticketInfo.Data, request); err != nil {
					logger.Errorln(err.Error())
					continue
				}
				go h.orderCreationSender.SendMessage(ctx, request)
			default:
				logger.Warningln("Ticket operation ", ticketInfo.OperationType, " unsupported, skipping...")
				continue
			}

			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}

		}
	}

}
