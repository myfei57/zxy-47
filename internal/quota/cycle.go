package quota

import (
	"time"

	"greenhouse/internal/audit"
	"greenhouse/internal/irrig"
	"greenhouse/internal/store"
)

type CycleService struct {
	quota  *QuotaStore
	plans  *PlanStore
	allow  *irrig.AllowanceService
	ledger *Ledger
	audit  *audit.Service
}

func NewCycleService(
	st *store.Store,
	plans *PlanStore,
	allow *irrig.AllowanceService,
	ledger *Ledger,
	audit *audit.Service,
) *CycleService {
	return &CycleService{
		quota:  NewQuotaStore(st),
		plans:  plans,
		allow:  allow,
		ledger: ledger,
		audit:  audit,
	}
}

func (c *CycleService) Release(shedID string) (Quota, error) {
	plan := c.plans.LoadOrDefault(shedID)
	cycle := c.quota.NextCycle(shedID)
	current, err := c.allow.Current(shedID)
	if err != nil {
		current = irrig.Allowance{ShedID: shedID}
	}
	used := current.Total - current.Remain
	if used < 0 {
		used = 0
	}
	if err := c.ledger.Append(shedID, cycle, used); err != nil {
		return Quota{}, err
	}
	total := current.Remain + plan.Daily
	allowance, err := c.allow.Reset(shedID, total)
	if err != nil {
		return Quota{}, err
	}
	quota := Quota{ShedID: shedID, Cycle: cycle, Daily: plan.Daily, Used: 0, Remain: allowance.Remain, At: time.Now()}
	if err := c.quota.Save(quota); err != nil {
		return Quota{}, err
	}
	if _, err := c.audit.Record(shedID, "quota", "release", itoa(int64(cycle))); err != nil {
		return Quota{}, err
	}
	return quota, nil
}

func (c *CycleService) Current(shedID string) (Quota, error) {
	return c.quota.Get(shedID)
}

func itoa(value int64) string {
	return formatInt(value)
}
