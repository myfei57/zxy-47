package env

import (
	"time"

	"greenhouse/internal/store"
)

type MoistureStore struct {
	kv *store.KeyValue
}

func NewMoistureStore(st *store.Store) *MoistureStore {
	return &MoistureStore{kv: store.NewKeyValue(st)}
}

func (m *MoistureStore) Record(shedID string, value float64) error {
	doc := Reading{ShedID: shedID, Type: TypeMoisture, Value: value, At: time.Now()}
	return m.kv.Save("moisture", shedID, doc)
}

func (m *MoistureStore) Read(shedID string) (float64, error) {
	var doc Reading
	if err := m.kv.Load("moisture", shedID, &doc); err != nil {
		return 0, err
	}
	return doc.Value, nil
}

func (m *MoistureStore) Value(shedID string, fallback float64) float64 {
	value, err := m.Read(shedID)
	if err != nil {
		return fallback
	}
	return value
}
