package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type MatchingData struct {
	FillVolume float64
	FillPrice  float64
	Date       uint64
}

type OrderInfo struct {
	Id string

	SellCurrency   string
	BuyCurrency    string
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
	logger   *logrus.Logger
}

func NewOrderManager(connectionString string, logger *logrus.Logger) OrderManager {
	redisCli := GetNewRedisCli(logger, connectionString)
	return OrderManager{redisCli: redisCli, logger: logger}
}

func genKeyByOrderData(orderData OrderInfo) string {
	//sell cur, buycur, price, creat_date, volume, ex address, id order
	keyData := [7]string{
		orderData.SellCurrency,
		orderData.BuyCurrency,
		fmt.Sprintf("%v", orderData.InitPrice),
		fmt.Sprintf("%d", orderData.CreationDate),
		fmt.Sprintf("%v", orderData.InitVolume),
		orderData.ExchangeWallet,
		orderData.Id}
	return strings.Join(keyData[:], ":")
}

func (om *OrderManager) InsertNewOrder(ctx context.Context, request OrderInfo) {
	key := genKeyByOrderData(request)
	value, err := json.Marshal(request)

	if err != nil {
		om.logger.Errorln("Error encode to JSON! message: ", err.Error())
		return
	}

	err = om.redisCli.Insert(ctx, key, value)
	if err != nil {
		om.logger.Errorln("Insert failed! message: ", err.Error())
	}
}

func (om *OrderManager) GetOrderById(ctx context.Context, id string) *OrderInfo {
	pattern := "*:*:*:*:*:*:" + id
	result, err := om.redisCli.GetByPattern(ctx, pattern)
	if err != nil {
		om.logger.Errorln("Search failed! message: ", err.Error())
		return nil
	}
	if result != nil {
		var dataModel OrderInfo
		err = json.Unmarshal([]byte(result[0]), dataModel)
		if err != nil {
			om.logger.Errorln("Err parse the value from redis! message: ", err.Error())
			return nil
		}
		return &dataModel
	}
	return nil
}
