package request

type CreateIngredientRequest struct {
	Name string `json:"name"`
}

type UpdateIngredientRequest struct {
	Name *string `json:"name"`
}
