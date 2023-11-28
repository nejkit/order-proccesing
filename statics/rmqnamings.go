package statics

const (
	ExNameOrders                 = "e.orders.forward"
	ExNameBalances               = "e.balances.forward"
	QueueNameCreateOrderRequest  = "q.orders.request.CreateOrderRequest"
	QueueNameCreateOrderResponse = "q.orders.response.CreateOrderResponse"
	QueueNameGetOrderResponse    = "q.orders.response.GetOrderResponse"
	QueueNameGetOrderRequest     = "q.orders.request.GetOrderRequest"
	QueueNameLockBalanceResponse = "q.balances.response.LockBalanceResponse"
	RkLockBalanceRequest         = "r.order-processing.request.LockBalanceRequest.#"
	RkCreateOrderRequest         = "r.request.#.CreateOrderRequest.#"
	RkCreateOrderResponse        = "r.response.#.CreateOrderResponse.#"
	RkGetOrderRequest            = "r.request.#.GetOrderRequest.#"
	RkGetOrderResponse           = "r.response.#.GetOrderResponse.#"
)
