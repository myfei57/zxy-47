package audit

import (
	"sort"
	"time"

	"greenhouse/internal/store"
)

type Service struct {
	records *store.RecordStore
}

func NewService(st *store.Store) *Service {
	return &Service{records: store.NewRecords(st)}
}

func (s *Service) Record(shedID, source, action, detail string) (Record, error) {
	rec := NewRecord(shedID, source, action, detail)
	if _, err := s.records.Append("executions", shedID, action, source+":"+detail); err != nil {
		return Record{}, err
	}
	return rec, nil
}

func (s *Service) Recent(shedID string, limit int) ([]Record, error) {
	rows, err := s.records.List("executions")
	if err != nil {
		return nil, err
	}
	out := []Record{}
	for _, row := range rows {
		if row.Subject == shedID {
			out = append(out, Record{
				ID:     row.ID,
				ShedID: row.Subject,
				Source: sourceOf(row.Detail),
				Action: row.Action,
				Detail: row.Detail,
				At:     row.At,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *Service) CountByAction(shedID, action string) (int, error) {
	return s.records.Count("executions", shedID, action)
}

func (s *Service) Since(shedID string, since time.Time) ([]Record, error) {
	rows, err := s.records.List("executions")
	if err != nil {
		return nil, err
	}
	out := []Record{}
	for _, row := range rows {
		if row.Subject == shedID && row.At.After(since) {
			out = append(out, Record{
				ID:     row.ID,
				ShedID: row.Subject,
				Source: sourceOf(row.Detail),
				Action: row.Action,
				Detail: row.Detail,
				At:     row.At,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	return out, nil
}

func sourceOf(detail string) string {
	for i := 0; i < len(detail); i++ {
		if detail[i] == ':' {
			return detail[:i]
		}
	}
	return "unknown"
}
