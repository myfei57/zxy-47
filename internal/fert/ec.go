package fert

import (
	"time"

	"greenhouse/internal/audit"
	"greenhouse/internal/store"
)

type ECState struct {
	ShedID    string    `json:"shed_id"`
	Target    float64   `json:"target"`
	Current   float64   `json:"current"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ECStore struct {
	kv *store.KeyValue
}

func NewECStore(st *store.Store) *ECStore {
	return &ECStore{kv: store.NewKeyValue(st)}
}

func (e *ECStore) Get(shedID string) (ECState, error) {
	var state ECState
	if err := e.kv.Load("ec", shedID, &state); err != nil {
		return ECState{}, err
	}
	return state, nil
}

func (e *ECStore) Save(state ECState) error {
	return e.kv.Save("ec", state.ShedID, state)
}

type ECController struct {
	store *ECStore
	audit *audit.Service
}

func NewECController(ecStore *ECStore, audit *audit.Service) *ECController {
	return &ECController{store: ecStore, audit: audit}
}

func (c *ECController) SetTarget(shedID string, target float64) (ECState, error) {
	state, err := c.store.Get(shedID)
	if err != nil {
		state = ECState{ShedID: shedID}
	}
	state.Target = target
	state.Current = target
	state.UpdatedAt = time.Now()
	if err := c.store.Save(state); err != nil {
		return ECState{}, err
	}
	if _, err := c.audit.Record(shedID, "fert", "ec-target", ftoa(target)); err != nil {
		return ECState{}, err
	}
	return state, nil
}

func (c *ECController) Target(shedID string) (ECState, error) {
	return c.store.Get(shedID)
}
