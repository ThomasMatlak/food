package model

type Ingredient struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	// TODO nutrition
	Resource
}

type IngredientRepository interface {
	GetAll() ([]Ingredient, error)
	GetById(id string) (*Ingredient, bool, error)
	Create(Ingredient) (*Ingredient, error)
	Update(Ingredient) (*Ingredient, error)
	Delete(id string) (string, error)
}
