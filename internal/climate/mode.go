package climate

import (
	"time"

	"greenhouse/internal/store"
)

const (
	ModeHeat = "heat"
	ModeAuto = "auto"
	ModeVent = "vent"
)

type ModeState struct {
	ShedID string    `json:"shed_id"`
	Mode   string    `json:"mode"`
	Since  time.Time `json:"since"`
}

type ModeStore struct {
	kv *store.KeyValue
}

func NewModeStore(st *store.Store) *ModeStore {
	return &ModeStore{kv: store.NewKeyValue(st)}
}

func (m *ModeStore) Save(state ModeState) error {
	return m.kv.Save("climate-modes", state.ShedID, state)
}

func (m *ModeStore) Get(shedID string) (ModeState, error) {
	var state ModeState
	if err := m.kv.Load("climate-modes", shedID, &state); err != nil {
		return ModeState{}, err
	}
	return state, nil
}

func (m *ModeStore) Default(shedID string) ModeState {
	return ModeState{ShedID: shedID, Mode: ModeAuto, Since: time.Now()}
}
