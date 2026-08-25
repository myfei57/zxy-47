package task

import (
	"greenhouse/internal/store"
)

type StageTemplate struct {
	Label string  `json:"label"`
	EC    float64 `json:"ec"`
	PH    float64 `json:"ph"`
}

type StagePlan struct {
	ShedID string          `json:"shed_id"`
	Stages []StageTemplate `json:"stages"`
}

type StagePlanStore struct {
	kv *store.KeyValue
}

func NewStagePlanStore(st *store.Store) *StagePlanStore {
	return &StagePlanStore{kv: store.NewKeyValue(st)}
}

func DefaultStagePlan(shedID string) StagePlan {
	return StagePlan{
		ShedID: shedID,
		Stages: []StageTemplate{
			{Label: "育苗期", EC: 1.4, PH: 5.8},
			{Label: "营养生长期", EC: 2.0, PH: 6.2},
			{Label: "开花坐果期", EC: 2.6, PH: 6.4},
			{Label: "膨果采收期", EC: 2.2, PH: 6.2},
		},
	}
}

func (p *StagePlanStore) Save(plan StagePlan) error {
	return p.kv.Save("stage-plans", plan.ShedID, plan)
}

func (p *StagePlanStore) Get(shedID string) (StagePlan, error) {
	var plan StagePlan
	if err := p.kv.Load("stage-plans", shedID, &plan); err != nil {
		return StagePlan{}, err
	}
	return plan, nil
}

func (p *StagePlanStore) LoadOrDefault(shedID string) StagePlan {
	plan, err := p.Get(shedID)
	if err != nil {
		return DefaultStagePlan(shedID)
	}
	return plan
}

func (p *StagePlanStore) Template(shedID, label string) (StageTemplate, bool) {
	for _, template := range p.LoadOrDefault(shedID).Stages {
		if template.Label == label {
			return template, true
		}
	}
	return StageTemplate{}, false
}
