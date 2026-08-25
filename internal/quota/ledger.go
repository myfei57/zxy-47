package quota

import (
	"encoding/json"
	"time"

	"greenhouse/internal/store"
)

type LedgerEntry struct {
	ShedID string    `json:"shed_id"`
	Cycle  int       `json:"cycle"`
	Amount float64   `json:"amount"`
	At     time.Time `json:"at"`
}

type Ledger struct {
	st *store.Store
}

func NewLedger(st *store.Store) *Ledger {
	return &Ledger{st: st}
}

func (l *Ledger) Append(shedID string, cycle int, amount float64) error {
	entry := LedgerEntry{ShedID: shedID, Cycle: cycle, Amount: amount, At: time.Now()}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return l.st.AppendLine("quota-ledger/"+shedID+".jsonl", string(data))
}

func (l *Ledger) Sum(shedID string, cycle int) (float64, error) {
	lines, err := l.st.ReadLines("quota-ledger/" + shedID + ".jsonl")
	if err != nil {
		return 0, err
	}
	total := 0.0
	for _, line := range lines {
		var entry LedgerEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry.Cycle == cycle {
			total += entry.Amount
		}
	}
	return total, nil
}
