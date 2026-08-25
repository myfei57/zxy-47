package irrig

import (
	"time"

	"greenhouse/internal/fert"
	"greenhouse/internal/store"
)

type Dose struct {
	ShedID      string    `json:"shed_id"`
	Water       float64   `json:"water"`
	Concentrate float64   `json:"concentrate"`
	At          time.Time `json:"at"`
}

type DoseService struct {
	recipes *fert.RecipeStore
	kv      *store.KeyValue
}

func NewDoseService(recipes *fert.RecipeStore, st *store.Store) *DoseService {
	return &DoseService{recipes: recipes, kv: store.NewKeyValue(st)}
}

func (d *DoseService) Compute(shedID string) (Dose, error) {
	recipe := d.recipes.LoadOrDefault(shedID)
	dose := Dose{
		ShedID:      shedID,
		Water:       recipe.WaterPerDose,
		Concentrate: recipe.ConcentrateA + recipe.ConcentrateB,
		At:          time.Now(),
	}
	if err := d.kv.Save("doses", shedID, dose); err != nil {
		return Dose{}, err
	}
	return dose, nil
}

func (d *DoseService) Last(shedID string) (Dose, error) {
	var dose Dose
	if err := d.kv.Load("doses", shedID, &dose); err != nil {
		return Dose{}, err
	}
	return dose, nil
}
