package model

import "context"

type Food struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	// TODO nutrition
	Resource
}

type FoodRepository interface {
	GetAll(ctx context.Context) ([]Food, error)
	GetById(ctx context.Context, id string) (*Food, bool, error)
	Create(ctx context.Context, food Food) (*Food, error)
	Update(ctx context.Context, food Food) (*Food, error)
	Delete(ctx context.Context, id string) (string, error)
}
