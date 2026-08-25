package irrig

import (
	"time"

	"greenhouse/internal/audit"
	"greenhouse/internal/store"
)

type Allowance struct {
	ShedID  string    `json:"shed_id"`
	Cycle   int       `json:"cycle"`
	Total   float64   `json:"total"`
	Used    float64   `json:"used"`
	Remain  float64   `json:"remain"`
	Updated time.Time `json:"updated_at"`
}

type AllowanceStore struct {
	kv *store.KeyValue
}

func NewAllowanceStore(st *store.Store) *AllowanceStore {
	return &AllowanceStore{kv: store.NewKeyValue(st)}
}

func (a *AllowanceStore) Save(state Allowance) error {
	return a.kv.Save("allowances", state.ShedID, state)
}

func (a *AllowanceStore) Get(shedID string) (Allowance, error) {
	var state Allowance
	if err := a.kv.Load("allowances", shedID, &state); err != nil {
		return Allowance{}, err
	}
	return state, nil
}

type AllowanceService struct {
	store *AllowanceStore
	audit *audit.Service
}

func NewAllowanceService(st *store.Store, audit *audit.Service) *AllowanceService {
	return &AllowanceService{store: NewAllowanceStore(st), audit: audit}
}

func (a *AllowanceService) Current(shedID string) (Allowance, error) {
	return a.store.Get(shedID)
}

func (a *AllowanceService) Consume(shedID string, amount float64) (Allowance, float64, error) {
	state, err := a.store.Get(shedID)
	if err != nil {
		return Allowance{}, 0, err
	}
	consumed := amount
	if consumed > state.Remain {
		consumed = state.Remain
	}
	state.Used += consumed
	state.Remain -= consumed
	state.Updated = time.Now()
	if err := a.store.Save(state); err != nil {
		return Allowance{}, 0, err
	}
	if _, err := a.audit.Record(shedID, "irrig", "consume", formatFloat(consumed)); err != nil {
		return Allowance{}, 0, err
	}
	return state, consumed, nil
}

func (a *AllowanceService) Reset(shedID string, total float64) (Allowance, error) {
	state := Allowance{ShedID: shedID, Total: total, Used: 0, Remain: total, Updated: time.Now()}
	if err := a.store.Save(state); err != nil {
		return Allowance{}, err
	}
	if _, err := a.audit.Record(shedID, "irrig", "allowance-reset", formatFloat(total)); err != nil {
		return Allowance{}, err
	}
	return state, nil
}

func formatFloat(value float64) string {
	return formatNumber(value)
}
