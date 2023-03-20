package model

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
	GetAll() ([]Recipe, error)
	GetById(id string) (*Recipe, bool, error)
	Create(recipe Recipe) (*Recipe, error)
	Update(recipe Recipe) (*Recipe, error)
	Delete(id string) (string, error)
}
