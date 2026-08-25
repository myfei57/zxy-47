package health

import "time"

type Status struct {
	Status  string    `json:"status"`
	Version string    `json:"version"`
	Checked time.Time `json:"checked"`
}

func Check(version string) Status {
	return Status{Status: "ok", Version: version, Checked: time.Now()}
}
