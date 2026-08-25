package audit

import (
	"sort"
	"time"

	"github.com/google/uuid"

	"greenhouse/internal/store"
)

type Alarm struct {
	ID       string    `json:"id"`
	ShedID   string    `json:"shed_id"`
	Severity string    `json:"severity"`
	Message  string    `json:"message"`
	Acked    bool      `json:"acked"`
	At       time.Time `json:"at"`
}

type AlarmService struct {
	kv *store.KeyValue
}

func NewAlarmService(st *store.Store) *AlarmService {
	return &AlarmService{kv: store.NewKeyValue(st)}
}

func (a *AlarmService) Raise(shedID, severity, message string) (Alarm, error) {
	alarm := Alarm{
		ID:       uuid.NewString(),
		ShedID:   shedID,
		Severity: severity,
		Message:  message,
		At:       time.Now(),
	}
	if err := a.kv.Save("alarms", alarm.ID, alarm); err != nil {
		return Alarm{}, err
	}
	return alarm, nil
}

func (a *AlarmService) Ack(id string) error {
	var alarm Alarm
	if err := a.kv.Load("alarms", id, &alarm); err != nil {
		return err
	}
	alarm.Acked = true
	return a.kv.Save("alarms", id, alarm)
}

func (a *AlarmService) List(shedID string, limit int) ([]Alarm, error) {
	ids, err := a.kv.List("alarms")
	if err != nil {
		return nil, err
	}
	out := []Alarm{}
	for _, id := range ids {
		var alarm Alarm
		if err := a.kv.Load("alarms", id, &alarm); err != nil {
			continue
		}
		if shedID == "" || alarm.ShedID == shedID {
			out = append(out, alarm)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (a *AlarmService) Unacked(shedID string) (int, error) {
	rows, err := a.List(shedID, 0)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, row := range rows {
		if !row.Acked {
			n++
		}
	}
	return n, nil
}
