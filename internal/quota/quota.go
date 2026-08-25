package quota

import (
	"time"

	"greenhouse/internal/store"
)

type Quota struct {
	ShedID string    `json:"shed_id"`
	Cycle  int       `json:"cycle"`
	Daily  float64   `json:"daily"`
	Used   float64   `json:"used"`
	Remain float64   `json:"remain"`
	At     time.Time `json:"at"`
}

type QuotaStore struct {
	kv *store.KeyValue
}

func NewQuotaStore(st *store.Store) *QuotaStore {
	return &QuotaStore{kv: store.NewKeyValue(st)}
}

func (q *QuotaStore) Save(quota Quota) error {
	return q.kv.Save("quotas", quota.ShedID, quota)
}

func (q *QuotaStore) Get(shedID string) (Quota, error) {
	var quota Quota
	if err := q.kv.Load("quotas", shedID, &quota); err != nil {
		return Quota{}, err
	}
	return quota, nil
}

func (q *QuotaStore) NextCycle(shedID string) int {
	quota, err := q.Get(shedID)
	if err != nil {
		return 1
	}
	return quota.Cycle + 1
}
