package climate

import (
	"greenhouse/internal/audit"
	"greenhouse/internal/store"
	"greenhouse/internal/vent"
)

type CoolPlanner struct {
	guards *GuardStore
	vent   *vent.Controller
	audit  *audit.Service
}

func NewCoolPlanner(st *store.Store, ventController *vent.Controller, audit *audit.Service) *CoolPlanner {
	return &CoolPlanner{guards: NewGuardStore(st), vent: ventController, audit: audit}
}

func (p *CoolPlanner) Cool(shedID string) error {
	wind := p.guards.Wind(shedID)
	if err := p.vent.OpenForCooling(shedID, wind); err != nil {
		return err
	}
	_, err := p.audit.Record(shedID, "climate", "cool", wind)
	return err
}

func (p *CoolPlanner) Guards(shedID string) GuardParams {
	return p.guards.LoadOrDefault(shedID)
}

func (p *CoolPlanner) SetGuards(shedID string, params GuardParams) error {
	return p.guards.Save(shedID, params)
}
