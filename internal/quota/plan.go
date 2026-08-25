package quota

import (
	"greenhouse/internal/store"
)

type Plan struct {
	ShedID string  `json:"shed_id"`
	Daily  float64 `json:"daily"`
	Unit   string  `json:"unit"`
}

type PlanStore struct {
	kv *store.KeyValue
}

func NewPlanStore(st *store.Store) *PlanStore {
	return &PlanStore{kv: store.NewKeyValue(st)}
}

func DefaultPlan(shedID string) Plan {
	return Plan{ShedID: shedID, Daily: 100, Unit: "升"}
}

func (p *PlanStore) Save(plan Plan) error {
	return p.kv.Save("quota-plans", plan.ShedID, plan)
}

func (p *PlanStore) Get(shedID string) (Plan, error) {
	var plan Plan
	if err := p.kv.Load("quota-plans", shedID, &plan); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (p *PlanStore) LoadOrDefault(shedID string) Plan {
	plan, err := p.Get(shedID)
	if err != nil {
		return DefaultPlan(shedID)
	}
	return plan
}
