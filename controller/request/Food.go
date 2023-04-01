package request

type CreateFoodRequest struct {
	Name string `json:"name"`
}

type UpdateFoodRequest struct {
	Name *string `json:"name"`
}
