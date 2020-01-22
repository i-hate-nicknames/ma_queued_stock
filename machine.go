package main

import (
	"sync"
)

type Machine struct {
	ID      uint
	in, out []int
	mux     sync.Mutex
}

func MakeMachine(id uint, items []int) *Machine {
	in := make([]int, 0)
	out := make([]int, len(items))
	copy(out, items)
	return &Machine{in: in, out: out, ID: id}
}

func (m *Machine) PutAll(items []int) {
	m.mux.Lock()
	defer m.mux.Unlock()
	for _, item := range items {
		m.put(item)
	}
}

func (m *Machine) put(item int) {
	m.in = append(m.in, item)
}

// TakeAll tries to take as many items as possible from this machine,
// return two slices (taken, remains), the first one contains all
// the item taken from this machine, the second one all the items
// in toTake that can't be taken from this machine
func (m *Machine) TakeAll(orderItems []int) ([]int, []int) {
	noneTaken := false
	toTake := make(map[int]int, 0)
	for _, orderItem := range orderItems {
		toTake[orderItem]++
	}

	for !noneTaken {
		noneTaken = true
		topItem, ok := m.peek()
		if !ok {
			break
		}

		if qty, ok := toTake[topItem]; ok && qty > 0 {
			m.take()
			toTake[topItem]--
			noneTaken = false
		}
	}

	taken := make([]int, 0)
	remains := make([]int, 0)
	for _, orderItem := range orderItems {
		if qty, ok := toTake[orderItem]; ok && qty > 0 {
			toTake[orderItem]--
			remains = append(remains, orderItem)
		} else {
			taken = append(taken, orderItem)
		}
	}
	return taken, remains
}

// RestoreFrom takes the same machine's copy, and resets its items
// state to that copy.
// Do nothing if given machine is not a copy (checked by machine id)
func (m *Machine) RestoreFrom(other *Machine) {
	if other.ID != m.ID {
		return
	}
	m.in = other.in
	m.out = other.out
}

func (m *Machine) Copy() *Machine {
	cp := &Machine{ID: m.ID}
	copy(cp.in, m.in)
	copy(cp.out, m.out)
	return cp
}

// GetAllItems gets all items in this machine in the queue order
func (m *Machine) GetAllItems() []int {
	result := make([]int, 0)
	for i := len(m.out) - 1; i >= 0; i-- {
		result = append(result, m.out[i])
	}
	for i := 0; i < len(m.in); i++ {
		result = append(result, m.in[i])
	}
	return result
}

func (m *Machine) take() (int, bool) {
	if len(m.out) > 0 {
		topItem := m.out[len(m.out)-1]
		m.out = m.out[:len(m.out)-1]
		return topItem, true
	} else if len(m.in) > 0 {
		// put everything in m.in into m.out in reversed order
		// except for the item to take
		m.out = make([]int, len(m.in))
		for i := len(m.in) - 1; i > 0; i-- {
			m.out = append(m.out, m.in[i])
		}
		item := m.in[0]
		m.in = make([]int, 0)
		return item, true
	} else {
		return 0, false
	}
}

func (m *Machine) peek() (int, bool) {
	if len(m.out) > 0 {
		return m.out[len(m.out)-1], true
	} else if len(m.in) > 0 {
		return m.in[0], true
	} else {
		return 0, false
	}
}
