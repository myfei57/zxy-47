package climate

import (
	"greenhouse/internal/curtain"
	"greenhouse/internal/env"
	"greenhouse/internal/vent"
)

type Monitor struct {
	guards  *GuardStore
	sampler *env.Sampler
	curtain *curtain.Controller
	vent    *vent.Controller
	fans    *vent.FanController
}

func NewMonitor(
	guards *GuardStore,
	sampler *env.Sampler,
	curtainController *curtain.Controller,
	ventController *vent.Controller,
	fanController *vent.FanController,
) *Monitor {
	return &Monitor{
		guards:  guards,
		sampler: sampler,
		curtain: curtainController,
		vent:    ventController,
		fans:    fanController,
	}
}

func (m *Monitor) Evaluate(shedID string) error {
	history, err := m.sampler.TempHistory(shedID)
	if err != nil {
		history = []float64{}
	}
	if err := m.fans.Evaluate(shedID, history); err != nil {
		return err
	}
	sun := m.sampler.Value(shedID, env.TypeSunlight, 0)
	th := m.curtain.Thresholds(shedID)
	zone := shedID + "-Z01"
	if sun >= th.RetractMax {
		if err := m.curtain.Retract(shedID, zone); err != nil {
			return err
		}
		return m.vent.OpenForCooling(shedID, m.guards.Wind(shedID))
	}
	if sun < th.RetractMin {
		return m.curtain.Extend(shedID, zone)
	}
	return nil
}
