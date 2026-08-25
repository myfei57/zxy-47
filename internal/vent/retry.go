package vent

func (r *FilmRoller) Retry(shedID, cmdID, filmID, action string) (float64, error) {
	if r.marks.Has("film", cmdID) {
		film, err := r.films.Get(filmID)
		if err != nil {
			return 0, err
		}
		return film.Position, nil
	}
	return r.Roll(shedID, cmdID, filmID, action)
}
