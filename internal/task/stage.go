package task

import (
	"errors"
	"time"

	"greenhouse/internal/audit"
	"greenhouse/internal/fert"
	"greenhouse/internal/store"
)

type Stage struct {
	ID     string    `json:"id"`
	ShedID string    `json:"shed_id"`
	Name   string    `json:"name"`
	Label  string    `json:"label"`
	EC     float64   `json:"ec"`
	PH     float64   `json:"ph"`
	At     time.Time `json:"at"`
}

type StageStore struct {
	kv *store.KeyValue
}

func NewStageStore(st *store.Store) *StageStore {
	return &StageStore{kv: store.NewKeyValue(st)}
}

func (s *StageStore) path(shedID string) string {
	return "stages/" + shedID + "/stage.json"
}

func (s *StageStore) Save(stage Stage) error {
	return s.kv.Store().WriteJSON(s.path(stage.ShedID), stage)
}

func (s *StageStore) Get(shedID string) (Stage, error) {
	var stage Stage
	if err := s.kv.Store().ReadJSON(s.path(shedID), &stage); err != nil {
		return Stage{}, err
	}
	return stage, nil
}

type StageService struct {
	stages *StageStore
	ec     *fert.ECController
	audit  *audit.Service
}

func NewStageService(st *store.Store, ec *fert.ECController, audit *audit.Service) *StageService {
	return &StageService{
		stages: NewStageStore(st),
		ec:     ec,
		audit:  audit,
	}
}

func (s *StageService) Switch(stage Stage) error {
	if stage.Label == "" {
		return errors.New("task: empty stage label")
	}
	stage.ID = stage.ShedID
	stage.At = time.Now()
	if _, err := s.ec.SetTarget(stage.ShedID, stage.EC); err != nil {
		return err
	}
	if err := s.stages.Save(stage); err != nil {
		return err
	}
	_, err := s.audit.Record(stage.ShedID, "task", "stage", stage.Label)
	return err
}

func (s *StageService) Current(shedID string) (Stage, error) {
	return s.stages.Get(shedID)
}
