package irrig

import (
	"greenhouse/internal/audit"
	"greenhouse/internal/fert"
)

type CycleResult struct {
	Run   RunResult      `json:"run"`
	Dose  Dose           `json:"dose"`
	Allow Allowance      `json:"allow"`
	Mix   fert.MixResult `json:"mix"`
}

type CycleService struct {
	runner *Runner
	doses  *DoseService
	mixer  *fert.Mixer
	allow  *AllowanceService
	audit  *audit.Service
}

func NewCycleService(
	runner *Runner,
	doses *DoseService,
	mixer *fert.Mixer,
	allow *AllowanceService,
	audit *audit.Service,
) *CycleService {
	return &CycleService{runner: runner, doses: doses, mixer: mixer, allow: allow, audit: audit}
}

func (c *CycleService) Run(shedID, planID string) (CycleResult, error) {
	run, err := c.runner.Run(shedID, planID)
	if err != nil {
		return CycleResult{}, err
	}
	if !run.Watered {
		return CycleResult{Run: run}, nil
	}
	dose, err := c.doses.Compute(shedID)
	if err != nil {
		return CycleResult{}, err
	}
	allowance, consumed, err := c.allow.Consume(shedID, dose.Water)
	if err != nil {
		return CycleResult{}, err
	}
	concentrate := dose.Concentrate
	if dose.Water > 0 && consumed < dose.Water {
		concentrate = dose.Concentrate * (consumed / dose.Water)
	}
	mix, err := c.mixer.Mix(shedID, consumed, concentrate)
	if err != nil {
		return CycleResult{}, err
	}
	if _, err := c.audit.Record(shedID, "irrig", "cycle", planID); err != nil {
		return CycleResult{}, err
	}
	return CycleResult{Run: run, Dose: dose, Allow: allowance, Mix: mix}, nil
}
