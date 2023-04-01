package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ThomasMatlak/food/model"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
)

type FoodRepository struct {
	driver neo4j.DriverWithContext // TODO *neo4j.DriverWithContext?
}

func NewFoodRepository(driver neo4j.DriverWithContext) *FoodRepository {
	return &FoodRepository{driver: driver}
}

func (r *FoodRepository) GetAll(ctx context.Context) ([]model.Food, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) ([]model.Food, error) {
		return neo4j.ExecuteRead(ctx, session, func(tx neo4j.ManagedTransaction) ([]model.Food, error) {
			*query = fmt.Sprintf("MATCH (i:`%s`) WHERE i.deleted IS NULL\n"+
				"RETURN i",
				FoodLabel)
			params = map[string]any{}

			result, err := tx.Run(ctx, *query, params)
			if err != nil {
				return nil, err
			}

			records, err := result.Collect(ctx)
			if err != nil {
				return nil, err
			}

			foods := make([]model.Food, len(records))

			for i := 0; i < len(records); i++ {
				node, found := TypedGet[neo4j.Node](records[i], "i")
				if !found {
					// TODO return error?
					continue
				}

				food, err := ParseFoodNode(node)
				if err != nil {
					return nil, err
				}

				foods[i] = *food
			}

			return foods, nil
		})
	}

	return RunQuery(ctx, r.driver, "get all foods", neo4j.AccessModeRead, work)
}

func (r *FoodRepository) GetById(ctx context.Context, id string) (*model.Food, bool, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) (*model.Food, error) {
		return neo4j.ExecuteRead(ctx, session, func(tx neo4j.ManagedTransaction) (*model.Food, error) {
			*query = fmt.Sprintf("%s WHERE i.deleted IS NULL\n"+
				"RETURN i",
				MatchNodeById("i", []string{FoodLabel}))
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

			return ParseFoodNode(node)
		})
	}

	food, err := RunQuery(ctx, r.driver, "get food", neo4j.AccessModeRead, work)

	if err != nil && err.Error() == "Result contains no more records" {
		return nil, false, nil
	} else if err != nil {
		fmt.Println(err)
		return nil, false, err
	}

	return food, true, nil
}

func (r *FoodRepository) Create(ctx context.Context, food model.Food) (*model.Food, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) (*model.Food, error) {
		return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (*model.Food, error) {
			labels := []string{FoodLabel, ResourceLabel}
			id, err := model.ResourceId(labels)
			if err != nil {
				return nil, err
			}

			*query = fmt.Sprintf("CREATE (i:`%s`) SET i = {id: $id, name: $name, created: $created}\n"+
				"RETURN i",
				strings.Join(labels, "`:`"))
			params = map[string]any{
				"id":      id,
				"name":    food.Name,
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

			return ParseFoodNode(node)
		})
	}

	return RunQuery(ctx, r.driver, "create food", neo4j.AccessModeWrite, work)
}

func (r *FoodRepository) Update(ctx context.Context, food model.Food) (*model.Food, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) (*model.Food, error) {
		return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (*model.Food, error) {
			*query = fmt.Sprintf("%s SET i += {name: $name, lastModified: $lastModified}\n"+
				"RETURN i",
				MatchNodeById("i", []string{FoodLabel}))
			params = map[string]any{
				"iId":          food.Id,
				"name":         food.Name,
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

			return ParseFoodNode(node)
		})
	}

	return RunQuery(ctx, r.driver, "update food", neo4j.AccessModeWrite, work)
}

func (r *FoodRepository) Delete(ctx context.Context, id string) (string, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) (string, error) {
		return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (string, error) {
			*query = fmt.Sprintf("%s OPTIONAL MATCH (i)-[rel]-(:`%s`) WHERE rel.deleted IS NULL\n"+
				"SET i.deleted = $deleted, rel.deleted = $deleted\n"+
				"RETURN i.id AS id",
				MatchNodeById("i", []string{FoodLabel}), ResourceLabel)
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

	return RunQuery(ctx, r.driver, "delete food", neo4j.AccessModeWrite, work)
}

func ParseFoodNode(node dbtype.Node) (*model.Food, error) {
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

	return &model.Food{Id: id, Name: name, Resource: *resource}, nil
}
