package response

import "github.com/ThomasMatlak/food/model"

type GetFoodsResponse struct {
	Foods []model.Food `json:"ingredients"`
}

type DeleteFoodResponse struct {
	Id string `json:"id"`
}
