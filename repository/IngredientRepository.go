package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ThomasMatlak/food/model"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
)

// TODO don't store context in structs https://pkg.go.dev/context#section-documentation
type IngredientRepository struct {
	ctx    context.Context
	driver neo4j.DriverWithContext
}

func NewIngredientRepository(ctx context.Context, driver neo4j.DriverWithContext) *IngredientRepository {
	return &IngredientRepository{ctx: ctx, driver: driver}
}

func (r *IngredientRepository) GetAll() ([]model.Ingredient, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	return neo4j.ExecuteRead(r.ctx, session, func(tx neo4j.ManagedTransaction) ([]model.Ingredient, error) {
		query := fmt.Sprintf("MATCH (i:`%s`) WHERE i.deleted IS NULL\n"+
			"RETURN i",
			IngredientLabel)
		params := map[string]any{}

		result, err := tx.Run(r.ctx, query, params)
		if err != nil {
			return nil, err
		}

		records, err := result.Collect(r.ctx)
		if err != nil {
			return nil, err
		}

		ingredients := make([]model.Ingredient, len(records))

		for i := 0; i < len(records); i++ {
			node, found := TypedGet[neo4j.Node](records[i], "i")
			if !found {
				// TODO return error?
				continue
			}

			ingredient, err := ParseIngredientNode(node)
			if err != nil {
				return nil, err
			}

			ingredients[i] = *ingredient
		}

		return ingredients, nil
	})
}

func (r *IngredientRepository) GetById(id string) (*model.Ingredient, bool, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	ingredient, err := neo4j.ExecuteRead(r.ctx, session, func(tx neo4j.ManagedTransaction) (*model.Ingredient, error) {
		query := fmt.Sprintf("%s WHERE i.deleted IS NULL\n"+
			"RETURN i",
			MatchNodeById("i", []string{IngredientLabel}))
		params := map[string]any{
			"id": id,
		}

		record, err := RunAndReturnSingleRecord(r.ctx, tx, query, params)
		if err != nil {
			return nil, err
		}

		node, found := TypedGet[neo4j.Node](record, "i")
		if !found {
			return nil, fmt.Errorf("could not find column i")
		}

		return ParseIngredientNode(node)
	})

	if err != nil && err.Error() == "Result contains no more records" {
		return nil, false, nil
	} else if err != nil {
		fmt.Println(err)
		return nil, false, err
	}

	return ingredient, true, nil
}

func (r *IngredientRepository) Create(ingredient model.Ingredient) (*model.Ingredient, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	return neo4j.ExecuteWrite(r.ctx, session, func(tx neo4j.ManagedTransaction) (*model.Ingredient, error) {
		// TODO different id function?
		query := fmt.Sprintf("CREATE (i:`%s`:`%s`) SET i.id = toString(id(i)), i.name = $name, i.created = $created\n"+
			"RETURN i",
			IngredientLabel, ResourceLabel)
		params := map[string]any{
			"name":    ingredient.Name,
			"created": neo4j.LocalDateTime(*ingredient.Created),
		}

		record, err := RunAndReturnSingleRecord(r.ctx, tx, query, params)
		if err != nil {
			return nil, err
		}

		node, found := TypedGet[neo4j.Node](record, "i")
		if !found {
			return nil, fmt.Errorf("could not find column i")
		}

		return ParseIngredientNode(node)
	})
}

func (r *IngredientRepository) Update(ingredient model.Ingredient) (*model.Ingredient, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	return neo4j.ExecuteWrite(r.ctx, session, func(tx neo4j.ManagedTransaction) (*model.Ingredient, error) {
		query := fmt.Sprintf("%s SET i.name = $name, i.lastModified = $lastModified\n"+
			"RETURN i",
			MatchNodeById("i", []string{IngredientLabel}))
		params := map[string]any{
			"id":           ingredient.Id,
			"name":         ingredient.Name,
			"lastModified": neo4j.LocalDateTime(*ingredient.LastModified),
		}

		record, err := RunAndReturnSingleRecord(r.ctx, tx, query, params)
		if err != nil {
			return nil, err
		}

		node, found := TypedGet[neo4j.Node](record, "i")
		if !found {
			return nil, fmt.Errorf("could not find column i")
		}

		return ParseIngredientNode(node)
	})
}

func (r *IngredientRepository) Delete(id string) (string, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	return neo4j.ExecuteWrite(r.ctx, session, func(tx neo4j.ManagedTransaction) (string, error) {
		query := fmt.Sprintf("%s SET i.deleted = $deleted\n"+
			"RETURN i.id AS id",
			MatchNodeById("i", []string{IngredientLabel}))
		params := map[string]any{
			"id":      id,
			"deleted": neo4j.LocalDateTime(time.Now()),
		}

		record, err := RunAndReturnSingleRecord(r.ctx, tx, query, params)
		if err != nil {
			return "", err
		}

		deletedId, found := TypedGet[string](record, "id")
		if !found {
			return "", errors.New("could not find column id")
		}

		return deletedId, nil
	})
}

func ParseIngredientNode(node dbtype.Node) (*model.Ingredient, error) {
	id, err := neo4j.GetProperty[string](node, "id")
	if err != nil {
		return nil, err
	}

	name, err := neo4j.GetProperty[string](node, "name")
	if err != nil {
		return nil, err
	}

	resource, err := ParseResourceNode(node)
	if err != nil {
		return nil, err
	}

	return &model.Ingredient{Id: id, Name: name, Resource: *resource}, nil
}
