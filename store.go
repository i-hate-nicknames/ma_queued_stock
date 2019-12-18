package main

const (
	STATUS_PENDING   = "pending"
	STATUS_CANCELLED = "cancelled"
	STATUS_COMPLETED = "completed"
)

type Store struct {
	machines    []*Machine
	orders      map[int]*Order
	nextOrderId int
}

func MakeStore() *Store {
	orders := make(map[int]*Order)
	// todo: init machines somehow
	machines := make([]*Machine, 0)
	return &Store{machines: machines, orders: orders, nextOrderId: 1}
}

type Order struct {
	items        []int
	fetchedItems []int
	status       string
}

func MakeOrder(items []int) *Order {
	fetched := make([]int, 0)
	return &Order{items: items, status: STATUS_PENDING, fetchedItems: fetched}
}

func (s *Store) SubmitOrder(items []int) int {
	// todo: lock because multiple requests can submit orders concurrently
	order := MakeOrder(items)
	s.orders[s.nextOrderId] = order
	s.nextOrderId += 1
	return s.nextOrderId - 1
}

func (s *Store) ResolveOrder(orderId int) {
	order, ok := s.GetOrder(orderId)
	if !ok {
		// todo: do something
	}
	// todo: lock order in case multiple routines may access the same order
	for _, m := range s.machines {
		taken, remains := m.TakeAll(order.items)
		order.items = remains
		for _, it := range taken {
			order.fetchedItems = append(order.fetchedItems, it)
		}
		if len(order.items) == 0 {
			break
		}
	}
	if len(order.items) == 0 {
		order.status = STATUS_COMPLETED
	} else {
		order.status = STATUS_PENDING
	}
}

func (s *Store) GetOrder(orderId int) (*Order, bool) {
	// todo: lock
	val, ok := s.orders[orderId]
	return val, ok
}

func (s *Store) CancelOrder(orderId int) bool {
	if order, ok := s.orders[orderId]; ok {
		order.status = STATUS_CANCELLED
	}
	return false
}
