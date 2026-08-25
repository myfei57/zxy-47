package ns

import (
	"time"

	"github.com/google/uuid"
)

type Site struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
}

func NewSite(name, location string) Site {
	return Site{ID: uuid.NewString(), Name: name, Location: location, CreatedAt: time.Now()}
}

func (s Site) DisplayName() string {
	if s.Name == "" {
		return s.ID
	}
	return s.Name
}
