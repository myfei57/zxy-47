package curtain

import (
	"errors"
	"time"

	"greenhouse/internal/audit"
	"greenhouse/internal/env"
	"greenhouse/internal/store"
)

const (
	PositionExtended  = "extended"
	PositionRetracted = "retracted"
)

type Controller struct {
	states     *StateStore
	sun        *env.SunlightStore
	sampler    *env.Sampler
	thresholds *ThresholdStore
	lamps      *LampStateStore
	audit      *audit.Service
}

func NewController(st *store.Store, sun *env.SunlightStore, sampler *env.Sampler, audit *audit.Service) *Controller {
	return &Controller{
		states:     NewStateStore(st),
		sun:        sun,
		sampler:    sampler,
		thresholds: NewThresholdStore(st),
		lamps:      NewLampStateStore(st),
		audit:      audit,
	}
}

func (c *Controller) Retract(shedID, zoneID string) error {
	reading := c.sampler.Value(shedID, env.TypeSunlight, 0)
	if err := c.states.Save(State{ShedID: shedID, ZoneID: zoneID, Position: PositionRetracted, At: time.Now()}); err != nil {
		return err
	}
	if err := c.syncLamp(shedID, zoneID); err != nil {
		return err
	}
	if err := c.sun.Record(shedID, reading); err != nil {
		return err
	}
	_, err := c.audit.Record(shedID, "curtain", "retract", zoneID)
	return err
}

func (c *Controller) Extend(shedID, zoneID string) error {
	reading := c.sampler.Value(shedID, env.TypeSunlight, 0)
	if err := c.sun.Record(shedID, reading); err != nil {
		return err
	}
	if err := c.states.Save(State{ShedID: shedID, ZoneID: zoneID, Position: PositionExtended, At: time.Now()}); err != nil {
		return err
	}
	if err := c.syncLamp(shedID, zoneID); err != nil {
		return err
	}
	_, err := c.audit.Record(shedID, "curtain", "extend", zoneID)
	return err
}

func (c *Controller) Position(shedID, zoneID string) (State, error) {
	state, err := c.states.Get(zoneID)
	if err != nil {
		return State{}, err
	}
	if state.ShedID != shedID {
		return State{}, errors.New("curtain: zone does not belong to shed")
	}
	return state, nil
}

func (c *Controller) Thresholds(shedID string) Thresholds {
	return c.thresholds.LoadOrDefault(shedID)
}

func (c *Controller) SetThresholds(shedID string, th Thresholds) error {
	return c.thresholds.Save(shedID, th)
}

func (c *Controller) Lamp(shedID, zoneID string) (LampState, error) {
	state, err := c.lamps.Get(zoneID)
	if err != nil {
		return LampState{}, err
	}
	if state.ZoneID != zoneID {
		return LampState{}, errors.New("curtain: missing lamp state")
	}
	return state, nil
}

func (c *Controller) syncLamp(shedID, zoneID string) error {
	sun := c.sun.Value(shedID, 0)
	th := c.thresholds.LoadOrDefault(shedID)
	on := sun < th.LampSunlight
	reason := "sunlight_above"
	if on {
		reason = "sunlight_below"
	}
	return c.lamps.Save(LampState{ZoneID: zoneID, On: on, Reason: reason, At: time.Now()})
}
