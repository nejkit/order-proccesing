package main

import (
	"context"
	"order-processing/api"
	"order-processing/external/balances"
	"order-processing/external/orders"
	"order-processing/handlers"
	"order-processing/services"
	"order-processing/statics"
	"order-processing/storage"
	transportrabbit "order-processing/transport_rabbit"
	"order-processing/util"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	ctxRoot := context.Background()
	ctx, cancel := context.WithCancel(ctxRoot)
	rmqFactory := transportrabbit.NewFactory(logger, "amqp://admin:admin@rabbitmq:5672")
	redisCli := storage.NewOrderManager("@redis:6379", logger)
	lockSender, err := rmqFactory.NewSender(ctx, statics.ExNameBalances, statics.RkLockBalanceRequest)
	orderSender, err := rmqFactory.NewSender(ctx, statics.ExNameOrders, statics.RkCreateOrderResponse)
	if err != nil {
		logger.Errorln(err.Error())
		return
	}
	lockListener := transportrabbit.NewListener[balances.LockBalanceResponse](
		ctx,
		rmqFactory,
		statics.QueueNameLockBalanceResponse,
		util.GetParserForLockBalanceResponse(),
		nil,
		util.GetIdFromLockBalanceResponse())

	balanceService := services.NewBalanceService(logger, *lockSender, *lockListener)
	orderService := services.NewMarketOrderService(logger, redisCli, balanceService)
	api := api.NewOrderApi(logger, orderService, *orderSender)
	handler := handlers.NewHandler(api)

	createOrderListener := transportrabbit.NewListener[orders.CreateOrderRequest](
		ctx,
		rmqFactory,
		statics.QueueNameCreateOrderRequest,
		util.GetParserForCreateOrderRequest(),
		handler.GetHandlerForCreateOrder(),
		nil)

	go createOrderListener.Run(ctx)
	exit := make(chan os.Signal, 1)
	for {
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
		select {
		case <-exit:
			{
				cancel()
			}
		}
	}
}
