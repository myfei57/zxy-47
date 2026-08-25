package curtain

import (
	"time"

	"greenhouse/internal/store"
)

type State struct {
	ShedID   string    `json:"shed_id"`
	ZoneID   string    `json:"zone_id"`
	Position string    `json:"position"`
	At       time.Time `json:"at"`
}

type StateStore struct {
	kv *store.KeyValue
}

func NewStateStore(st *store.Store) *StateStore {
	return &StateStore{kv: store.NewKeyValue(st)}
}

func (s *StateStore) Save(state State) error {
	return s.kv.Save("curtains", state.ZoneID, state)
}

func (s *StateStore) Get(zoneID string) (State, error) {
	var state State
	if err := s.kv.Load("curtains", zoneID, &state); err != nil {
		return State{}, err
	}
	return state, nil
}

type LampState struct {
	ZoneID string    `json:"zone_id"`
	On     bool      `json:"on"`
	Reason string    `json:"reason"`
	At     time.Time `json:"at"`
}

type LampStateStore struct {
	kv *store.KeyValue
}

func NewLampStateStore(st *store.Store) *LampStateStore {
	return &LampStateStore{kv: store.NewKeyValue(st)}
}

func (l *LampStateStore) Save(state LampState) error {
	return l.kv.Save("lamps", state.ZoneID, state)
}

func (l *LampStateStore) Get(zoneID string) (LampState, error) {
	var state LampState
	if err := l.kv.Load("lamps", zoneID, &state); err != nil {
		return LampState{}, err
	}
	return state, nil
}
