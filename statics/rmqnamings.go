package statics

const (
	ExNameOrders                 = "e.orders.forward"
	ExNameBalances               = "e.balances.forward"
	QueueNameCreateOrderRequest  = "q.orders.request.CreateOrderRequest"
	QueueNameCreateOrderResponse = "q.orders.response.CreateOrderResponse"
	QueueNameLockBalanceResponse = "q.balances.response.LockBalanceResponse"
	RkLockBalanceRequest         = "r.order-processing.request.LockBalanceRequest.#"
	RkCreateOrderRequest         = "r.request.#.CreateOrderRequest.#"
	RkCreateOrderResponse        = "r.response.#.CreateOrderResponse.#"
	RkGetOrderResponse           = "r.response.#.GetOrderResponse.#"
)
