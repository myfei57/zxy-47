package shed

import (
	"sort"
	"time"

	"greenhouse/internal/env"
	"greenhouse/internal/store"
)

type Service struct {
	kv    *store.KeyValue
	env   *env.BindingService
	light *LightStore
}

func NewService(st *store.Store, bindings *env.BindingService) *Service {
	return &Service{kv: store.NewKeyValue(st), env: bindings, light: NewLightStore(st)}
}

func (s *Service) Register(sh Shed) error {
	if err := s.kv.Save("sheds", sh.ID, sh); err != nil {
		return err
	}
	layout := BuildLayout(sh.ID, sh.Zones, sh.EffectiveLightZones())
	partition := Partition{ShedID: sh.ID, Zones: sh.Zones, LightZones: sh.EffectiveLightZones(), UpdatedAt: time.Now()}
	if err := s.kv.Save("partitions", sh.ID, partition); err != nil {
		return err
	}
	if err := s.env.RefreshBindings(sh.ID, layout.ZoneIDs()); err != nil {
		return err
	}
	return s.light.Ensure(sh.ID, sh.EffectiveLightZones())
}

func (s *Service) Get(id string) (Shed, error) {
	var sh Shed
	if err := s.kv.Load("sheds", id, &sh); err != nil {
		return Shed{}, err
	}
	return sh, nil
}

func (s *Service) List(siteID string) ([]Shed, error) {
	ids, err := s.kv.List("sheds")
	if err != nil {
		return nil, err
	}
	out := []Shed{}
	for _, id := range ids {
		sh, err := s.Get(id)
		if err == nil && (siteID == "" || sh.SiteID == siteID) {
			out = append(out, sh)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *Service) Summary(sh Shed) map[string]any {
	return map[string]any{
		"id":          sh.ID,
		"name":        sh.Name,
		"area":        sh.Area,
		"zones":       sh.Zones,
		"light_zones": sh.EffectiveLightZones(),
		"created_at":  sh.CreatedAt.Format(time.RFC3339),
	}
}
