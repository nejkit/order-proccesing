package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"order-processing/external/orders"
	"order-processing/statics"
	"strings"
	"sync"
	"time"

	logger "github.com/sirupsen/logrus"
)

var (
	OrdersHash              = "orders"
	OrdersPrice             = "orders:price"
	OrdersCreation          = "orders:creation"
	OrdersCurrencySets      = "orders:currency:"
	OrdersDirectionSets     = "orders:direction:"
	OrdersMatchingCandidate = "orders:match:"
	OrdersAvailable         = "orders:available"
	Transfers               = "transfers:"
	MatchingCandidates      = "orders:candidates:"
	OrderPriceSortDirection = map[int]int{
		int(orders.Direction_DIRECTION_TYPE_BUY.Number()):  1,
		int(orders.Direction_DIRECTION_TYPE_SELL.Number()): -1,
	}
)

type MatchingData struct {
	FillVolume float64
	FillPrice  float64
	Date       uint64
	State      orders.MatchState
	TransferId string
}

type OrderInfo struct {
	Id string

	CurrencyPair   string
	Direction      int
	InitPrice      float64
	MatchInfo      []MatchingData
	InitVolume     float64
	ExchangeWallet string

	CreationDate   uint64
	UpdatedDate    uint64
	ExpirationDate uint64

	OrderState int
	OrderType  int
}

type OrderManager struct {
	redisCli RedisClient
	mtx      sync.Mutex
}

func NewOrderManager(address string) OrderManager {
	redisCli := GetNewRedisCli(address)
	return OrderManager{redisCli: redisCli, mtx: sync.Mutex{}}
}

func (o *OrderManager) InsertNewOrder(ctx context.Context, request OrderInfo) error {
	value, err := json.Marshal(request)
	if err != nil {
		return err
	}
	err = o.redisCli.InsertHash(ctx, OrdersHash, request.Id, value)
	if err != nil {
		return err
	}
	err = o.redisCli.InsertZadd(ctx, OrdersPrice, request.Id, request.InitPrice)
	if err != nil {
		return err
	}
	err = o.redisCli.InsertZadd(ctx, OrdersCreation, request.Id, float64(request.CreationDate))
	if err != nil {
		return err
	}
	err = o.redisCli.InsertSet(ctx, OrdersCurrencySets+request.CurrencyPair, request.Id)
	if err != nil {
		return err
	}
	if orders.Direction(request.Direction) == orders.Direction_DIRECTION_TYPE_BUY {
		err = o.redisCli.InsertSet(ctx, OrdersDirectionSets+fmt.Sprintf("%d", orders.Direction_DIRECTION_TYPE_SELL.Number()), request.Id)
	} else {
		err = o.redisCli.InsertSet(ctx, OrdersDirectionSets+fmt.Sprintf("%d", orders.Direction_DIRECTION_TYPE_BUY.Number()), request.Id)
	}
	if err != nil {
		return err
	}
	err = o.redisCli.InsertSet(ctx, OrdersAvailable, request.Id)

	return nil
}

func (o *OrderManager) GetOrderById(ctx context.Context, id string) (*OrderInfo, error) {
	result, err := o.redisCli.GetFromHash(ctx, OrdersHash, id)
	if err != nil {
		return nil, err
	}
	if result != nil {
		var dataModel OrderInfo
		err = json.Unmarshal([]byte(*result), &dataModel)
		if err != nil {
			logger.Errorln("Err parse the value from redis! message: ", err.Error())
			return nil, err
		}
		return &dataModel, nil
	}
	return nil, err
}

func (o *OrderManager) UpdateOrderState(orderInfo OrderInfo) OrderInfo {
	switch orderInfo.OrderState {
	case int(orders.OrderState_ORDER_STATE_NEW):
		orderInfo.OrderState = int(orders.OrderState_ORDER_STATE_IN_PROCESS)
	case int(orders.OrderState_ORDER_STATE_IN_PROCESS):
		orderInfo.OrderState = o.fillingStateOrder(orderInfo)
	}
	return orderInfo
}

func (o *OrderManager) fillingStateOrder(orderInfo OrderInfo) int {
	availableVolume := CalculateAvailableVolume(orderInfo)
	if availableVolume == 0 {
		return int(orders.OrderState_ORDER_STATE_FILL)
	}
	if availableVolume != orderInfo.InitVolume {
		return int(orders.OrderState_ORDER_STATE_PART_FILL)
	}

	return int(orders.OrderState_ORDER_STATE_IN_PROCESS)
}

func ApproveOrder(orderInfo *OrderInfo) {
	txApproved := true
	for _, matchInfo := range orderInfo.MatchInfo {
		if matchInfo.State != orders.MatchState_MATCH_STATE_REJECT && matchInfo.State != orders.MatchState_MATCH_STATE_DONE {
			txApproved = false
		}
	}

	if txApproved {
		orderInfo.OrderState = int(orders.OrderState_ORDER_STATE_DONE)
	}
}

func (o *OrderManager) UpdateOrderData(ctx context.Context, orderInfo OrderInfo) error {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	if orderInfo.OrderState == int(orders.OrderState_ORDER_STATE_FILL) {
		o.redisCli.DeleteFromSet(ctx, OrdersAvailable, orderInfo.Id)
	}

	orderInfo.UpdatedDate = uint64(time.Now().Unix())

	data, err := json.Marshal(orderInfo)
	if err != nil {
		return err
	}
	err = o.redisCli.InsertHash(ctx, OrdersHash, orderInfo.Id, data)
	if err != nil {
		return err
	}

	return nil
}

func (o *OrderManager) DeleteOrderById(ctx context.Context, id string) error {
	o.mtx.Lock()
	defer o.mtx.Unlock()
	orderInfo, err := o.redisCli.GetFromHash(ctx, OrdersHash, id)
	if err != nil {
		return err
	}
	var orderModel OrderInfo
	err = json.Unmarshal([]byte(*orderInfo), &orderModel)
	if err != nil {
		return errors.New(statics.InternalError)
	}
	o.redisCli.DeleteFromHash(ctx, OrdersHash, id)
	o.redisCli.DeleteFromZAdd(ctx, OrdersPrice, id)
	o.redisCli.DeleteFromZAdd(ctx, OrdersPrice, id)
	o.redisCli.DeleteFromSet(ctx, OrdersCurrencySets+orderModel.CurrencyPair, id)
	if orderModel.Direction == int(orders.Direction_DIRECTION_TYPE_BUY) {
		o.redisCli.DeleteFromSet(ctx, OrdersDirectionSets+fmt.Sprintf("%d", orders.Direction_DIRECTION_TYPE_SELL), id)
		return nil
	}
	o.redisCli.DeleteFromSet(ctx, OrdersDirectionSets+fmt.Sprintf("%d", orders.Direction_DIRECTION_TYPE_BUY), id)
	o.redisCli.DeleteFromSet(ctx, OrdersAvailable, id)
	return nil
}

func (o *OrderManager) GetOrderIdsForMatching(ctx context.Context, id string) ([]string, error) {
	o.mtx.Lock()
	defer o.mtx.Unlock()
	info, err := o.redisCli.GetFromHash(ctx, OrdersHash, id)
	var orderInfo OrderInfo
	err = json.Unmarshal([]byte(*info), &orderInfo)
	if err != nil {
		return nil, err
	}
	priceFilter, err := o.getPriceFilterForLimitOrderByInfo(ctx, orderInfo)
	if err != nil {
		return nil, err
	}
	candidatesSet := o.redisCli.ZInterStorage(ctx, ZInterOptions{
		prefix:  MatchingCandidates,
		keys:    []string{*priceFilter, OrdersCreation, OrdersCurrencySets + orderInfo.CurrencyPair, OrdersDirectionSets + fmt.Sprintf("%d", orderInfo.Direction), OrdersAvailable},
		weights: []float64{float64(time.Now().Unix()*100) * float64(OrderPriceSortDirection[orderInfo.Direction]), 1, 0, 0, 0},
	}, id)
	defer o.redisCli.client.Del(ctx, candidatesSet)
	matchOrderIds, err := o.redisCli.ZRange(ctx, candidatesSet, -1)
	logger.Infoln("Candidates for matching: ", strings.Join(matchOrderIds, ", "))
	if matchOrderIds == nil {
		return nil, errors.New(statics.ErrorOrderNotFound)
	}
	return matchOrderIds, nil
}

func (o *OrderManager) getPriceFilterForLimitOrderByInfo(ctx context.Context, oInfo OrderInfo) (*string, error) {
	if orders.Direction_DIRECTION_TYPE_BUY == orders.Direction(oInfo.Direction) {
		return o.redisCli.PrepareIndexWithLimitOption(ctx, LimitOptions{
			maxPrice: oInfo.InitPrice,
			minPrice: 0,
		})
	}
	return o.redisCli.PrepareIndexWithLimitOption(ctx, LimitOptions{
		maxPrice: 0,
		minPrice: oInfo.InitPrice,
	})
}

func CalculateAvailableVolume(orderInfo OrderInfo) float64 {
	if orders.Direction(orderInfo.Direction) == orders.Direction_DIRECTION_TYPE_SELL {
		fillVolume := calculateFillVolumeForSellOrder(orderInfo.MatchInfo)
		return orderInfo.InitVolume - fillVolume
	}
	fillVolume := calculateFillVolumeForBuyOrder(orderInfo.MatchInfo)
	return orderInfo.InitVolume*orderInfo.InitPrice - fillVolume

}

func calculateFillVolumeForBuyOrder(matchingInfo []MatchingData) float64 {
	if matchingInfo == nil {
		return 0.0
	}
	fillVolume := 0.0
	for _, matchInfo := range matchingInfo {
		fillVolume += matchInfo.FillPrice * matchInfo.FillVolume
	}
	return fillVolume
}

func calculateFillVolumeForSellOrder(matchingInfo []MatchingData) float64 {
	if matchingInfo == nil {
		return 0.0
	}
	fillVolume := 0.0
	for _, matchInfo := range matchingInfo {
		if matchInfo.State != orders.MatchState_MATCH_STATE_REJECT {
			fillVolume += matchInfo.FillVolume
		}
	}
	return fillVolume

}

func (m *OrderManager) GetOrdersByTransferId(ctx context.Context, transferId string) ([]OrderInfo, error) {
	ids, err := m.redisCli.GetFromSet(ctx, Transfers)
	if err != nil {
		return nil, errors.New(statics.ErrorOrderNotFound)
	}
	var result []OrderInfo
	for _, id := range ids {
		order, err := m.GetOrderById(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, *order)
	}
	return result, nil
}

func GetMatchingDataFromOrderByTransferId(orderInfo OrderInfo, transferId string) (*MatchingData, error) {
	for _, matchInfo := range orderInfo.MatchInfo {
		if matchInfo.TransferId == transferId {
			return &matchInfo, nil
		}
	}
	return nil, errors.New(statics.ErrorOrderNotFound)
}

func (m *OrderManager) DeleteTransferInfo(ctx context.Context, transferId string) {
	m.redisCli.DeleteFromSet(ctx, Transfers, transferId)
}

func (m *OrderManager) AddAvailabilityForOrder(ctx context.Context, order OrderInfo) {
	if order.OrderType == int(orders.OrderType_ORDER_TYPE_LIMIT) {
		m.redisCli.InsertSet(ctx, OrdersAvailable, order.Id)
	}
}

func (m *OrderManager) AddTransferData(ctx context.Context, transferId string, firstOrder string, secondOrder string) error {
	err := m.redisCli.InsertSet(ctx, Transfers+transferId, firstOrder)
	if err != nil {
		return err
	}
	err = m.redisCli.InsertSet(ctx, Transfers+transferId, secondOrder)
	if err != nil {
		return err
	}
	return nil
}
