package services

import (
	"context"
	"errors"
	"order-processing/external/balances"
	"order-processing/external/orders"
	"order-processing/statics"
	"order-processing/storage"
	transportrabbit "order-processing/transport_rabbit"
	"order-processing/util"
	"sync"
	"time"

	"github.com/google/uuid"
	logger "github.com/sirupsen/logrus"
)

var (
	ErrorNotFound = errors.New(statics.ErrorOrderNotFound)
)

type MatcherService struct {
	store           *storage.OrderManager
	transferSender  transportrabbit.AmqpSender
	transferStorage transportrabbit.AmqpStorage[balances.Transfer]
	mtx             sync.Mutex
}

func NewMatcherService(om *storage.OrderManager, tsender transportrabbit.AmqpSender, tstorage transportrabbit.AmqpStorage[balances.Transfer]) MatcherService {
	return MatcherService{
		store:           om,
		transferSender:  tsender,
		transferStorage: tstorage,
		mtx:             sync.Mutex{},
	}
}

func (s *MatcherService) MatchOrderById(ctx context.Context, id string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	orderInfo, err := s.store.GetOrderById(ctx, id)
	if err != nil {
		return err
	}

	listOrdersForMatching, err := s.store.GetOrderIdsForMatching(ctx, id)
	if err == ErrorNotFound {
		if orderInfo.OrderType == int(orders.OrderType_ORDER_TYPE_MARKET) {
			orderInfo.OrderState = int(orders.OrderState_ORDER_STATE_REJECT)
			s.store.UpdateOrderData(ctx, *orderInfo)
		}
		return nil
	}
	if err != nil {
		return err
	}

	orderInfo.OrderState = int(orders.OrderState_ORDER_STATE_IN_PROCESS)
	s.store.UpdateOrderData(ctx, *orderInfo)

	for _, id := range listOrdersForMatching {
		candidateMatchingInfo, err := s.store.GetOrderById(ctx, id)
		if err != nil {
			continue
		}
		availableAmountRootOrder := storage.CalculateAvailableVolume(*orderInfo)
		if availableAmountRootOrder == 0 {
			s.store.UpdateOrderData(ctx, *orderInfo)
			return nil
		}
		if candidateMatchingInfo.OrderState == int(orders.OrderState_ORDER_STATE_NEW) {
			candidateMatchingInfo.OrderState = int(orders.OrderState_ORDER_STATE_IN_PROCESS)
			s.store.UpdateOrderData(ctx, *candidateMatchingInfo)
		}

		availableAmountCandidateOrder := storage.CalculateAvailableVolume(*candidateMatchingInfo)
		fillVolumeBid := util.Min(availableAmountRootOrder, availableAmountCandidateOrder)
		availableAmountRootOrder -= fillVolumeBid
		transferId := uuid.NewString()
		matchingDataRootOrder := buildMatchingInfo(candidateMatchingInfo.InitPrice, fillVolumeBid, transferId, orderInfo.Direction)
		matchingDataCandidateOrder := buildMatchingInfo(candidateMatchingInfo.InitPrice, fillVolumeBid, transferId, candidateMatchingInfo.Direction)
		orderInfo.MatchInfo = append(orderInfo.MatchInfo, matchingDataRootOrder)
		candidateMatchingInfo.MatchInfo = append(candidateMatchingInfo.MatchInfo, matchingDataCandidateOrder)
		s.store.UpdateOrderData(ctx, *orderInfo)
		s.store.UpdateOrderData(ctx, *candidateMatchingInfo)
		s.store.AddTransferData(ctx, transferId, orderInfo.Id, candidateMatchingInfo.Id)
		go s.sendTransferRequest(ctx, transferId, *orderInfo, *candidateMatchingInfo)
	}

	return nil
}

func (s *MatcherService) HandleTransfersResponse(ctx context.Context, transfer *balances.Transfer) {
	switch transfer.GetState() {
	case balances.TransferState_TRANSFER_STATE_NEW:
		{
			logger.Infoln("Receive transfer in state New. Id: ", transfer.GetId())
		}
	case balances.TransferState_TRANSFER_STATE_IN_PROGRESS:
		{
			logger.Infoln("Receive transfer in state In Progress. Id: ", transfer.GetId(), " Try update matching order...")
			orderInfo, _ := s.store.GetOrdersByTransferId(ctx, transfer.GetId())

			for _, order := range orderInfo {
				s.updateStateMatching(ctx, transfer.GetId(), order, orders.MatchState_MATCH_STATE_IN_PROGRESS)
				go s.store.UpdateOrderData(ctx, order)
			}

		}
	case balances.TransferState_TRANSFER_STATE_REJECT:
		{
			logger.Infoln("Receive transfer in state Reject. Id: ", transfer.GetId(), " Error: ", transfer.GetError().String())
			orderInfo, _ := s.store.GetOrdersByTransferId(ctx, transfer.GetId())

			for _, order := range orderInfo {
				s.updateStateMatching(ctx, transfer.GetId(), order, orders.MatchState_MATCH_STATE_REJECT)
				go s.store.UpdateOrderData(ctx, order)
				go s.store.AddAvailabilityForOrder(ctx, order)
			}
			go s.store.DeleteTransferInfo(ctx, transfer.GetId())

		}
	case balances.TransferState_TRANSFER_STATE_DONE:
		{
			logger.Infoln("Receive transfer in state Done. Id: ", transfer.GetId(), " Try complete matching...")
			orderInfo, _ := s.store.GetOrdersByTransferId(ctx, transfer.GetId())

			for _, order := range orderInfo {
				s.updateStateMatching(ctx, transfer.GetId(), order, orders.MatchState_MATCH_STATE_DONE)
				storage.ApproveOrder(&order)
				go s.store.UpdateOrderData(ctx, order)
			}
			go s.store.DeleteTransferInfo(ctx, transfer.GetId())
		}
	}
}

func buildMatchingInfo(price float64, buyVolume float64, transferId string, direction int) storage.MatchingData {
	baseMatchingData := storage.MatchingData{
		TransferId: transferId,
		State:      orders.MatchState_MATCH_STATE_NEW,
		Date:       uint64(time.Now().Unix()),
		FillPrice:  price,
	}
	if orders.Direction(direction) == orders.Direction_DIRECTION_TYPE_BUY {
		baseMatchingData.FillVolume = buyVolume
		return baseMatchingData
	}
	baseMatchingData.FillVolume = buyVolume / price
	return baseMatchingData
}

func (s *MatcherService) sendTransferRequest(ctx context.Context, transferId string, firstOrder storage.OrderInfo, secondOrder storage.OrderInfo) {
	matchInfoFirst, _ := storage.GetMatchingDataFromOrderByTransferId(firstOrder, transferId)
	matchInfoSecond, _ := storage.GetMatchingDataFromOrderByTransferId(secondOrder, transferId)
	event := balances.CreateTransferRequest{
		Id: transferId,
		SenderData: &balances.TransferOptions{
			Address:  firstOrder.ExchangeWallet,
			Currency: util.ParseCurrencyFromDirection(firstOrder.CurrencyPair, firstOrder.Direction),
			Amount:   float32(matchInfoFirst.FillVolume),
		},
		RecepientData: &balances.TransferOptions{
			Address:  secondOrder.ExchangeWallet,
			Currency: util.ParseCurrencyFromDirection(secondOrder.CurrencyPair, secondOrder.Direction),
			Amount:   float32(matchInfoSecond.FillVolume),
		}}
	s.transferSender.SendMessage(ctx, &event)

}

func (s *MatcherService) updateStateMatching(ctx context.Context, transferId string, orderInfo storage.OrderInfo, state orders.MatchState) {
	for _, matchInfo := range orderInfo.MatchInfo {
		if matchInfo.TransferId == transferId {
			matchInfo.State = state
			break
		}
	}

}
