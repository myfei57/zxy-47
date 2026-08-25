package fert

import (
	"greenhouse/internal/audit"
)

type ECCorrector interface {
	Correct(shedID string, target float64) (ECState, error)
	Current(shedID string) (ECState, error)
}

type AdjustResult struct {
	PH  PHResult  `json:"ph"`
	EC  ECState   `json:"ec"`
	Mix MixResult `json:"mix"`
}

type NutrientService struct {
	ph      *PHController
	ec      ECCorrector
	mixer   *Mixer
	recipes *RecipeStore
	audit   *audit.Service
}

func NewNutrientService(
	ph *PHController,
	ec ECCorrector,
	mixer *Mixer,
	recipes *RecipeStore,
	audit *audit.Service,
) *NutrientService {
	return &NutrientService{ph: ph, ec: ec, mixer: mixer, recipes: recipes, audit: audit}
}

func (n *NutrientService) Adjust(shedID string, currentPH float64) (AdjustResult, error) {
	recipe := n.recipes.LoadOrDefault(shedID)
	phResult, err := n.ph.Adjust(shedID, currentPH, func(id string) error {
		_, err := n.ec.Correct(id, recipe.TargetEC)
		return err
	})
	if err != nil {
		return AdjustResult{}, err
	}
	mix, err := n.mixer.Mix(shedID, recipe.WaterPerDose, recipe.ConcentrateA+recipe.ConcentrateB)
	if err != nil {
		return AdjustResult{}, err
	}
	ecState, err := n.ec.Current(shedID)
	if err != nil {
		return AdjustResult{}, err
	}
	if _, err := n.audit.Record(shedID, "fert", "adjust", ftoa(ecState.Current)); err != nil {
		return AdjustResult{}, err
	}
	return AdjustResult{PH: phResult, EC: ecState, Mix: mix}, nil
}

func (n *NutrientService) Recipe(shedID string) Recipe {
	return n.recipes.LoadOrDefault(shedID)
}

func (n *NutrientService) SetRecipe(shedID string, recipe Recipe) error {
	return n.recipes.Save(recipe)
}
