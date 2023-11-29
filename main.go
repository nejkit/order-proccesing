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
	"time"

	logger "github.com/sirupsen/logrus"
)

func main() {
	logger.SetLevel(logger.InfoLevel)
	ctxRoot := context.Background()
	ctx, cancel := context.WithCancel(ctxRoot)
	rmqFactory := transportrabbit.NewFactory("amqp://admin:admin@rabbitmq:5672")
	rmqFactory.InitRmq()
	redisCli := storage.NewOrderManager("redis:6379")
	lockSender, err := rmqFactory.NewSender(ctx, statics.ExNameBalances, statics.RkLockBalanceRequest)
	if err != nil {
		cancel()
		return
	}
	createOrderSender, err := rmqFactory.NewSender(ctx, statics.ExNameOrders, statics.RkCreateOrderResponse)
	if err != nil {
		cancel()
		return
	}
	getOrderSender, err := rmqFactory.NewSender(ctx, statics.ExNameOrders, statics.RkGetOrderResponse)
	if err != nil {
		cancel()
		return
	}
	lockStorage := transportrabbit.NewAmqpStorage[balances.LockBalanceResponse](util.GetIdFromLockBalanceResponse())
	lockProcessor := transportrabbit.NewAmqpProcessor[balances.LockBalanceResponse](handlers.GetHandlerForLockBalanceProcessor(lockStorage), util.GetParserForLockBalanceResponse())
	lockListener, err := transportrabbit.NewListener[balances.LockBalanceResponse](
		ctx,
		rmqFactory,
		statics.QueueNameLockBalanceResponse,
		lockProcessor)
	if err != nil {
		cancel()
		return
	}

	transferSender, err := rmqFactory.NewSender(ctx, statics.ExNameBalances, statics.RkTransferBalanceRequest)
	if err != nil {
		cancel()
		return
	}
	balanceService := services.NewBalanceService(*lockSender, lockStorage)
	orderService := services.NewMarketOrderService(&redisCli, balanceService)
	matchingService := services.NewMatcherService(&redisCli, *transferSender)
	transferProcessor := transportrabbit.NewAmqpProcessor[balances.Transfer](handlers.GetHandlerForTransferProcessor(&matchingService), util.GetParserForTransfer())
	transferListener, err := transportrabbit.NewListener[balances.Transfer](
		ctx,
		rmqFactory,
		statics.TransferBalanceResponseQueue,
		transferProcessor)
	api := api.NewOrderApi(orderService, &matchingService, *createOrderSender, *getOrderSender)
	handler := handlers.NewHandler(api)
	getOrderProcessor := transportrabbit.NewAmqpProcessor[orders.GetOrderRequest](handler.GetHandlerForGetOrder(), util.GetParserForGetOrderRequest())
	createOrderProcessor := transportrabbit.NewAmqpProcessor[orders.CreateOrderRequest](handler.GetHandlerForCreateOrder(), util.GetParserForCreateOrderRequest())
	createOrderListener, err := transportrabbit.NewListener[orders.CreateOrderRequest](
		ctx,
		rmqFactory,
		statics.QueueNameCreateOrderRequest,
		createOrderProcessor)
	if err != nil {
		cancel()
		return
	}
	getOrderListener, err := transportrabbit.NewListener[orders.GetOrderRequest](
		ctx,
		rmqFactory,
		statics.QueueNameGetOrderRequest,
		getOrderProcessor)
	if err != nil {
		cancel()
		return
	}
	go lockListener.Run(ctx)
	go createOrderListener.Run(ctx)
	go getOrderListener.Run(ctx)
	go transferListener.Run(ctx)

	exit := make(chan os.Signal, 1)
	for {
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
		select {
		case <-exit:
			{
				cancel()
				return
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}
