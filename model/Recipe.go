package model

import "context"

type Recipe struct {
	Id          string               `json:"id"`
	Title       string               `json:"title"`
	Description *string              `json:"description"`
	Ingredients []ContainsIngredient `json:"ingredients"`
	Steps       []string             `json:"steps"` // TODO step templates? (e.g. preheat oven to {x} degress, bake for {y} time) // TODO reusable (linkable) steps?
	// TODO categories
	// TODO images
	Resource
}

type RecipeRepository interface {
	GetAll(ctx context.Context) ([]Recipe, error)
	GetById(ctx context.Context, id string) (*Recipe, bool, error)
	Create(ctx context.Context, recipe Recipe) (*Recipe, error)
	Update(ctx context.Context, recipe Recipe) (*Recipe, error)
	Delete(ctx context.Context, id string) (string, error)
}
