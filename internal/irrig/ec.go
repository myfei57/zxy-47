package irrig

import (
	"time"

	"greenhouse/internal/audit"
	"greenhouse/internal/fert"
	"greenhouse/internal/store"
)

type ECService struct {
	kv    *store.KeyValue
	st    *store.Store
	audit *audit.Service
}

func NewECService(st *store.Store, audit *audit.Service) *ECService {
	return &ECService{kv: store.NewKeyValue(st), st: st, audit: audit}
}

func (e *ECService) Correct(shedID string, target float64) (fert.ECState, error) {
	state, err := e.Current(shedID)
	if err != nil {
		state = fert.ECState{ShedID: shedID}
	}
	state.Target = target
	state.Current = target
	state.UpdatedAt = time.Now()
	if err := e.kv.Save("ec", shedID, state); err != nil {
		return fert.ECState{}, err
	}
	if err := e.st.AppendLine("adjust-order/"+shedID+".jsonl", "ec"); err != nil {
		return fert.ECState{}, err
	}
	if _, err := e.audit.Record(shedID, "irrig", "ec-correct", ftoa(target)); err != nil {
		return fert.ECState{}, err
	}
	return state, nil
}

func (e *ECService) Current(shedID string) (fert.ECState, error) {
	var state fert.ECState
	if err := e.kv.Load("ec", shedID, &state); err != nil {
		return fert.ECState{}, err
	}
	return state, nil
}

func ftoa(value float64) string {
	return formatFloat(value)
}
