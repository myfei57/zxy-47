package vent

func (r *FilmRoller) Retry(shedID, cmdID, filmID, action string) (float64, error) {
	return r.Roll(shedID, cmdID, filmID, action)
}
