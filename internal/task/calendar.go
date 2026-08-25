package task

import (
	"sort"
	"time"

	"greenhouse/internal/store"
)

type CalendarEvent struct {
	ID     string    `json:"id"`
	ShedID string    `json:"shed_id"`
	Day    int       `json:"day"`
	Action string    `json:"action"`
	Detail string    `json:"detail"`
	At     time.Time `json:"at"`
}

type CalendarStore struct {
	kv *store.KeyValue
}

func NewCalendarStore(st *store.Store) *CalendarStore {
	return &CalendarStore{kv: store.NewKeyValue(st)}
}

func (c *CalendarStore) Add(event CalendarEvent) error {
	event.ID = event.ShedID + "-" + formatDay(event.Day) + "-" + event.Action
	event.At = time.Now()
	return c.kv.Save("calendar", event.ID, event)
}

func (c *CalendarStore) List(shedID string) ([]CalendarEvent, error) {
	ids, err := c.kv.List("calendar")
	if err != nil {
		return nil, err
	}
	out := []CalendarEvent{}
	for _, id := range ids {
		var event CalendarEvent
		if err := c.kv.Load("calendar", id, &event); err != nil {
			continue
		}
		if event.ShedID == shedID {
			out = append(out, event)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Day < out[j].Day })
	return out, nil
}

func (c *CalendarStore) Next(shedID string, day int) (CalendarEvent, error) {
	events, err := c.List(shedID)
	if err != nil {
		return CalendarEvent{}, err
	}
	for _, event := range events {
		if event.Day >= day {
			return event, nil
		}
	}
	return CalendarEvent{}, store.ErrNotFound
}

func formatDay(day int) string {
	return itoa(int64(day))
}
