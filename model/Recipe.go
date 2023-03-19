package model

import "time"

type Recipe struct {
	Id            string     `json:"id"`
	Title         string     `json:"title"`
	Description   *string    `json:"description"`
	IngredientIds []string   `json:"ingredient_ids"`
	Steps         []string   `json:"steps"`
	Created       *time.Time `json:"created"`
	LastModified  *time.Time `json:"last_modified"`
	Deleted       *time.Time `json:"deleted"`
}

type RecipeRepository interface {
	GetAll() ([]Recipe, error)
	GetById(id string) (*Recipe, error)
	Create(recipe Recipe) (*Recipe, error)
	Update(recipe Recipe) (*Recipe, error)
	Delete(id string) (string, error)
}
