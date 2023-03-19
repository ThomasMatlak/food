package request

import "github.com/ThomasMatlak/food/model"

type CreateRecipeRequest struct {
	Title       string                     `json:"title"`
	Description *string                    `json:"description"`
	Ingredients []model.ContainsIngredient `json:"ingredients"`
	Steps       []string                   `json:"steps"`
}

type UpdateRecipeRequest struct {
	Title       *string                     `json:"title"`
	Description *string                     `json:"description"`
	Ingredients *[]model.ContainsIngredient `json:"ingredients"`
	Steps       *[]string                   `json:"steps"`
}
