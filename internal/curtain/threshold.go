package curtain

import (
	"greenhouse/internal/store"
)

type Thresholds struct {
	LampSunlight float64 `json:"lamp_sunlight"`
	RetractMax   float64 `json:"retract_max"`
	RetractMin   float64 `json:"retract_min"`
}

type ThresholdStore struct {
	params *store.ParamsStore
}

func NewThresholdStore(st *store.Store) *ThresholdStore {
	return &ThresholdStore{params: store.NewParams(st)}
}

func DefaultThresholds() Thresholds {
	return Thresholds{LampSunlight: 300, RetractMax: 800, RetractMin: 200}
}

func (t *ThresholdStore) Save(shedID string, th Thresholds) error {
	return t.params.Save("curtain", shedID, map[string]any{
		"lamp_sunlight": th.LampSunlight,
		"retract_max":   th.RetractMax,
		"retract_min":   th.RetractMin,
	})
}

func (t *ThresholdStore) Load(shedID string) (Thresholds, error) {
	raw, err := t.params.Load("curtain", shedID)
	if err != nil {
		return Thresholds{}, err
	}
	th := DefaultThresholds()
	if value, ok := raw["lamp_sunlight"].(float64); ok {
		th.LampSunlight = value
	}
	if value, ok := raw["retract_max"].(float64); ok {
		th.RetractMax = value
	}
	if value, ok := raw["retract_min"].(float64); ok {
		th.RetractMin = value
	}
	return th, nil
}

func (t *ThresholdStore) LoadOrDefault(shedID string) Thresholds {
	th, err := t.Load(shedID)
	if err != nil {
		return DefaultThresholds()
	}
	return th
}
