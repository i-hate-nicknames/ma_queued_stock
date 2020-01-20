package main

import (
	"errors"
	"sync"
)

const (
	STATUS_PENDING   = "pending"
	STATUS_CANCELLED = "cancelled"
	STATUS_COMPLETED = "completed"
)

type Store struct {
	machines    []*Machine
	orders      map[int]*Order
	nextOrderId int
	mux         sync.Mutex // todo: switch to RW mutext for better performance of read operations
}

func MakeStore(machineItems [][]int) *Store {
	orders := make(map[int]*Order)
	machines := make([]*Machine, 0)
	machineId := 1
	for _, mItems := range machineItems {
		machines = append(machines, MakeMachine(machineId, mItems))
		machineId++
	}
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
	s.mux.Lock()
	defer s.mux.Unlock()
	order := MakeOrder(items)
	s.orders[s.nextOrderId] = order
	s.nextOrderId += 1
	return s.nextOrderId - 1
}

func (s *Store) ResolveOrder(orderId int) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	order, ok := s.GetOrder(orderId)
	if !ok {
		return "", errors.New("Order not found!")
	}
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
	// todo: try to resolve all other orders if we changed state of at least one machine
	// todo: later we can use some scheduler structure with a separate routine, and schedule
	// order retries via it.
	// Probably can add some checks that will only schedule orders that can take something from
	// the updated state
	return order.status, nil
}

func (s *Store) GetOrder(orderId int) (*Order, bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	val, ok := s.orders[orderId]
	return val, ok
}

func (s *Store) CancelOrder(orderId int) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	if order, ok := s.orders[orderId]; ok {
		// todo: return all items the order has taken to some machine
		order.status = STATUS_CANCELLED
		return false
	}
	return false
}
