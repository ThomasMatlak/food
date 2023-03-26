package repository

import (
	"github.com/ThomasMatlak/food/model"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
)

func ParseContainsIngredientRelationship(ingredient *dbtype.Node, rel *dbtype.Relationship) (*model.ContainsIngredient, error) {
	ingredientId, err := neo4j.GetProperty[string](ingredient, "id")

	unit, err := neo4j.GetProperty[string](rel, "unit")
	if err != nil {
		return nil, err
	}

	amount, err := neo4j.GetProperty[int64](rel, "amount")
	if err != nil {
		return nil, err
	}

	resource, err := ParseResourceEntity(rel)
	if err != nil {
		return nil, err
	}

	return &model.ContainsIngredient{Unit: unit, Amount: amount, IngredientId: ingredientId, Resource: *resource}, nil
}
