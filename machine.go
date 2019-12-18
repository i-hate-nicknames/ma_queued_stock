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

func (m *Machine) Lock() {
	m.mux.Lock()
}

func (m *Machine) Unlock() {
	m.mux.Unlock()
}
