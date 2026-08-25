package shed

import (
	"time"

	"github.com/google/uuid"
)

type Shed struct {
	ID         string    `json:"id"`
	SiteID     string    `json:"site_id"`
	Name       string    `json:"name"`
	Area       float64   `json:"area"`
	Zones      int       `json:"zones"`
	LightZones int       `json:"light_zones"`
	CreatedAt  time.Time `json:"created_at"`
}

func New(siteID, name string, area float64, zones, lightZones int) Shed {
	return Shed{
		ID:         uuid.NewString(),
		SiteID:     siteID,
		Name:       name,
		Area:       area,
		Zones:      zones,
		LightZones: lightZones,
		CreatedAt:  time.Now(),
	}
}

func (s Shed) EffectiveLightZones() int {
	if s.LightZones <= 0 {
		return s.Zones
	}
	if s.LightZones > s.Zones {
		return s.Zones
	}
	return s.LightZones
}

func (s Shed) ZoneIDs() []string {
	out := make([]string, 0, s.Zones)
	for i := 1; i <= s.Zones; i++ {
		out = append(out, zoneID(s.ID, i))
	}
	return out
}
