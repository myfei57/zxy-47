package vent

import (
	"sort"
	"time"

	"greenhouse/internal/audit"
	"greenhouse/internal/store"
)

type Film struct {
	ID        string    `json:"id"`
	ShedID    string    `json:"shed_id"`
	Name      string    `json:"name"`
	Position  float64   `json:"position"`
	Max       float64   `json:"max"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FilmStore struct {
	kv *store.KeyValue
}

func NewFilmStore(st *store.Store) *FilmStore {
	return &FilmStore{kv: store.NewKeyValue(st)}
}

func (f *FilmStore) Save(film Film) error {
	return f.kv.Save("films", film.ID, film)
}

func (f *FilmStore) Get(id string) (Film, error) {
	var film Film
	if err := f.kv.Load("films", id, &film); err != nil {
		return Film{}, err
	}
	return film, nil
}

func (f *FilmStore) List(shedID string) ([]Film, error) {
	ids, err := f.kv.List("films")
	if err != nil {
		return nil, err
	}
	out := []Film{}
	for _, id := range ids {
		film, err := f.Get(id)
		if err == nil && film.ShedID == shedID {
			out = append(out, film)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *FilmRoller) checkLimits(shedID string) error {
	films, err := r.films.List(shedID)
	if err != nil {
		return err
	}
	for _, film := range films {
		if film.Position > film.Max {
			if _, err := r.alarms.Raise(shedID, "warning", "film "+film.ID+" past limit "+film.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

type FilmRoller struct {
	films  *FilmStore
	marks  *store.MarkStore
	audit  *audit.Service
	alarms *audit.AlarmService
}

func NewFilmRoller(st *store.Store, audit *audit.Service, alarms *audit.AlarmService) *FilmRoller {
	return &FilmRoller{
		films:  NewFilmStore(st),
		marks:  store.NewMarks(st),
		audit:  audit,
		alarms: alarms,
	}
}

func (r *FilmRoller) Register(shedID, filmID, name string, max float64) error {
	return r.films.Save(Film{ID: filmID, ShedID: shedID, Name: name, Max: max, UpdatedAt: time.Now()})
}

func (r *FilmRoller) Roll(shedID, cmdID, filmID, action string) (float64, error) {
	film, err := r.films.Get(filmID)
	if err != nil {
		return 0, err
	}
	if action == "open" {
		film.Position += 1.0
	} else if action == "close" {
		film.Position -= 1.0
		if film.Position < 0 {
			film.Position = 0
		}
	}
	film.UpdatedAt = time.Now()
	if err := r.films.Save(film); err != nil {
		return 0, err
	}
	if err := r.marks.Set("film", cmdID, action); err != nil {
		return 0, err
	}
	if _, err := r.audit.Record(shedID, "vent", "film", filmID+":"+action); err != nil {
		return 0, err
	}
	return film.Position, nil
}

func (r *FilmRoller) Films(shedID string) ([]Film, error) {
	if err := r.checkLimits(shedID); err != nil {
		return nil, err
	}
	return r.films.List(shedID)
}

func (r *FilmRoller) ExecutedCommands() (map[string]string, error) {
	keys, err := r.marks.Keys("film")
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, key := range keys {
		payload, err := r.marks.Payload("film", key)
		if err != nil {
			continue
		}
		out[key] = payload
	}
	return out, nil
}

func (r *FilmRoller) ResetCommands() error {
	keys, err := r.marks.Keys("film")
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err := r.marks.Clear("film", key); err != nil {
			return err
		}
	}
	return nil
}
