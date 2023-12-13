package services

import (
	"context"
	"errors"
	"order-processing/external/balances"
	"order-processing/external/orders"
	"order-processing/external/tickets"
	"order-processing/statics"
	"order-processing/storage"
	transportrabbit "order-processing/transport_rabbit"
	"order-processing/util"
	"strings"
	"time"

	"github.com/google/uuid"
	logger "github.com/sirupsen/logrus"
)

var (
	ErrorNotFound = errors.New(statics.ErrorOrderNotFound)
)

type MatcherService struct {
	store        *storage.OrderManager
	ticketStore  storage.TicketStorage
	unlockSender transportrabbit.AmqpSender
}

func NewMatcherService(om *storage.OrderManager, tsender transportrabbit.AmqpSender, ticketStore storage.TicketStorage) MatcherService {
	return MatcherService{
		store:       om,
		ticketStore: ticketStore,
	}
}

func (s *MatcherService) MatchOrderById(ctx context.Context, id string) error {
	orderInfo, err := s.store.GetOrderById(ctx, id)
	if err != nil {
		logger.Errorln(err.Error())
		return err
	}
	listOrdersForMatching, err := s.store.GetOrderIdsForMatching(ctx, id)
	logger.Infoln("Orders for matching: ", strings.Join(listOrdersForMatching, ", "))
	if err != nil {
		logger.Errorln(err.Error())
		return err
	}

	for _, id := range listOrdersForMatching {
		if err = s.tryMatchOrder(ctx, id, orderInfo); err != nil {
			continue
		}
		return nil
	}
	if orderInfo.OrderType == int(orders.OrderType_ORDER_TYPE_MARKET) {
		go s.rejectOrder(ctx, *orderInfo)
		return nil
	}
	s.store.AddLimitOrderToStockBook(ctx, *orderInfo)
	return nil
}

func (s *MatcherService) HandleTransfersResponse(ctx context.Context, transfer *balances.Transfer) error {
	switch transfer.GetState() {
	case balances.TransferState_TRANSFER_STATE_NEW:
		{
			logger.Infoln("Receive transfer in state New. Id: ", transfer.GetId())
		}
	case balances.TransferState_TRANSFER_STATE_IN_PROGRESS:
		{
			logger.Infoln("Receive transfer in state In Progress. Id: ", transfer.GetId())

		}
	case balances.TransferState_TRANSFER_STATE_REJECT:
		{
			s.catchRejectTransfer(ctx, transfer)
		}
	case balances.TransferState_TRANSFER_STATE_DONE:
		{
			s.catchDoneTransfer(ctx, transfer)
		}
	}
	return nil
}

func buildMatchingInfo(price float64, buyVolume float64, transferId string, orderInfo *storage.OrderInfo) {
	orderInfo.TransferId = transferId
	orderInfo.MatchingDate = uint64(time.Now().UTC().UnixMilli())
	orderInfo.FillPrice = price
	orderInfo.FillVolume = buyVolume / price

}

func (s *MatcherService) sendTransferRequest(ctx context.Context, transferId string, firstOrder storage.OrderInfo, secondOrder storage.OrderInfo) {
	event := balances.CreateTransferRequest{
		Id: transferId,
		SenderData: &balances.TransferOptions{
			Address:  firstOrder.ExchangeWallet,
			Currency: util.ParseCurrencyFromDirection(firstOrder.CurrencyPair, firstOrder.Direction),
			Amount:   util.CalculateAmountInTransfer(firstOrder),
		},
		RecepientData: &balances.TransferOptions{
			Address:  secondOrder.ExchangeWallet,
			Currency: util.ParseCurrencyFromDirection(secondOrder.CurrencyPair, secondOrder.Direction),
			Amount:   util.CalculateAmountInTransfer(secondOrder),
		}}
	s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_CREATE_TRANSFER, &event)
}

func (s *MatcherService) tryLockOrderInfo(ctx context.Context, orderInfo storage.OrderInfo) bool {
	logger.Infoln("Start attempt to block resource")
	if err := s.store.TryLockOrder(ctx, orderInfo.Id); err != nil {
		logger.Errorln(err.Error())
		return false
	}
	return true
}

func (s *MatcherService) matchOrders(ctx context.Context, fillVolumeBid float64, orderInfo *storage.OrderInfo, candidateMatchingInfo *storage.OrderInfo) string {
	transferId := uuid.NewString()
	buildMatchingInfo(candidateMatchingInfo.InitPrice, fillVolumeBid, transferId, orderInfo)
	buildMatchingInfo(candidateMatchingInfo.InitPrice, fillVolumeBid, transferId, candidateMatchingInfo)
	s.store.UpdateOrderState(orderInfo)
	s.store.UpdateOrderState(candidateMatchingInfo)
	if err := s.store.UpdateOrderData(ctx, *orderInfo); err != nil {
		logger.Infoln(err.Error())
	}
	go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_ORDER_INFO, util.ConvertOrderModelToProto(orderInfo))
	if err := s.store.UpdateOrderData(ctx, *candidateMatchingInfo); err != nil {
		logger.Infoln(err.Error())
	}
	go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_ORDER_INFO, util.ConvertOrderModelToProto(candidateMatchingInfo))
	return transferId
}

func (s *MatcherService) rejectOrder(ctx context.Context, orderInfo storage.OrderInfo) {
	go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_UNLOCK_BALANCE, util.GetUnLockBalanceRequest(&orderInfo))
	orderInfo.OrderState = int(orders.OrderState_ORDER_STATE_REJECT)
	s.store.UpdateOrderData(ctx, orderInfo)
	go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_ORDER_INFO, util.ConvertOrderModelToProto(&orderInfo))
}

func (s *MatcherService) calculateAmountMatching(orderInfo storage.OrderInfo) float64 {
	return orderInfo.InitPrice * orderInfo.InitVolume
}

func (s *MatcherService) catchRejectTransfer(ctx context.Context, transfer *balances.Transfer) {
	logger.Infoln("Receive transfer in state Reject. Id: ", transfer.GetId(), " Error: ", transfer.GetError().String())
	orderInfo, _ := s.store.GetOrdersByTransferId(ctx, transfer.GetId())

	for _, order := range orderInfo {
		if err := s.tryLockOrderInfo(ctx, order); err == false {
			continue
		}
		order.FillPrice = 0
		order.FillVolume = 0
		s.rejectOrder(ctx, order)
		s.store.TryUnlockOrder(ctx, order.Id)
	}
}

func (s *MatcherService) catchDoneTransfer(ctx context.Context, transfer *balances.Transfer) {
	logger.Infoln("Receive transfer in state Done. Id: ", transfer.GetId(), " Try complete matching...")
	orderInfo, _ := s.store.GetOrdersByTransferId(ctx, transfer.GetId())

	for _, order := range orderInfo {
		logger.Infoln("Complete order with id: ", order.Id)
		if err := s.tryLockOrderInfo(ctx, order); err == false {
			continue
		}
		if order.OrderState == int(orders.OrderState_ORDER_STATE_PART_FILL) {
			go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_RECREATE_ORDER, util.ConvertOrderModelToProto(&order))
		}
		if order.OrderState == int(orders.OrderState_ORDER_STATE_FILL) {
			go s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_UNLOCK_BALANCE, util.GetUnLockBalanceRequest(&order))
		}

		order.OrderState = int(orders.OrderState_ORDER_STATE_DONE)
		s.store.UpdateOrderData(ctx, order)
		s.ticketStore.SaveTicketForOperation(ctx, tickets.OperationType_OPERATION_TYPE_ORDER_INFO, util.ConvertOrderModelToProto(&order))
		s.store.TryUnlockOrder(ctx, order.Id)
	}
}

func (s *MatcherService) tryMatchOrder(ctx context.Context, id string, rootOrder *storage.OrderInfo) error {
	candidateMatchingInfo, err := s.store.GetOrderById(ctx, id)
	if err != nil {
		return err
	}
	if lock := s.tryLockOrderInfo(ctx, *candidateMatchingInfo); lock == false {
		return errors.New(statics.ErrorResourceLocked)
	}

	if candidateMatchingInfo.ExpirationDate < uint64(time.Now().UTC().UnixMilli()) {
		s.rejectOrder(ctx, *candidateMatchingInfo)
		return errors.New("OrderExpired")
	}
	availableAmountRootOrder := s.calculateAmountMatching(*rootOrder)

	availableAmountCandidateOrder := s.calculateAmountMatching(*candidateMatchingInfo)
	fillVolumeBid := util.Min(availableAmountRootOrder, availableAmountCandidateOrder)
	transferId := s.matchOrders(ctx, fillVolumeBid, rootOrder, candidateMatchingInfo)
	s.store.AddTransferData(ctx, transferId, rootOrder.Id, candidateMatchingInfo.Id)
	s.store.DropLimitOrderToStockBook(ctx, *candidateMatchingInfo)
	s.store.TryUnlockOrder(ctx, candidateMatchingInfo.Id)
	go s.sendTransferRequest(ctx, transferId, *rootOrder, *candidateMatchingInfo)
	return nil
}
