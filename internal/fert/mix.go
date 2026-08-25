package fert

import (
	"time"

	"greenhouse/internal/audit"
	"greenhouse/internal/store"
)

type MixStep struct {
	Order  int     `json:"order"`
	Kind   string  `json:"kind"`
	Amount float64 `json:"amount"`
}

type MixResult struct {
	ShedID string    `json:"shed_id"`
	EC     float64   `json:"ec"`
	Steps  []MixStep `json:"steps"`
	At     time.Time `json:"at"`
}

type Mixer struct {
	kv    *store.KeyValue
	audit *audit.Service
}

func NewMixer(st *store.Store, audit *audit.Service) *Mixer {
	return &Mixer{kv: store.NewKeyValue(st), audit: audit}
}

func (m *Mixer) Mix(shedID string, water, concentrate float64) (MixResult, error) {
	steps := []MixStep{
		{Order: 1, Kind: "concentrate", Amount: concentrate},
		{Order: 2, Kind: "water", Amount: water},
	}
	result := MixResult{ShedID: shedID, EC: m.mixEC(steps), Steps: steps, At: time.Now()}
	if err := m.kv.Save("mixes", shedID, result); err != nil {
		return MixResult{}, err
	}
	if _, err := m.audit.Record(shedID, "fert", "mix", "water+"+concat(concentrate)); err != nil {
		return MixResult{}, err
	}
	return result, nil
}

func (m *Mixer) Latest(shedID string) (MixResult, error) {
	var result MixResult
	if err := m.kv.Load("mixes", shedID, &result); err != nil {
		return MixResult{}, err
	}
	return result, nil
}

func (m *Mixer) mixEC(steps []MixStep) float64 {
	ec := 1.2
	for _, step := range steps {
		if step.Kind == "water" {
			ec += step.Amount * 0.01
		}
		if step.Kind == "concentrate" {
			ec += step.Amount * 0.8
		}
	}
	if len(steps) > 0 && steps[0].Kind == "concentrate" {
		ec += 1.0
	}
	return ec
}

func concat(value float64) string {
	if value == float64(int64(value)) {
		return itoa(int64(value))
	}
	return ftoa(value)
}
