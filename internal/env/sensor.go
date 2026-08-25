package env

import (
	"sort"
	"time"

	"github.com/google/uuid"

	"greenhouse/internal/store"
)

type SensorType string

const (
	TypeTemperature SensorType = "temperature"
	TypeHumidity    SensorType = "humidity"
	TypeSunlight    SensorType = "sunlight"
	TypeMoisture    SensorType = "moisture"
)

type Sensor struct {
	ID        string     `json:"id"`
	ShedID    string     `json:"shed_id"`
	ZoneID    string     `json:"zone_id"`
	Type      SensorType `json:"type"`
	Name      string     `json:"name"`
	Installed time.Time  `json:"installed"`
}

func NewSensor(shedID, zoneID string, kind SensorType, name string) Sensor {
	return Sensor{
		ID:        uuid.NewString(),
		ShedID:    shedID,
		ZoneID:    zoneID,
		Type:      kind,
		Name:      name,
		Installed: time.Now(),
	}
}

type Reading struct {
	SensorID string     `json:"sensor_id"`
	ShedID   string     `json:"shed_id"`
	ZoneID   string     `json:"zone_id"`
	Type     SensorType `json:"type"`
	Value    float64    `json:"value"`
	At       time.Time  `json:"at"`
}

type SensorRegistry struct {
	kv *store.KeyValue
}

func NewSensorRegistry(st *store.Store) *SensorRegistry {
	return &SensorRegistry{kv: store.NewKeyValue(st)}
}

func (r *SensorRegistry) Register(s Sensor) error {
	return r.kv.Save("sensors", s.ID, s)
}

func (r *SensorRegistry) Get(id string) (Sensor, error) {
	var s Sensor
	if err := r.kv.Load("sensors", id, &s); err != nil {
		return Sensor{}, err
	}
	return s, nil
}

func (r *SensorRegistry) ListByShed(shedID string) ([]Sensor, error) {
	ids, err := r.kv.List("sensors")
	if err != nil {
		return nil, err
	}
	out := []Sensor{}
	for _, id := range ids {
		s, err := r.Get(id)
		if err == nil && s.ShedID == shedID {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

type Sampler struct {
	registry *SensorRegistry
	kv       *store.KeyValue
	sun      *SunlightStore
	moisture *MoistureStore
	filter   *TemperatureFilter
}

func NewSampler(registry *SensorRegistry, st *store.Store, sun *SunlightStore, moisture *MoistureStore, filter *TemperatureFilter) *Sampler {
	return &Sampler{registry: registry, kv: store.NewKeyValue(st), sun: sun, moisture: moisture, filter: filter}
}

func (s *Sampler) SetReading(shedID string, kind SensorType, value float64) (Reading, error) {
	sensors, err := s.registry.ListByShed(shedID)
	if err != nil {
		return Reading{}, err
	}
	sensorID := ""
	zoneID := ""
	for _, sensor := range sensors {
		if sensor.Type == kind {
			sensorID = sensor.ID
			zoneID = sensor.ZoneID
			break
		}
	}
	reading := Reading{SensorID: sensorID, ShedID: shedID, ZoneID: zoneID, Type: kind, Value: value, At: time.Now()}
	if err := s.kv.Save("readings/"+string(kind), shedID, reading); err != nil {
		return Reading{}, err
	}
	if kind == TypeSunlight {
		if err := s.sun.Record(shedID, value); err != nil {
			return Reading{}, err
		}
	}
	if kind == TypeMoisture {
		if err := s.moisture.Record(shedID, value); err != nil {
			return Reading{}, err
		}
	}
	return reading, nil
}

func (s *Sampler) Latest(shedID string, kind SensorType) (Reading, error) {
	var reading Reading
	if err := s.kv.Load("readings/"+string(kind), shedID, &reading); err != nil {
		return Reading{}, err
	}
	return reading, nil
}

func (s *Sampler) Value(shedID string, kind SensorType, fallback float64) float64 {
	reading, err := s.Latest(shedID, kind)
	if err != nil {
		return fallback
	}
	return reading.Value
}

func (s *Sampler) AppendTemp(shedID string, value float64) error {
	history, err := s.TempHistory(shedID)
	if err != nil {
		history = []float64{}
	}
	history = s.filter.Smooth(history, value)
	return s.kv.Save("temperature-history", shedID, history)
}

func (s *Sampler) TempHistory(shedID string) ([]float64, error) {
	var history []float64
	if err := s.kv.Load("temperature-history", shedID, &history); err != nil {
		return nil, err
	}
	return history, nil
}
