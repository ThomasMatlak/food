package response

import "github.com/ThomasMatlak/food/model"

type GetRecipesResponse struct {
	Recipes []model.Recipe `json:"recipes"`
}

type DeleteRecipeResponse struct {
	Id string `json:"id"`
}
