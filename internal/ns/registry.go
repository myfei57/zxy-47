package ns

import (
	"sort"

	"greenhouse/internal/store"
)

type Registry struct {
	kv *store.KeyValue
}

func NewRegistry(st *store.Store) *Registry {
	return &Registry{kv: store.NewKeyValue(st)}
}

func (r *Registry) Save(site Site) error {
	return r.kv.Save("sites", site.ID, site)
}

func (r *Registry) Get(id string) (Site, error) {
	var site Site
	if err := r.kv.Load("sites", id, &site); err != nil {
		return Site{}, err
	}
	return site, nil
}

func (r *Registry) Delete(id string) error {
	return r.kv.Delete("sites", id)
}

func (r *Registry) List() ([]Site, error) {
	ids, err := r.kv.List("sites")
	if err != nil {
		return nil, err
	}
	out := make([]Site, 0, len(ids))
	for _, id := range ids {
		site, err := r.Get(id)
		if err == nil {
			out = append(out, site)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (r *Registry) Count() (int, error) {
	ids, err := r.kv.List("sites")
	if err != nil {
		return 0, err
	}
	return len(ids), nil
}
