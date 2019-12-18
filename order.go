package main

type OrderRegistry struct {
	orders      map[int]*Order
	nextOrderId int
}

const (
	STATUS_PENDING   = "pending"
	STATUS_CANCELLED = "cancelled"
)

type Order struct {
	items        []int
	fetchedItems []int
	status       string
}

func MakeOrderRegistry() *OrderRegistry {
	orders := make(map[int]*Order)
	return &OrderRegistry{orders, 1}
}

func (reg *OrderRegistry) AddOrder(items []int) {
	// todo: lock
	order := MakeOrder(items)
	reg.orders[reg.nextOrderId] = order
	reg.nextOrderId += 1
}

func (reg *OrderRegistry) GetOrder(orderId int) (*Order, bool) {
	val, ok := reg.orders[orderId]
	return val, ok
}

func (reg *OrderRegistry) CancelOrder(orderId int) bool {
	if order, ok := reg.orders[orderId]; ok {
		order.status = STATUS_CANCELLED
	}
	return false
}

func MakeOrder(items []int) *Order {
	fetched := make([]int, 0)
	return &Order{items: items, status: STATUS_PENDING, fetchedItems: fetched}
}
