package env

import (
	"sort"
	"time"

	"greenhouse/internal/store"
)

type Binding struct {
	ShedID    string            `json:"shed_id"`
	ZoneIDs   []string          `json:"zone_ids"`
	BySensor  map[string]string `json:"by_sensor"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type BindingService struct {
	kv  *store.KeyValue
	reg *SensorRegistry
}

func NewBindingService(st *store.Store, reg *SensorRegistry) *BindingService {
	return &BindingService{kv: store.NewKeyValue(st), reg: reg}
}

func (b *BindingService) RefreshBindings(shedID string, zoneIDs []string) error {
	sensors, err := b.reg.ListByShed(shedID)
	if err != nil {
		return err
	}
	valid := map[string]bool{}
	for _, zoneID := range zoneIDs {
		valid[zoneID] = true
	}
	fallback := ""
	if len(zoneIDs) > 0 {
		fallback = zoneIDs[len(zoneIDs)-1]
	}
	bySensor := map[string]string{}
	for _, sensor := range sensors {
		zone := sensor.ZoneID
		if !valid[zone] {
			zone = fallback
		}
		bySensor[sensor.ID] = zone
	}
	doc := Binding{ShedID: shedID, ZoneIDs: zoneIDs, BySensor: bySensor, UpdatedAt: time.Now()}
	return b.kv.Save("bindings", shedID, doc)
}

func (b *BindingService) Load(shedID string) (Binding, error) {
	var doc Binding
	if err := b.kv.Load("bindings", shedID, &doc); err != nil {
		return Binding{}, err
	}
	return doc, nil
}

func (b *BindingService) ZoneOf(shedID, sensorID string) (string, error) {
	doc, err := b.Load(shedID)
	if err != nil {
		return "", err
	}
	zone := doc.BySensor[sensorID]
	if zone == "" {
		return "", store.ErrNotFound
	}
	return zone, nil
}

func (b *BindingService) Zones(shedID string) ([]string, error) {
	doc, err := b.Load(shedID)
	if err != nil {
		return nil, err
	}
	out := append([]string{}, doc.ZoneIDs...)
	sort.Strings(out)
	return out, nil
}
