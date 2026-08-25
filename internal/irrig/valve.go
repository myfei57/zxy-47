package irrig

import (
	"time"

	"greenhouse/internal/store"
)

type ValveState struct {
	ID        string    `json:"id"`
	ShedID    string    `json:"shed_id"`
	PlanID    string    `json:"plan_id"`
	Open      bool      `json:"open"`
	Watered   float64   `json:"watered"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ValveStore struct {
	kv *store.KeyValue
}

func NewValveStore(st *store.Store) *ValveStore {
	return &ValveStore{kv: store.NewKeyValue(st)}
}

func (v *ValveStore) Save(state ValveState) error {
	return v.kv.Save("valves", state.ShedID, state)
}

func (v *ValveStore) Get(shedID string) (ValveState, error) {
	var state ValveState
	if err := v.kv.Load("valves", shedID, &state); err != nil {
		return ValveState{}, err
	}
	return state, nil
}

func (v *ValveStore) Open(shedID, planID string, amount float64) error {
	state, err := v.Get(shedID)
	if err != nil {
		state = ValveState{ID: shedID + "-valve", ShedID: shedID}
	}
	state.PlanID = planID
	state.Open = true
	state.Watered += amount
	state.UpdatedAt = time.Now()
	return v.Save(state)
}

func (v *ValveStore) Close(shedID string) error {
	state, err := v.Get(shedID)
	if err != nil {
		return err
	}
	state.Open = false
	state.UpdatedAt = time.Now()
	return v.Save(state)
}
