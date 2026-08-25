package console

import (
	"greenhouse/internal/audit"
	"greenhouse/internal/climate"
	"greenhouse/internal/curtain"
	"greenhouse/internal/env"
	"greenhouse/internal/fert"
	"greenhouse/internal/irrig"
	"greenhouse/internal/ns"
	"greenhouse/internal/quota"
	"greenhouse/internal/shed"
	"greenhouse/internal/store"
	"greenhouse/internal/task"
	"greenhouse/internal/vent"
)

type Services struct {
	Store     *store.Store
	Sites     *ns.Registry
	Sheds     *shed.Service
	Sensors   *env.SensorRegistry
	Sampler   *env.Sampler
	Bindings  *env.BindingService
	Sun       *env.SunlightStore
	Moisture  *env.MoistureStore
	Audit     *audit.Service
	Alarms    *audit.AlarmService
	Curtain   *curtain.Controller
	Vent      *vent.Controller
	Fans      *vent.FanController
	Roller    *vent.FilmRoller
	Climate   *climate.Service
	Cool      *climate.CoolPlanner
	Monitor   *climate.Monitor
	Fert      *fert.NutrientService
	PH        *fert.PHController
	ECIrrig   *irrig.ECService
	Mixer     *fert.Mixer
	Plans     *irrig.PlanStore
	Doses     *irrig.DoseService
	Allow     *irrig.AllowanceService
	Cycle     *irrig.CycleService
	Stages    *task.StageService
	StagePlan *task.StagePlanStore
	Tasks     *task.TaskService
	Calendar  *task.CalendarStore
	Quotas    *quota.CycleService
	QuotaPlan *quota.PlanStore
	Ledger    *quota.Ledger
}

func Build(st *store.Store) *Services {
	auditSvc := audit.NewService(st)
	alarms := audit.NewAlarmService(st)
	sensors := env.NewSensorRegistry(st)
	bindings := env.NewBindingService(st, sensors)
	sun := env.NewSunlightStore(st)
	moisture := env.NewMoistureStore(st)
	filter := env.NewTemperatureFilter(5)
	sampler := env.NewSampler(sensors, st, sun, moisture, filter)
	sites := ns.NewRegistry(st)
	sheds := shed.NewService(st, bindings)
	curtains := curtain.NewController(st, sun, sampler, auditSvc)
	vents := vent.NewController(st, auditSvc)
	fans := vent.NewFanController(st, filter, auditSvc, alarms)
	roller := vent.NewFilmRoller(st, auditSvc, alarms)
	climateSvc := climate.NewService(st, auditSvc)
	cool := climate.NewCoolPlanner(st, vents, auditSvc)
	guards := climate.NewGuardStore(st)
	monitor := climate.NewMonitor(guards, sampler, curtains, vents, fans)
	recipes := fert.NewRecipeStore(st)
	ph := fert.NewPHController(st, auditSvc)
	ecFert := fert.NewECController(fert.NewECStore(st), auditSvc)
	ecIrrig := irrig.NewECService(st, auditSvc)
	mixer := fert.NewMixer(st, auditSvc)
	nutrient := fert.NewNutrientService(ph, ecIrrig, mixer, recipes, auditSvc)
	plans := irrig.NewPlanStore(st)
	valves := irrig.NewValveStore(st)
	runner := irrig.NewRunner(plans, moisture, valves, auditSvc)
	doses := irrig.NewDoseService(recipes, st)
	allow := irrig.NewAllowanceService(st, auditSvc)
	cycle := irrig.NewCycleService(runner, doses, mixer, allow, auditSvc)
	stages := task.NewStageService(st, ecFert, auditSvc)
	stagePlan := task.NewStagePlanStore(st)
	tasks := task.NewTaskService(st, auditSvc)
	calendar := task.NewCalendarStore(st)
	quotaPlan := quota.NewPlanStore(st)
	ledger := quota.NewLedger(st)
	quotas := quota.NewCycleService(st, quotaPlan, allow, ledger, auditSvc)
	return &Services{
		Store:     st,
		Sites:     sites,
		Sheds:     sheds,
		Sensors:   sensors,
		Sampler:   sampler,
		Bindings:  bindings,
		Sun:       sun,
		Moisture:  moisture,
		Audit:     auditSvc,
		Alarms:    alarms,
		Curtain:   curtains,
		Vent:      vents,
		Fans:      fans,
		Roller:    roller,
		Climate:   climateSvc,
		Cool:      cool,
		Monitor:   monitor,
		Fert:      nutrient,
		PH:        ph,
		ECIrrig:   ecIrrig,
		Mixer:     mixer,
		Plans:     plans,
		Doses:     doses,
		Allow:     allow,
		Cycle:     cycle,
		Stages:    stages,
		StagePlan: stagePlan,
		Tasks:     tasks,
		Calendar:  calendar,
		Quotas:    quotas,
		QuotaPlan: quotaPlan,
		Ledger:    ledger,
	}
}
