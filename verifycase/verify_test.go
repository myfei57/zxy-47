package verifycase

import (
	"testing"

	"greenhouse/internal/console"
	"greenhouse/internal/env"
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
	if _, err := svc.Allow.Reset(sh.ID, 100); err != nil {
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

func TestGhIrrigQuotaCycleReset(t *testing.T) {
	svc, _ := consoleServices(t)
	shedID := seed(t, svc, 8, 8)
	if _, err := svc.Allow.Reset(shedID, 100); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Allow.Consume(shedID, 30); err != nil {
		t.Fatal(err)
	}
	state, err := svc.Quotas.Release(shedID)
	if err != nil {
		t.Fatal(err)
	}
	if state.Remain != 100 {
		t.Fatalf("each quota cycle must start from the fresh allowance, remain=%v", state.Remain)
	}
}
