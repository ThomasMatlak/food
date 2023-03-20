package model

import "context"

type Ingredient struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	// TODO nutrition
	Resource
}

type IngredientRepository interface {
	GetAll(ctx context.Context) ([]Ingredient, error)
	GetById(ctx context.Context, id string) (*Ingredient, bool, error)
	Create(ctx context.Context, ingredient Ingredient) (*Ingredient, error)
	Update(ctx context.Context, ingredient Ingredient) (*Ingredient, error)
	Delete(ctx context.Context, id string) (string, error)
}
