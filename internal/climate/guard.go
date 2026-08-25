package climate

import (
	"greenhouse/internal/store"
)

type GuardParams struct {
	Wind        string  `json:"wind"`
	TempHigh    float64 `json:"temp_high"`
	TempLow     float64 `json:"temp_low"`
	SunHigh     float64 `json:"sun_high"`
	MoistureLow float64 `json:"moisture_low"`
}

type GuardStore struct {
	params *store.ParamsStore
}

func NewGuardStore(st *store.Store) *GuardStore {
	return &GuardStore{params: store.NewParams(st)}
}

func DefaultGuard() GuardParams {
	return GuardParams{Wind: "north", TempHigh: 32, TempLow: 12, SunHigh: 800, MoistureLow: 30}
}

func (g *GuardStore) Save(shedID string, params GuardParams) error {
	return g.params.Save("climate-guard", shedID, map[string]any{
		"wind":         params.Wind,
		"temp_high":    params.TempHigh,
		"temp_low":     params.TempLow,
		"sun_high":     params.SunHigh,
		"moisture_low": params.MoistureLow,
	})
}

func (g *GuardStore) Load(shedID string) (GuardParams, error) {
	raw, err := g.params.Load("climate-guard", shedID)
	if err != nil {
		return GuardParams{}, err
	}
	params := DefaultGuard()
	if value, ok := raw["wind"].(string); ok {
		params.Wind = value
	}
	if value, ok := raw["temp_high"].(float64); ok {
		params.TempHigh = value
	}
	if value, ok := raw["temp_low"].(float64); ok {
		params.TempLow = value
	}
	if value, ok := raw["sun_high"].(float64); ok {
		params.SunHigh = value
	}
	if value, ok := raw["moisture_low"].(float64); ok {
		params.MoistureLow = value
	}
	return params, nil
}

func (g *GuardStore) LoadOrDefault(shedID string) GuardParams {
	params, err := g.Load(shedID)
	if err != nil {
		return DefaultGuard()
	}
	return params
}

func (g *GuardStore) Wind(shedID string) string {
	return g.LoadOrDefault(shedID).Wind
}
