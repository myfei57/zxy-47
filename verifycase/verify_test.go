package verifycase

import (
	"testing"

	"greenhouse/internal/console"
	"greenhouse/internal/env"
	"greenhouse/internal/irrig"
	"greenhouse/internal/ns"
	"greenhouse/internal/shed"
	"greenhouse/internal/store"
)

func consoleServices(t *testing.T) (*console.Services, string) {
	t.Helper()
	dir := t.TempDir()
	return console.Build(store.New(dir)), dir
}

func seed(t *testing.T, svc *console.Services, zones, lightZones int) string {
	t.Helper()
	site := ns.NewSite("test", "local")
	if err := svc.Sites.Save(site); err != nil {
		t.Fatal(err)
	}
	sh := shed.New(site.ID, "shed", 100, zones, lightZones)
	if err := svc.Sheds.Register(sh); err != nil {
		t.Fatal(err)
	}
	if err := svc.Vent.Register(sh.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.Fans.Register(sh.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.Roller.Register(sh.ID, sh.ID+"-film-east", "east", 5); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Allow.Reset(sh.ID, 500); err != nil {
		t.Fatal(err)
	}
	zone := sh.ID + "-Z01"
	for _, seed := range []struct {
		kind env.SensorType
		name string
	}{
		{env.TypeTemperature, "temp"},
		{env.TypeHumidity, "humidity"},
		{env.TypeSunlight, "sun"},
		{env.TypeMoisture, "moisture"},
	} {
		if err := svc.Sensors.Register(env.NewSensor(sh.ID, zone, seed.kind, seed.name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := svc.Bindings.RefreshBindings(sh.ID, sh.ZoneIDs()); err != nil {
		t.Fatal(err)
	}
	return sh.ID
}

func TestGhFertMixOrder(t *testing.T) {
	svc, _ := consoleServices(t)
	shedID := seed(t, svc, 8, 8)
	plan := irrig.Plan{
		ID:               "p1",
		ShedID:           shedID,
		Name:             "fert",
		BaselineMoisture: 15,
		Threshold:        50,
		DoseWater:        100,
		Enabled:          true,
	}
	if err := svc.Plans.Save(plan); err != nil {
		t.Fatal(err)
	}
	if err := svc.Moisture.Record(shedID, 10); err != nil {
		t.Fatal(err)
	}
	result, err := svc.Cycle.Run(shedID, plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Mix.Steps) != 2 {
		t.Fatalf("expected a two-step mix, got %v", result.Mix.Steps)
	}
	if result.Mix.Steps[0].Kind != "water" || result.Mix.Steps[1].Kind != "concentrate" {
		t.Fatalf("dilution water must be added before the concentrate, got %s then %s",
			result.Mix.Steps[0].Kind, result.Mix.Steps[1].Kind)
	}
}
