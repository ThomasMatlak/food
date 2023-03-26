package model

type ContainsIngredient struct {
	Unit         string `json:"unit"`
	Amount       int64  `json:"amount"`
	IngredientId string `json:"ingredient_id"`
	// TODO order
	Resource
}
