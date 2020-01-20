package main

import "sync"

type Machine struct {
	in, out []int
	mux     sync.Mutex
}

func (m *Machine) Put(item int) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.in = append(m.in, item)
}

func (m *Machine) Take() (int, bool) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if len(m.out) > 0 {
		out := m.out[len(m.out)-1]
		m.out = m.out[:len(m.out)]
		return out, true
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

// TakeAll tries to take as many items as possible from this machine,
// return two slices (taken, remains), the first one contains all
// the item taken from this machine, the second one all the items
// in toTake that can't be taken from this machine
func (m *Machine) TakeAll(toTake []int) ([]int, []int) {
	m.mux.Lock()
	defer m.mux.Unlock()
	takenSet := make(map[int]bool, 0)
	noneTaken := false
	// instead of this nonsence it would have been much easier
	// to just instantiate order items as a map and lookup
	// top item tbh
	for !noneTaken {
		noneTaken = true
		for _, toTakeItem := range toTake {
			item, ok := m.peek()
			if !ok {
				return groupItems(toTake, takenSet)
			}
			if item == toTakeItem {
				_, _ = m.Take()
				noneTaken = false
				takenSet[item] = true
			}
		}
	}
	return groupItems(toTake, takenSet)
}

// group items in given slice in two groups, depending whether a given
// item is in set or not. Return slices (present, absent)
func groupItems(items []int, set map[int]bool) ([]int, []int) {
	taken, remains := make([]int, 0), make([]int, 0)
	for _, item := range items {
		if set[item] {
			taken = append(taken, item)
		} else {
			remains = append(remains, item)
		}
	}
	return taken, remains
}
