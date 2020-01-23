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
	machines []*Machine
	orders   map[uint]*Order
	mux      sync.Mutex // todo: switch to RW mutext for better performance of read operations
	db       *gorm.DB
}

// LoadStore loads state of the store from the database
func LoadStore(db *gorm.DB) *Store {
	dbOrders := LoadOrders(db)
	machines := LoadMachines(db)
	orders := make(map[uint]*Order, 0)
	for _, dbOrder := range dbOrders {
		orders[dbOrder.ID] = dbOrder
	}
	return &Store{machines: machines, orders: orders, db: db}
}

func (s *Store) generateMachines() {
	ms := make([]*Machine, 0)
	ms = append(ms, MakeMachine(10, []int{5, 4, 3, 2, 1}))
	ms = append(ms, MakeMachine(15, []int{44, 32, 12}))
	ms = append(ms, MakeMachine(25, []int{1, 2, 3}))
	for _, m := range ms {
		SaveMachine(s.db, m)
	}
	s.machines = ms
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

func (o *Order) Copy() *Order {
	cp := &Order{ID: o.ID, status: o.status}
	copy(cp.items, o.items)
	copy(cp.fetchedItems, cp.fetchedItems)
	return cp
}

func (o *Order) RestoreFrom(other *Order) {
	o.items = other.items
	o.fetchedItems = other.fetchedItems
	o.status = other.status
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
	// this assumes s.machines will never be updated simultaneously with this method
	for _, machine := range s.machines {
		// lock machine, try to take as many items as possible
		// if taken any, start db transaction, save both updated order and machine
		// within a transaction. If that fails, rollback the states of order and machine
		machine.mux.Lock()
		err := ExecOrRestore(order, machine, func() error {
			var err error
			taken, remains := machine.TakeAll(order.items)
			order.items = remains
			for _, it := range taken {
				order.fetchedItems = append(order.fetchedItems, it)
			}
			if len(taken) > 0 {
				if len(order.items) == 0 {
					order.status = STATUS_COMPLETED
				} else {
					order.status = STATUS_PENDING
				}
				err = UpdateAtomically(s.db, order, machine)
			}
			return err
		})
		if err != nil {
			log.Println("Error when taking items from machine", err)
		}
		machine.mux.Unlock()
		if len(order.items) == 0 {
			break
		}
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
	machine := s.machines[0]
	machine.mux.Lock()
	err := ExecOrRestore(order, machine, func() error {
		machine.PutAll(order.fetchedItems)
		order.fetchedItems = []int{}
		order.status = STATUS_CANCELLED
		if err := UpdateAtomically(s.db, order, machine); err != nil {
			return err
		}
		return nil
	})
	machine.mux.Unlock()

	return err
}

// ExecOrRestore copies order and machine data, executes f. If f returns an error,
// the state of order and machine are rolled back
func ExecOrRestore(o *Order, m *Machine, f func() error) error {
	// todo: maybe consider splitting this into two methods SafeExec for
	// order and machine and just chain them together here tbh
	orderCopy := o.Copy()
	machineCopy := m.Copy()
	if err := f(); err != nil {
		o.RestoreFrom(orderCopy)
		m.RestoreFrom(machineCopy)
		return err
	}
	return nil
}
