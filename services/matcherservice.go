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
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	logger "github.com/sirupsen/logrus"
)

var (
	ErrorNotFound = errors.New(statics.ErrorOrderNotFound)
)

type MatcherService struct {
	store          *storage.OrderManager
	transferSender transportrabbit.AmqpSender
	mtx            sync.Mutex
}

func NewMatcherService(om *storage.OrderManager, tsender transportrabbit.AmqpSender) MatcherService {
	return MatcherService{
		store:          om,
		transferSender: tsender,
		mtx:            sync.Mutex{},
	}
}

func (s *MatcherService) MatchOrderById(ctx context.Context, id string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	orderInfo, err := s.store.GetOrderById(ctx, id)
	if err != nil {
		logger.Errorln(err.Error())
		return err
	}

	listOrdersForMatching, err := s.store.GetOrderIdsForMatching(ctx, id)
	logger.Infoln("Orders for matching: ", strings.Join(listOrdersForMatching, ", "))
	if err != nil && err.Error() == statics.ErrorOrderNotFound {
		if orderInfo.OrderType == int(orders.OrderType_ORDER_TYPE_MARKET) {
			orderInfo.OrderState = int(orders.OrderState_ORDER_STATE_REJECT)
			err := s.store.UpdateOrderData(ctx, *orderInfo)
			logger.Errorln("Order was rejected")
			if err != nil {
				logger.Errorln(err.Error())
			}
		}
		return nil
	}
	if err != nil {
		logger.Errorln(err.Error())
		return err
	}

	s.store.UpdateOrderState(orderInfo)
	s.store.UpdateOrderData(ctx, *orderInfo)

	for _, id := range listOrdersForMatching {
		candidateMatchingInfo, err := s.store.GetOrderById(ctx, id)
		if err != nil {
			continue
		}
		availableAmountRootOrder := storage.CalculateAvailableVolume(*orderInfo)
		if availableAmountRootOrder == 0 {
			break
		}

		s.store.UpdateOrderState(candidateMatchingInfo)
		s.store.UpdateOrderData(ctx, *candidateMatchingInfo)

		availableAmountCandidateOrder := storage.CalculateAvailableVolume(*candidateMatchingInfo)
		fillVolumeBid := util.Min(availableAmountRootOrder, availableAmountCandidateOrder)
		availableAmountRootOrder -= fillVolumeBid
		transferId := uuid.NewString()
		matchingDataRootOrder := buildMatchingInfo(candidateMatchingInfo.InitPrice, fillVolumeBid, transferId, orderInfo.Direction)
		matchingDataCandidateOrder := buildMatchingInfo(candidateMatchingInfo.InitPrice, fillVolumeBid, transferId, candidateMatchingInfo.Direction)
		orderInfo.MatchInfo = append(orderInfo.MatchInfo, matchingDataRootOrder)
		candidateMatchingInfo.MatchInfo = append(candidateMatchingInfo.MatchInfo, matchingDataCandidateOrder)
		s.store.UpdateOrderState(orderInfo)
		s.store.UpdateOrderState(candidateMatchingInfo)
		if err = s.store.UpdateOrderData(ctx, *orderInfo); err != nil {
			logger.Infoln(err.Error())
		}
		if err = s.store.UpdateOrderData(ctx, *candidateMatchingInfo); err != nil {
			logger.Infoln(err.Error())
		}
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
				s.store.ApproveOrder(ctx, &order)
				s.store.UpdateOrderData(ctx, order)
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
		baseMatchingData.FillVolume = buyVolume / price
		return baseMatchingData
	}
	baseMatchingData.FillVolume = buyVolume
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
