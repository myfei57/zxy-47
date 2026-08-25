package fert

import (
	"time"

	"greenhouse/internal/store"
)

type Recipe struct {
	ShedID       string    `json:"shed_id"`
	ConcentrateA float64   `json:"concentrate_a"`
	ConcentrateB float64   `json:"concentrate_b"`
	WaterPerDose float64   `json:"water_per_dose"`
	TargetEC     float64   `json:"target_ec"`
	TargetPH     float64   `json:"target_ph"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RecipeStore struct {
	kv *store.KeyValue
}

func NewRecipeStore(st *store.Store) *RecipeStore {
	return &RecipeStore{kv: store.NewKeyValue(st)}
}

func DefaultRecipe(shedID string) Recipe {
	return Recipe{
		ShedID:       shedID,
		ConcentrateA: 1.5,
		ConcentrateB: 0.5,
		WaterPerDose: 100,
		TargetEC:     2.2,
		TargetPH:     6.2,
		UpdatedAt:    time.Now(),
	}
}

func (r *RecipeStore) Save(recipe Recipe) error {
	return r.kv.Save("recipes", recipe.ShedID, recipe)
}

func (r *RecipeStore) Get(shedID string) (Recipe, error) {
	var recipe Recipe
	if err := r.kv.Load("recipes", shedID, &recipe); err != nil {
		return Recipe{}, err
	}
	return recipe, nil
}

func (r *RecipeStore) LoadOrDefault(shedID string) Recipe {
	recipe, err := r.Get(shedID)
	if err != nil {
		return DefaultRecipe(shedID)
	}
	return recipe
}
