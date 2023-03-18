package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ThomasMatlak/food/model"
	"github.com/ThomasMatlak/food/util"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var dbUri string = "bolt://localhost:7687" // todo get from configuration

func CreateRecipe(recipe model.Recipe) (*model.Recipe, error) {
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.NoAuth()) // todo implement auth
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
		// todo different id function?
		query := "CREATE (r:Recipe) SET r.id = toString(id(r)), r.title = $title, r.steps = $steps, r.created = $created\n" +
			"RETURN r"
		params := map[string]any{
			"title":   recipe.Title,
			"steps":   recipe.Steps,
			"created": neo4j.LocalDateTime(*recipe.Created),
		}

		records, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := records.Single(ctx)
		if err != nil {
			return nil, err
		}

		rawNode, found := record.Get("r")
		if !found {
			return nil, fmt.Errorf("could not find column")
		}

		node := rawNode.(neo4j.Node)

		id, err := neo4j.GetProperty[string](node, "id")
		if err != nil {
			return nil, err
		}

		title, err := neo4j.GetProperty[string](node, "title")
		if err != nil {
			return nil, err
		}

		// ingredientIds, err := neo4j.GetProperty[string](node, "ingredientIds")
		// if err != nil {
		// 	return nil, err
		// }

		rawSteps, err := neo4j.GetProperty[[]any](node, "steps")
		if err != nil {
			return nil, err
		}
		steps := util.UnpackArray[string](rawSteps)

		rawCreated, err := neo4j.GetProperty[neo4j.LocalDateTime](node, "created")
		if err != nil {
			return nil, err
		}
		created := new(time.Time)
		*created = rawCreated.Time()

		return &model.Recipe{Id: id, Title: title, Steps: steps, Created: created}, nil
	})
}

// func Replace(recipe model.Recipe) (model.Recipe, error) {
// 	query := "MERGE (r:Recipe {id: $id}) SET r.title = $title, r.steps = $steps"
// }

func DeleteRecipe(id string) (string, error) {
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.NoAuth()) // todo implement auth
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (string, error) {
		query := "MATCH (r:Recipe {id: $id}) SET r.deleted = $deleted\n" +
			"RETURN r"
		params := map[string]any{
			"id":      id,
			"deleted": neo4j.LocalDateTime(time.Now()),
		}

		records, err := tx.Run(ctx, query, params)
		if err != nil {
			return "", err
		}

		record, err := records.Single(ctx)
		if err != nil {
			return "", err
		}

		rawNode, found := record.Get("r")
		if !found {
			return "", fmt.Errorf("could not find column")
		}

		node := rawNode.(neo4j.Node)

		deletedId, err := neo4j.GetProperty[string](node, "id")
		if err != nil {
			return "", err
		}

		return deletedId, nil
	})
}
