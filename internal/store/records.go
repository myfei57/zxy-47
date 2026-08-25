package store

import (
	"encoding/json"
	"time"
)

type LineRecord struct {
	ID      string    `json:"id"`
	Kind    string    `json:"kind"`
	Subject string    `json:"subject"`
	Action  string    `json:"action"`
	Detail  string    `json:"detail"`
	At      time.Time `json:"at"`
}

type RecordStore struct {
	st *Store
}

func NewRecords(st *Store) *RecordStore {
	return &RecordStore{st: st}
}

func (r *RecordStore) Append(kind, subject, action, detail string) (LineRecord, error) {
	rec := LineRecord{Kind: kind, Subject: subject, Action: action, Detail: detail, At: time.Now()}
	data, err := json.Marshal(rec)
	if err != nil {
		return LineRecord{}, err
	}
	if err := r.st.AppendLine("records/"+kind+".jsonl", string(data)); err != nil {
		return LineRecord{}, err
	}
	return rec, nil
}

func (r *RecordStore) List(kind string) ([]LineRecord, error) {
	lines, err := r.st.ReadLines("records/" + kind + ".jsonl")
	if err != nil {
		return nil, err
	}
	out := make([]LineRecord, 0, len(lines))
	for _, line := range lines {
		var rec LineRecord
		if err := json.Unmarshal([]byte(line), &rec); err == nil {
			out = append(out, rec)
		}
	}
	return out, nil
}

func (r *RecordStore) Count(kind, subject, action string) (int, error) {
	rows, err := r.List(kind)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, row := range rows {
		if row.Subject == subject && row.Action == action {
			n++
		}
	}
	return n, nil
}
