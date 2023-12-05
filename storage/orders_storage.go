package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"order-processing/external/orders"
	"order-processing/statics"
	"time"

	"github.com/google/uuid"
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
	FillVolume   float64
	FillPrice    float64
	MatchingDate uint64
	TransferId   string
}

type OrderInfo struct {
	Id string

	CurrencyPair string
	Direction    int
	InitPrice    float64
	MatchingData
	InitVolume     float64
	ExchangeWallet string

	CreationDate   uint64
	UpdatedDate    uint64
	ExpirationDate uint64

	OrderState int
	OrderType  int
}

type OrderManager struct {
	redisCli   RedisClient
	instanceId string
}

func NewOrderManager(address string) OrderManager {
	redisCli := GetNewRedisCli(address)
	return OrderManager{redisCli: redisCli, instanceId: uuid.NewString()}
}

func (o *OrderManager) InsertNewOrder(ctx context.Context, request OrderInfo) error {
	value, err := json.Marshal(request)
	if err != nil {
		return err
	}
	go o.redisCli.InsertHash(ctx, OrdersHash, request.Id, value)

	return nil
}

func (o *OrderManager) AddLimitOrderToStockBook(ctx context.Context, orderInfo OrderInfo) {
	go o.redisCli.InsertZadd(ctx, OrdersPrice, orderInfo.Id, orderInfo.InitPrice)

	go o.redisCli.InsertZadd(ctx, OrdersCreation, orderInfo.Id, float64(orderInfo.CreationDate))

	go o.redisCli.InsertSet(ctx, OrdersCurrencySets+orderInfo.CurrencyPair, orderInfo.Id)

	if orders.Direction(orderInfo.Direction) == orders.Direction_DIRECTION_TYPE_BUY {
		go o.redisCli.InsertSet(ctx, OrdersDirectionSets+fmt.Sprintf("%d", orders.Direction_DIRECTION_TYPE_SELL.Number()), orderInfo.Id)
	} else {
		go o.redisCli.InsertSet(ctx, OrdersDirectionSets+fmt.Sprintf("%d", orders.Direction_DIRECTION_TYPE_BUY.Number()), orderInfo.Id)
	}
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

func (o *OrderManager) UpdateOrderState(orderInfo *OrderInfo) {
	switch orderInfo.OrderState {
	case int(orders.OrderState_ORDER_STATE_IN_PROCESS):
		o.fillingStateOrder(orderInfo)
	case int(orders.OrderState_ORDER_STATE_PART_FILL):
		o.fillingStateOrder(orderInfo)
	}

}

func (o *OrderManager) fillingStateOrder(orderInfo *OrderInfo) {
	availableVolume := CalculateAvailableVolume(*orderInfo)
	if availableVolume == 0.0 {
		orderInfo.OrderState = int(orders.OrderState_ORDER_STATE_FILL)
		return
	}
	if availableVolume != orderInfo.InitVolume {
		orderInfo.OrderState = int(orders.OrderState_ORDER_STATE_PART_FILL)
		return
	}
}

func (o *OrderManager) UpdateOrderData(ctx context.Context, orderInfo OrderInfo) error {
	if orderInfo.OrderState == int(orders.OrderState_ORDER_STATE_FILL) {
		o.redisCli.DeleteFromSet(ctx, OrdersAvailable, orderInfo.Id)
	}

	orderInfo.UpdatedDate = uint64(time.Now().UTC().UnixMilli())

	data, err := json.Marshal(orderInfo)
	if err != nil {
		return err
	}
	logger.Infoln("Update order. Request: ", string(data))
	err = o.redisCli.InsertHash(ctx, OrdersHash, orderInfo.Id, data)
	if err != nil {
		return err
	}

	return nil
}

func (o *OrderManager) DeleteOrderById(ctx context.Context, id string) error {

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
	o.redisCli.DeleteFromZAdd(ctx, OrdersCreation, id)
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

	info, err := o.redisCli.GetFromHash(ctx, OrdersHash, id)
	var orderInfo OrderInfo
	err = json.Unmarshal([]byte(*info), &orderInfo)
	if err != nil {
		return nil, err
	}
	priceFilter := &OrdersPrice
	if orderInfo.OrderType == int(orders.OrderType_ORDER_TYPE_LIMIT) {
		priceFilter, err = o.getPriceFilterForLimitOrderByInfo(ctx, orderInfo)
		defer o.redisCli.DelKey(ctx, *priceFilter)
	}
	if err != nil {
		return nil, err
	}
	candidatesSet := o.redisCli.ZInterStorage(ctx, ZInterOptions{
		prefix:  MatchingCandidates,
		keys:    []string{*priceFilter, OrdersCreation, OrdersCurrencySets + orderInfo.CurrencyPair, OrdersDirectionSets + fmt.Sprintf("%d", orderInfo.Direction)},
		weights: []float64{float64(time.Now().Unix()*100) * float64(OrderPriceSortDirection[orderInfo.Direction]), 1, 0, 0},
	}, id)
	defer o.redisCli.client.Del(ctx, candidatesSet)
	matchOrderIds, err := o.redisCli.ZRange(ctx, candidatesSet, -1)
	if err != nil && err.Error() != statics.ErrorOrderNotFound {
		return nil, err
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
		return orderInfo.InitVolume
	}
	return orderInfo.InitVolume * orderInfo.InitPrice

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

func (m *OrderManager) DeleteTransferInfo(ctx context.Context, transferId string) {
	m.redisCli.DeleteFromSet(ctx, Transfers, transferId)
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

func (m *OrderManager) checkOrderAvailability(ctx context.Context, orderId string) (bool, error) {
	return m.redisCli.CheckInSet(ctx, OrdersAvailable, orderId)
}

func (s *OrderManager) TryLockOrder(ctx context.Context, id string) error {
	exists, err := s.redisCli.SetNXKey(ctx, "lock_"+OrdersHash+":"+id, s.instanceId)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("ResourceIsBlocked")
	}
	return nil
}

func (s *OrderManager) CheckLockOrder(ctx context.Context, id string) (bool, error) {
	_, err := s.redisCli.GetKey(ctx, "lock_"+OrdersHash+":"+id)
	if err != nil && err.Error() == statics.ErrorOrderNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil

}

func (s *OrderManager) TryUnlockOrder(ctx context.Context, id string) error {
	if err := s.redisCli.DelKeyWithValue(ctx, "lock_"+OrdersHash+":"+id, s.instanceId); err != nil {
		return err
	}
	return nil
}
