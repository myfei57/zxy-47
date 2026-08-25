package climate

import (
	"errors"
	"time"

	"greenhouse/internal/audit"
	"greenhouse/internal/store"
)

type Service struct {
	modes *ModeStore
	audit *audit.Service
}

func NewService(st *store.Store, audit *audit.Service) *Service {
	return &Service{modes: NewModeStore(st), audit: audit}
}

func (s *Service) Switch(shedID, mode string) (ModeState, error) {
	if mode != ModeHeat && mode != ModeAuto && mode != ModeVent {
		return ModeState{}, errors.New("climate: unknown mode " + mode)
	}
	state, err := s.modes.Get(shedID)
	if err != nil {
		state = ModeState{ShedID: shedID, Mode: mode, Since: time.Now()}
	} else {
		if state.Mode == mode {
			return state, nil
		}
		state.Mode = mode
		state.Since = time.Now()
	}
	if err := s.modes.Save(state); err != nil {
		return ModeState{}, err
	}
	_, err = s.audit.Record(shedID, "climate", "switch", mode)
	return state, err
}

func (s *Service) Current(shedID string) (ModeState, error) {
	state, err := s.modes.Get(shedID)
	if err != nil {
		return s.modes.Default(shedID), nil
	}
	return state, nil
}
