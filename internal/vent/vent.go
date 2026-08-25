package vent

import (
	"sort"
	"time"

	"greenhouse/internal/store"
)

const (
	SideWindward = "windward"
	SideLeeward  = "leeward"
)

type Vent struct {
	ID      string    `json:"id"`
	ShedID  string    `json:"shed_id"`
	Side    string    `json:"side"`
	Name    string    `json:"name"`
	Open    bool      `json:"open"`
	Updated time.Time `json:"updated_at"`
}

type VentStore struct {
	kv *store.KeyValue
}

func NewVentStore(st *store.Store) *VentStore {
	return &VentStore{kv: store.NewKeyValue(st)}
}

func (v *VentStore) Save(vent Vent) error {
	return v.kv.Save("vents", vent.ID, vent)
}

func (v *VentStore) Get(id string) (Vent, error) {
	var vent Vent
	if err := v.kv.Load("vents", id, &vent); err != nil {
		return Vent{}, err
	}
	return vent, nil
}

func (v *VentStore) List(shedID string) ([]Vent, error) {
	ids, err := v.kv.List("vents")
	if err != nil {
		return nil, err
	}
	out := []Vent{}
	for _, id := range ids {
		vent, err := v.Get(id)
		if err == nil && vent.ShedID == shedID {
			out = append(out, vent)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (v *VentStore) SetOpen(id string, open bool) error {
	vent, err := v.Get(id)
	if err != nil {
		return err
	}
	vent.Open = open
	vent.Updated = time.Now()
	return v.Save(vent)
}
