package auth

import (
	"github.com/google/uuid"
	"strings"
	"sync"
	"time"
)

const StateTTL = 5 * time.Minute

type StateManager struct {
	mu     sync.Mutex
	states map[string]string
}

func (m *StateManager) NewState(provider Provider) string {
	if m.states == nil {
		m.states = map[string]string{}
	}

	state := strings.Replace(uuid.New().String(), "-", "", 4)

	m.mu.Lock()
	m.states[state] = provider.Name()
	m.mu.Unlock()

	go func() {
		<-time.After(StateTTL)

		m.mu.Lock()
		delete(m.states, state)
		m.mu.Unlock()
	}()

	return state
}

func (m *StateManager) getProviderName(stateId string) (providerName string) {
	m.mu.Lock()

	providerName = m.states[stateId]
	delete(m.states, stateId)

	m.mu.Unlock()

	return
}
