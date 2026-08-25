package vent

import (
	"sort"
	"time"

	"greenhouse/internal/audit"
	"greenhouse/internal/env"
	"greenhouse/internal/store"
)

type Fan struct {
	ID        string    `json:"id"`
	ShedID    string    `json:"shed_id"`
	Name      string    `json:"name"`
	Running   bool      `json:"running"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FanStore struct {
	kv *store.KeyValue
}

func NewFanStore(st *store.Store) *FanStore {
	return &FanStore{kv: store.NewKeyValue(st)}
}

func (f *FanStore) Save(fan Fan) error {
	return f.kv.Save("fans", fan.ID, fan)
}

func (f *FanStore) Get(id string) (Fan, error) {
	var fan Fan
	if err := f.kv.Load("fans", id, &fan); err != nil {
		return Fan{}, err
	}
	return fan, nil
}

func (f *FanStore) List(shedID string) ([]Fan, error) {
	ids, err := f.kv.List("fans")
	if err != nil {
		return nil, err
	}
	out := []Fan{}
	for _, id := range ids {
		fan, err := f.Get(id)
		if err == nil && fan.ShedID == shedID {
			out = append(out, fan)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

type FanController struct {
	fans    *FanStore
	filter  *env.TemperatureFilter
	params  *store.ParamsStore
	audit   *audit.Service
	alarms  *audit.AlarmService
}

func NewFanController(st *store.Store, filter *env.TemperatureFilter, audit *audit.Service, alarms *audit.AlarmService) *FanController {
	return &FanController{
		fans:   NewFanStore(st),
		filter: filter,
		params: store.NewParams(st),
		audit:  audit,
		alarms: alarms,
	}
}

func (c *FanController) Register(shedID string) error {
	fans := []Fan{
		{ID: shedID + "-F1", ShedID: shedID, Name: "环流风机"},
		{ID: shedID + "-F2", ShedID: shedID, Name: "负压风机"},
	}
	for _, fan := range fans {
		if err := c.fans.Save(fan); err != nil {
			return err
		}
	}
	return nil
}

func (c *FanController) Evaluate(shedID string, samples []float64) error {
	filtered := c.filter.LastAverage(samples)
	upper := c.upperBound(shedID)
	run := filtered > upper
	fans, err := c.fans.List(shedID)
	if err != nil {
		return err
	}
	for _, fan := range fans {
		if fan.Running != run {
			fan.Running = run
			fan.UpdatedAt = time.Now()
			if err := c.fans.Save(fan); err != nil {
				return err
			}
		}
	}
	state := "off"
	if run {
		state = "on"
	}
	if env.RawMax(samples) > upper+2 {
		if _, err := c.alarms.Raise(shedID, "warning", "temperature spike over bound"); err != nil {
			return err
		}
	}
	_, err = c.audit.Record(shedID, "vent", "fan", state)
	return err
}

func (c *FanController) SetUpperBound(shedID string, value float64) error {
	return c.params.Save("climate", shedID, map[string]any{"fan_upper": value})
}

func (c *FanController) upperBound(shedID string) float64 {
	raw, err := c.params.Load("climate", shedID)
	if err != nil {
		return 32
	}
	if value, ok := raw["fan_upper"].(float64); ok {
		return value
	}
	return 32
}

func (c *FanController) Fans(shedID string) ([]Fan, error) {
	return c.fans.List(shedID)
}

func (c *FanController) FilterWindow() int {
	return c.filter.Window()
}
