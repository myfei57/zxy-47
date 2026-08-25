package verifycase

import (
	"strings"
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

func TestGhSensorZoneMappingFresh(t *testing.T) {
	svc, _ := consoleServices(t)
	shedID := seed(t, svc, 8, 8)
	if _, err := svc.Sheds.RePartition(shedID, 6, 6); err != nil {
		t.Fatal(err)
	}
	binding, err := svc.Bindings.Load(shedID)
	if err != nil {
		t.Fatal(err)
	}
	if len(binding.ZoneIDs) != 6 {
		t.Fatalf("sensor bindings must follow the new partition, got %v", binding.ZoneIDs)
	}
	for _, zone := range binding.ZoneIDs {
		if strings.HasSuffix(zone, "-Z08") {
			t.Fatalf("stale zone %s remains bound after re-partition", zone)
		}
	}
}
