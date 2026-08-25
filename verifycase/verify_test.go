package verifycase

import (
	"os"
	"path/filepath"
	"testing"

	"greenhouse/internal/console"
	"greenhouse/internal/env"
	"greenhouse/internal/fert"
	"greenhouse/internal/ns"
	"greenhouse/internal/shed"
	"greenhouse/internal/store"
	"greenhouse/internal/task"
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

func TestGhCropStageSwitchOrder(t *testing.T) {
	svc, dir := consoleServices(t)
	shedID := seed(t, svc, 8, 8)
	block := filepath.Join(dir, "stages", shedID)
	if err := os.MkdirAll(filepath.Dir(block), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(block, []byte("block"), 0o644); err != nil {
		t.Fatal(err)
	}
	stage := task.Stage{ShedID: shedID, Name: "开花坐果期", Label: "开花坐果期", EC: 2.6, PH: 6.4}
	if err := svc.Stages.Switch(stage); err == nil {
		t.Fatal("stage switch must fail when the stage label cannot be persisted")
	}
	ecController := fert.NewECController(fert.NewECStore(svc.Store), svc.Audit)
	ec, err := ecController.Target(shedID)
	if err == nil {
		if ec.Target == stage.EC {
			t.Fatalf("EC target changed to %v although the stage label failed to persist", ec.Target)
		}
	} else if err != store.ErrNotFound {
		t.Fatal(err)
	}
}
