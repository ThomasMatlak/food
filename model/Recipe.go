package model

type Recipe struct {
	// TODO categories
	// TODO images
	Id            string   `json:"id"`
	Title         string   `json:"title"`
	Description   *string  `json:"description"`
	IngredientIds []string `json:"ingredient_ids"`
	Steps         []string `json:"steps"`
	Resource
}

type RecipeRepository interface {
	GetAll() ([]Recipe, error)
	GetById(id string) (*Recipe, bool, error)
	Create(recipe Recipe) (*Recipe, error)
	Update(recipe Recipe) (*Recipe, error)
	Delete(id string) (string, error)
}
