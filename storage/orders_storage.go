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
		return errors.New("Internal")
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

func (o *OrderManager) DeleteOrderById(ctx context.Context, id string) error {
	o.mtx.Lock()
	defer o.mtx.Unlock()
	orderInfo, err := o.redisCli.GetFromHash(ctx, OrdersHash, id)
	if err != nil {
		return err
	}
	var orderModel OrderInfo
	err = json.Unmarshal([]byte(*orderInfo), orderModel)
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
		keys:    []string{*priceFilter, OrdersCreation, OrdersCurrencySets + orderInfo.CurrencyPair, OrdersDirectionSets + fmt.Sprintf("%d", orderInfo.Direction)},
		weights: []float64{float64(time.Now().Unix() * 100), 1, 0, 0},
	}, id)
	defer o.redisCli.client.Del(ctx, candidatesSet)
	matchOrderIds, err := o.redisCli.ZRange(ctx, candidatesSet, -1)
	logger.Infoln("Candidates for matching: ", strings.Join(matchOrderIds, ", "))
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
