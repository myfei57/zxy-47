package shed

import (
	"fmt"
	"time"

	"greenhouse/internal/ns"
)

type Partition struct {
	ShedID     string    `json:"shed_id"`
	Zones      int       `json:"zones"`
	LightZones int       `json:"light_zones"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Layout struct {
	ShedID string `json:"shed_id"`
	Zones  []Zone `json:"zones"`
}

type Zone struct {
	Index     int    `json:"index"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	LightZone int    `json:"light_zone"`
}

func BuildLayout(shedID string, zones, lightZones int) Layout {
	layout := Layout{ShedID: shedID, Zones: make([]Zone, 0, zones)}
	for i := 1; i <= zones; i++ {
		lightZone := ns.BoundIndex(((i-1)*lightZones/zones)+1, 1, lightZones)
		layout.Zones = append(layout.Zones, Zone{
			Index:     i,
			ID:        zoneID(shedID, i),
			Name:      fmt.Sprintf("区%02d", i),
			LightZone: lightZone,
		})
	}
	return layout
}

func (l Layout) ZoneAt(index int) (Zone, bool) {
	if index < 1 || index > len(l.Zones) {
		return Zone{}, false
	}
	return l.Zones[index-1], true
}

func (l Layout) LightZoneCount() int {
	max := 0
	for _, z := range l.Zones {
		if z.LightZone > max {
			max = z.LightZone
		}
	}
	return max
}

func (l Layout) ZoneIDs() []string {
	out := make([]string, 0, len(l.Zones))
	for _, z := range l.Zones {
		out = append(out, z.ID)
	}
	return out
}

func (s *Service) LoadPartition(shedID string) (Partition, error) {
	var p Partition
	if err := s.kv.Load("partitions", shedID, &p); err != nil {
		return Partition{}, err
	}
	return p, nil
}

func (s *Service) RePartition(shedID string, zones, lightZones int) (Layout, error) {
	if zones < 1 || lightZones < 1 || lightZones > zones {
		return Layout{}, fmt.Errorf("shed: invalid partition %d/%d", zones, lightZones)
	}
	layout := BuildLayout(shedID, zones, lightZones)
	partition := Partition{ShedID: shedID, Zones: zones, LightZones: lightZones, UpdatedAt: time.Now()}
	if err := s.kv.Save("partitions", shedID, partition); err != nil {
		return Layout{}, err
	}
	return layout, nil
}

func zoneID(shedID string, index int) string {
	return fmt.Sprintf("%s-Z%02d", shedID, index)
}
