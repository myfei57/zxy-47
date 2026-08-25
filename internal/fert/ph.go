package fert

import (
	"math"
	"time"

	"greenhouse/internal/audit"
	"greenhouse/internal/store"
)

type PHResult struct {
	ShedID string    `json:"shed_id"`
	Before float64   `json:"before"`
	After  float64   `json:"after"`
	Stable bool      `json:"stable"`
	At     time.Time `json:"at"`
}

type PHController struct {
	kv    *store.KeyValue
	st    *store.Store
	audit *audit.Service
}

func NewPHController(st *store.Store, audit *audit.Service) *PHController {
	return &PHController{kv: store.NewKeyValue(st), st: st, audit: audit}
}

func (p *PHController) Stabilize(shedID string, current float64) (PHResult, error) {
	after := current
	if current < 5.5 {
		after = 5.8
	}
	if current > 6.8 {
		after = 6.4
	}
	result := PHResult{
		ShedID: shedID,
		Before: current,
		After:  after,
		Stable: math.Abs(after-current) < 0.05,
		At:     time.Now(),
	}
	if err := p.kv.Save("ph", shedID, result); err != nil {
		return PHResult{}, err
	}
	if err := p.order(shedID, "ph"); err != nil {
		return PHResult{}, err
	}
	if _, err := p.audit.Record(shedID, "fert", "ph", ftoa(after)); err != nil {
		return PHResult{}, err
	}
	return result, nil
}

func (p *PHController) Adjust(shedID string, current float64, ecCorrect func(string) error) (PHResult, error) {
	if err := ecCorrect(shedID); err != nil {
		return PHResult{}, err
	}
	result, err := p.Stabilize(shedID, current)
	if err != nil {
		return PHResult{}, err
	}
	return result, nil
}

func (p *PHController) Latest(shedID string) (PHResult, error) {
	var result PHResult
	if err := p.kv.Load("ph", shedID, &result); err != nil {
		return PHResult{}, err
	}
	return result, nil
}

func (p *PHController) order(shedID, step string) error {
	return p.st.AppendLine("adjust-order/"+shedID+".jsonl", step)
}
