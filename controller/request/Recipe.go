package request

type CreateRecipeRequest struct {
	Title         string   `json:"title"`
	IngredientIds []string `json:"ingredient_ids"`
	Steps         []string `json:"steps"`
}

type UpdateRecipeRequest struct {
	Title         *string   `json:"title"`
	IngredientIds *[]string `json:"ingredient_ids"`
	Steps         *[]string `json:"steps"`
}
