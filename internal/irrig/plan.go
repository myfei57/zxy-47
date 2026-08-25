package irrig

import (
	"sort"
	"time"

	"greenhouse/internal/audit"
	"greenhouse/internal/env"
	"greenhouse/internal/store"
)

type Plan struct {
	ID               string    `json:"id"`
	ShedID           string    `json:"shed_id"`
	Name             string    `json:"name"`
	BaselineMoisture float64   `json:"baseline_moisture"`
	Threshold        float64   `json:"threshold"`
	DoseWater        float64   `json:"dose_water"`
	Enabled          bool      `json:"enabled"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type PlanStore struct {
	kv *store.KeyValue
}

func NewPlanStore(st *store.Store) *PlanStore {
	return &PlanStore{kv: store.NewKeyValue(st)}
}

func (p *PlanStore) Save(plan Plan) error {
	return p.kv.Save("irrig-plans", plan.ID, plan)
}

func (p *PlanStore) Get(id string) (Plan, error) {
	var plan Plan
	if err := p.kv.Load("irrig-plans", id, &plan); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (p *PlanStore) List(shedID string) ([]Plan, error) {
	ids, err := p.kv.List("irrig-plans")
	if err != nil {
		return nil, err
	}
	out := []Plan{}
	for _, id := range ids {
		plan, err := p.Get(id)
		if err == nil && plan.ShedID == shedID {
			out = append(out, plan)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (p *PlanStore) SetEnabled(id string, enabled bool) error {
	plan, err := p.Get(id)
	if err != nil {
		return err
	}
	plan.Enabled = enabled
	return p.Save(plan)
}

type RunResult struct {
	PlanID  string    `json:"plan_id"`
	ShedID  string    `json:"shed_id"`
	Baseline float64  `json:"baseline"`
	Watered bool      `json:"watered"`
	At      time.Time `json:"at"`
}

type Runner struct {
	plans  *PlanStore
	env    *env.MoistureStore
	valves *ValveStore
	audit  *audit.Service
}

func NewRunner(plans *PlanStore, moisture *env.MoistureStore, valves *ValveStore, audit *audit.Service) *Runner {
	return &Runner{plans: plans, env: moisture, valves: valves, audit: audit}
}

func (r *Runner) Run(shedID, planID string) (RunResult, error) {
	plan, err := r.plans.Get(planID)
	if err != nil {
		return RunResult{}, err
	}
	baseline, err := r.env.Read(shedID)
	if err != nil {
		baseline = plan.BaselineMoisture
	}
	watered := baseline < plan.Threshold
	if watered {
		if err := r.valves.Open(shedID, planID, plan.DoseWater); err != nil {
			return RunResult{}, err
		}
	}
	result := RunResult{PlanID: planID, ShedID: shedID, Baseline: baseline, Watered: watered, At: time.Now()}
	if _, err := r.audit.Record(shedID, "irrig", "run", planID); err != nil {
		return RunResult{}, err
	}
	return result, nil
}
