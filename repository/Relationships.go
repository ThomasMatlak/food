package repository

import (
	"strings"

	"github.com/ThomasMatlak/food/model"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
)

func ParseContainsIngredientRelationship(rel dbtype.Relationship) (*model.ContainsIngredient, error) {
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

	relationship := ParseRelationship(rel)

	return &model.ContainsIngredient{Unit: unit, Amount: amount, Relationship: *relationship, Resource: *resource}, nil
}

func ParseRelationship(rel dbtype.Relationship) *model.Relationship {
	source := strings.Split(rel.StartElementId, ":")
	target := strings.Split(rel.EndElementId, ":")

	return &model.Relationship{SourceId: &source[len(source)-1], TargetId: target[len(target)-1]}
}
