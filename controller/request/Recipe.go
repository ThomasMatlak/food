package request

import (
	"strings"

	"github.com/ThomasMatlak/food/model"
)

type CreateRecipeRequest struct {
	Title       string                     `json:"title"`
	Description *string                    `json:"description"`
	Ingredients []model.ContainsIngredient `json:"ingredients"`
	Steps       []string                   `json:"steps"`
}

func CanCreateRecipe(request *CreateRecipeRequest) bool {
	if len(strings.TrimSpace(request.Title)) == 0 {
		return false
	}
	if request.Description != nil && len(strings.TrimSpace(*request.Description)) == 0 {
		return false
	}
	if len(request.Ingredients) == 0 {
		return false
	}
	// TODO validation of steps?

	return true
}

type UpdateRecipeRequest struct {
	Title       *string                     `json:"title"`
	Description *string                     `json:"description"`
	Ingredients *[]model.ContainsIngredient `json:"ingredients"`
	Steps       *[]string                   `json:"steps"`
}

func CanUpdateRecipe(request *UpdateRecipeRequest) bool {
	if request.Title != nil && len(strings.TrimSpace(*request.Title)) == 0 {
		return false
	}
	if request.Description != nil && len(strings.TrimSpace(*request.Description)) == 0 {
		return false
	}
	if request.Ingredients != nil && len(*request.Ingredients) == 0 {
		return false
	}
	// TODO validation of steps?

	return true
}
