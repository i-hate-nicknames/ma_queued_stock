package main

type OrderRegistry struct {
    orders map[int]*Order
    nextOrderId int
}

type Order struct {
    items []int
    status string
}

func MakeOrderRegistry() *OrderRegistry {
    orders := make(map[int]*Order)
    return &OrderRegistry{orders, 1}
}

func (reg *OrderRegistry) AddOrder(items []int) {
    // todo: lock
    order := &Order{items, "pending"}
    reg.orders[reg.nextOrderId] = order
    reg.nextOrderId += 1
}

func (reg *OrderRegistry) GetOrder(orderId int) (*Order, bool) {
    val, ok := reg.orders[orderId]
    return val, ok
}

func (reg *OrderRegistry) CancelOrder(orderId int) bool {
    if order, ok := reg.orders[orderId]; ok {
        order.status = "canceled"
    }
    return false
}
