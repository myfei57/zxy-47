package vent

import (
	"sort"

	"greenhouse/internal/audit"
	"greenhouse/internal/store"
)

type SequenceStore struct {
	st *store.Store
}

func NewSequenceStore(st *store.Store) *SequenceStore {
	return &SequenceStore{st: st}
}

func (s *SequenceStore) Append(shedID, step string) error {
	return s.st.AppendLine("vent-sequences/"+shedID+".jsonl", step)
}

func (s *SequenceStore) List(shedID string) ([]string, error) {
	return s.st.ReadLines("vent-sequences/" + shedID + ".jsonl")
}

type Controller struct {
	vents *VentStore
	seq   *SequenceStore
	audit *audit.Service
}

func NewController(st *store.Store, audit *audit.Service) *Controller {
	return &Controller{
		vents: NewVentStore(st),
		seq:   NewSequenceStore(st),
		audit: audit,
	}
}

func (c *Controller) Register(shedID string) error {
	vents := []Vent{
		{ID: shedID + "-V1", ShedID: shedID, Side: SideWindward, Name: "上风侧窗"},
		{ID: shedID + "-V2", ShedID: shedID, Side: SideLeeward, Name: "下风侧窗"},
	}
	for _, vent := range vents {
		if err := c.vents.Save(vent); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) OpenForCooling(shedID, wind string) error {
	vents, err := c.vents.List(shedID)
	if err != nil {
		return err
	}
	if len(vents) < 2 {
		return store.ErrNotFound
	}
	sort.Slice(vents, func(i, j int) bool { return vents[i].ID < vents[j].ID })
	windward, leeward := vents[0], vents[1]
	if wind != "north" {
		windward, leeward = vents[1], vents[0]
	}
	// 夏季降温开侧窗的顺序：先开上风侧，再开下风侧。
	// 若反过来先开下风，风会从门口灌入，吹蔫靠门的苗。
	order := []Vent{windward, leeward}
	for _, vent := range order {
		if err := c.vents.SetOpen(vent.ID, true); err != nil {
			return err
		}
		if err := c.seq.Append(shedID, vent.Side); err != nil {
			return err
		}
	}
	_, err = c.audit.Record(shedID, "vent", "open", wind)
	return err
}

func (c *Controller) Sequence(shedID string) ([]string, error) {
	return c.seq.List(shedID)
}

func (c *Controller) List(shedID string) ([]Vent, error) {
	return c.vents.List(shedID)
}
