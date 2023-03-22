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

type IngredientRepository struct {
	driver neo4j.DriverWithContext // TODO *neo4j.DriverWithContext?
}

func NewIngredientRepository(driver neo4j.DriverWithContext) *IngredientRepository {
	return &IngredientRepository{driver: driver}
}

func (r *IngredientRepository) GetAll(ctx context.Context) ([]model.Ingredient, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) ([]model.Ingredient, error) {
		return neo4j.ExecuteRead(ctx, session, func(tx neo4j.ManagedTransaction) ([]model.Ingredient, error) {
			*query = fmt.Sprintf("MATCH (i:`%s`) WHERE i.deleted IS NULL\n"+
				"RETURN i",
				IngredientLabel)
			params = map[string]any{}

			result, err := tx.Run(ctx, *query, params)
			if err != nil {
				return nil, err
			}

			records, err := result.Collect(ctx)
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

	return RunQuery(ctx, r.driver, "get all ingredients", neo4j.AccessModeRead, work)
}

func (r *IngredientRepository) GetById(ctx context.Context, id string) (*model.Ingredient, bool, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) (*model.Ingredient, error) {
		return neo4j.ExecuteRead(ctx, session, func(tx neo4j.ManagedTransaction) (*model.Ingredient, error) {
			*query = fmt.Sprintf("%s WHERE i.deleted IS NULL\n"+
				"RETURN i",
				MatchNodeById("i", []string{IngredientLabel}))
			params = map[string]any{
				"iId": id,
			}

			record, err := RunAndReturnSingleRecord(ctx, tx, *query, params)
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

	ingredient, err := RunQuery(ctx, r.driver, "get ingredient", neo4j.AccessModeRead, work)

	if err != nil && err.Error() == "Result contains no more records" {
		return nil, false, nil
	} else if err != nil {
		fmt.Println(err)
		return nil, false, err
	}

	return ingredient, true, nil
}

func (r *IngredientRepository) Create(ctx context.Context, ingredient model.Ingredient) (*model.Ingredient, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) (*model.Ingredient, error) {
		return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (*model.Ingredient, error) {
			// TODO different id function?
			*query = fmt.Sprintf("CREATE (i:`%s`:`%s`) SET i.id = toString(id(i)), i.name = $name, i.created = $created\n"+
				"RETURN i",
				IngredientLabel, ResourceLabel)
			params = map[string]any{
				"name":    ingredient.Name,
				"created": neo4j.LocalDateTime(time.Now()),
			}

			record, err := RunAndReturnSingleRecord(ctx, tx, *query, params)
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

	return RunQuery(ctx, r.driver, "create ingredient", neo4j.AccessModeWrite, work)
}

func (r *IngredientRepository) Update(ctx context.Context, ingredient model.Ingredient) (*model.Ingredient, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) (*model.Ingredient, error) {
		return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (*model.Ingredient, error) {
			*query = fmt.Sprintf("%s SET i.name = $name, i.lastModified = $lastModified\n"+
				"RETURN i",
				MatchNodeById("i", []string{IngredientLabel}))
			params = map[string]any{
				"iId":          ingredient.Id,
				"name":         ingredient.Name,
				"lastModified": neo4j.LocalDateTime(time.Now()),
			}

			record, err := RunAndReturnSingleRecord(ctx, tx, *query, params)
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

	return RunQuery(ctx, r.driver, "update ingredient", neo4j.AccessModeWrite, work)
}

func (r *IngredientRepository) Delete(ctx context.Context, id string) (string, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) (string, error) {
		return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (string, error) {
			*query = fmt.Sprintf("%s MATCH (i)-[rel]-(:`%s`) WHERE rel.deleted IS NULL\n"+
				"SET i.deleted = $deleted, rel.deleted = $deleted\n"+
				"RETURN i.id AS id",
				MatchNodeById("i", []string{IngredientLabel}), ResourceLabel)
			params = map[string]any{
				"iId":     id,
				"deleted": neo4j.LocalDateTime(time.Now()),
			}

			record, err := RunAndReturnSingleRecord(ctx, tx, *query, params)
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

	return RunQuery(ctx, r.driver, "delete ingredient", neo4j.AccessModeWrite, work)
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

	resource, err := ParseResourceEntity(node)
	if err != nil {
		return nil, err
	}

	return &model.Ingredient{Id: id, Name: name, Resource: *resource}, nil
}
