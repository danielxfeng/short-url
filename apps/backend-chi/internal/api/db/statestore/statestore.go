package statestore

import "sync"

type StateStore interface {
	Add(state string, verifier string)
	GetAndDelete(state string) string
}

type MemoryStateStore struct {
	mu     sync.Mutex
	Status map[string]string // state -> verifier
}

func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		Status: make(map[string]string),
	}
}

func (m *MemoryStateStore) Add(state string, verifier string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Status[state] = verifier
}

func (m *MemoryStateStore) GetAndDelete(state string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	verifier, exists := m.Status[state]
	if exists {
		delete(m.Status, state) // State is single-use, delete it after validating
	}
	return verifier
}