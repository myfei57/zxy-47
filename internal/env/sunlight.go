package env

import (
	"time"

	"greenhouse/internal/store"
)

type SunlightStore struct {
	kv *store.KeyValue
}

func NewSunlightStore(st *store.Store) *SunlightStore {
	return &SunlightStore{kv: store.NewKeyValue(st)}
}

func (s *SunlightStore) Record(shedID string, value float64) error {
	doc := Reading{ShedID: shedID, Type: TypeSunlight, Value: value, At: time.Now()}
	return s.kv.Save("sunlight", shedID, doc)
}

func (s *SunlightStore) Read(shedID string) (Reading, error) {
	var doc Reading
	if err := s.kv.Load("sunlight", shedID, &doc); err != nil {
		return Reading{}, err
	}
	return doc, nil
}

func (s *SunlightStore) Value(shedID string, fallback float64) float64 {
	doc, err := s.Read(shedID)
	if err != nil {
		return fallback
	}
	return doc.Value
}
