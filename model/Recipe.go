package model

import "time"

type Recipe struct {
	Id            string     `json:"id"`
	Title         string     `json:"title"`
	IngredientIds []string   `json:"ingredient_ids"`
	Steps         []string   `json:"steps"`
	Created       *time.Time `json:"created"`
	LastModified  *time.Time `json:"last_modified"`
	Deleted       *time.Time `json:"deleted"`
}
