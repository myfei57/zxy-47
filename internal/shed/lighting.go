package shed

import (
	"fmt"
	"sort"
	"time"

	"greenhouse/internal/store"
)

type LightZone struct {
	Index     int       `json:"index"`
	ShedID    string    `json:"shed_id"`
	Name      string    `json:"name"`
	LampCount int       `json:"lamp_count"`
	Enabled   bool      `json:"enabled"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LightStore struct {
	kv *store.KeyValue
}

func NewLightStore(st *store.Store) *LightStore {
	return &LightStore{kv: store.NewKeyValue(st)}
}

func (ls *LightStore) Save(z LightZone) error {
	return ls.kv.Save("lights", keyOf(z.ShedID, z.Index), z)
}

func (ls *LightStore) List(shedID string) ([]LightZone, error) {
	keys, err := ls.kv.List("lights")
	if err != nil {
		return nil, err
	}
	out := []LightZone{}
	for _, key := range keys {
		var z LightZone
		if err := ls.kv.Load("lights", key, &z); err != nil {
			continue
		}
		if z.ShedID == shedID {
			out = append(out, z)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Index < out[j].Index })
	return out, nil
}

func (ls *LightStore) SetEnabled(shedID string, index int, enabled bool) error {
	var z LightZone
	if err := ls.kv.Load("lights", keyOf(shedID, index), &z); err != nil {
		return err
	}
	z.Enabled = enabled
	z.UpdatedAt = time.Now()
	return ls.Save(z)
}

func (ls *LightStore) Ensure(shedID string, count int) error {
	existing, err := ls.List(shedID)
	if err != nil {
		return err
	}
	have := map[int]bool{}
	for _, z := range existing {
		have[z.Index] = true
	}
	for i := 1; i <= count; i++ {
		if have[i] {
			continue
		}
		if err := ls.Save(LightZone{
			Index: i, ShedID: shedID, Name: fmt.Sprintf("补光区%d", i), LampCount: 8, Enabled: false, UpdatedAt: time.Now(),
		}); err != nil {
			return err
		}
	}
	return nil
}

func keyOf(shedID string, index int) string {
	return fmt.Sprintf("%s-L%02d", shedID, index)
}
