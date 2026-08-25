package audit

import (
	"time"

	"github.com/google/uuid"
)

type Record struct {
	ID     string    `json:"id"`
	ShedID string    `json:"shed_id"`
	Source string    `json:"source"`
	Action string    `json:"action"`
	Detail string    `json:"detail"`
	At     time.Time `json:"at"`
}

func NewRecord(shedID, source, action, detail string) Record {
	return Record{
		ID:     uuid.NewString(),
		ShedID: shedID,
		Source: source,
		Action: action,
		Detail: detail,
		At:     time.Now(),
	}
}
