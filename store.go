package main

import (
	"errors"
	"log"
	"sync"

	"github.com/jinzhu/gorm"
)

const (
	STATUS_PENDING   = "pending"
	STATUS_CANCELLED = "cancelled"
	STATUS_COMPLETED = "completed"
)

type Store struct {
	machines    []*Machine
	orders      map[uint]*Order
	nextOrderId uint
	mux         sync.Mutex // todo: switch to RW mutext for better performance of read operations
	db          *gorm.DB
}

// Load state of the store from the database
func LoadStore(db *gorm.DB) *Store {
	dbOrders := LoadOrders(db)
	machines := LoadMachines(db)
	// todo: remove debug machine
	machines = append(machines, MakeMachine(10, []int{5, 4, 3, 2, 1}))
	orders := make(map[uint]*Order, 0)
	for _, dbOrder := range dbOrders {
		orders[dbOrder.ID] = dbOrder
	}
	return &Store{machines: machines, orders: orders, nextOrderId: 1, db: db}
}

type Order struct {
	ID           uint
	items        []int
	fetchedItems []int
	status       string
	mux          sync.Mutex
}

func MakeOrder(items []int) *Order {
	fetched := make([]int, 0)
	return &Order{items: items, status: STATUS_PENDING, fetchedItems: fetched}
}

func (s *Store) SubmitOrder(items []int) uint {
	s.mux.Lock()
	defer s.mux.Unlock()
	order := MakeOrder(items)
	id := SaveOrder(s.db, order)
	log.Println("Created order with id", id)
	s.orders[id] = order
	return id
}

func (s *Store) ResolveOrder(orderId uint) (string, error) {
	order, ok := s.GetOrder(orderId)
	if !ok {
		return "", errors.New("Order not found!")
	}
	order.mux.Lock()
	defer order.mux.Unlock()
	orderChanged := false
	// this assumes s.machines will never be updated simultaneously with this method
	for _, m := range s.machines {
		taken, remains := m.TakeAll(order.items)
		order.items = remains
		for _, it := range taken {
			order.fetchedItems = append(order.fetchedItems, it)
		}
		if len(taken) > 0 {
			orderChanged = true
		}
		if len(order.items) == 0 {
			break
		}
	}
	if orderChanged {
		if len(order.items) == 0 {
			order.status = STATUS_COMPLETED
		} else {
			order.status = STATUS_PENDING
		}
		SaveOrder(s.db, order)
	}
	// todo: try to resolve all other orders if we changed state of at least one machine
	// todo: later we can use some scheduler structure with a separate routine, and schedule
	// order retries via it.
	// Probably can add some checks that will only schedule orders that can take something from
	// the updated state
	return order.status, nil
}

func (s *Store) GetOrder(orderId uint) (*Order, bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	val, ok := s.orders[orderId]
	return val, ok
}

func (s *Store) CancelOrder(orderId uint) error {
	order, ok := s.GetOrder(orderId)
	if !ok {
		return errors.New("Order not found")
	}
	s.mux.Lock()
	if len(s.machines) == 0 {
		return errors.New("no machines to put cancelled order items")
	}
	s.mux.Unlock()
	order.mux.Lock()
	defer order.mux.Unlock()
	// this assumes s.machines will never be updated simultaneously with this method
	m := s.machines[0]
	m.PutAll(order.fetchedItems)
	order.status = STATUS_CANCELLED
	return nil
}
