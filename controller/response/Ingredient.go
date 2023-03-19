package response

import "github.com/ThomasMatlak/food/model"

type GetIngredientsResponse struct {
	Ingredients []model.Ingredient `json:"ingredients"`
}

type DeleteIngredientResponse struct {
	Id string `json:"id"`
}
