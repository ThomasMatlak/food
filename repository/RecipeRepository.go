package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ThomasMatlak/food/model"
	"github.com/ThomasMatlak/food/util"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
)

var dbUri string = "bolt://localhost:7687" // todo get from configuration
// todo abstract out the driver and context creation

func GetRecipes() ([]model.Recipe, error) {
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.NoAuth()) // todo implement auth
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) ([]model.Recipe, error) {
		query := "MATCH (r:Recipe) WHERE r.deleted IS NULL\n" +
			"RETURN r"
		params := map[string]any{}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		records, err := result.Collect(ctx)
		if err != nil {
			return nil, err
		}

		recipes := make([]model.Recipe, len(records))

		for i := 0; i < len(records); i++ {
			node, found := TypedGet[neo4j.Node](records[i], "r")
			if !found {
				continue
			}

			recipe, err := ParseRecipeNode(node)
			if err != nil {
				return nil, err
			}

			recipes[i] = *recipe
		}

		return recipes, nil
	})
}

func GetRecipe(id string) (*model.Recipe, error) {
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.NoAuth()) // todo implement auth
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
		query := fmt.Sprintf("%s WHERE r.deleted IS NULL\n"+
			"RETURN r",
			matchRecipeById)
		params := map[string]any{
			"id": id,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}

		node, found := TypedGet[neo4j.Node](record, "r")
		if !found {
			return nil, fmt.Errorf("could not find column")
		}

		return ParseRecipeNode(node)
	})
}

func CreateRecipe(recipe model.Recipe) (*model.Recipe, error) {
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.NoAuth()) // todo implement auth
	if err != nil {
		return nil, err
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

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}

		node, found := TypedGet[neo4j.Node](record, "r")
		if !found {
			return nil, fmt.Errorf("could not find column")
		}

		return ParseRecipeNode(node)
	})
}

func UpdateRecipe(recipe model.Recipe) (*model.Recipe, error) {
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.NoAuth()) // todo implement auth
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
		query := fmt.Sprintf("%s SET r.title = $title, r.steps = $steps, r.lastModified = $lastModified\n"+
			"RETURN r",
			matchRecipeById)
		params := map[string]any{
			"id":           recipe.Id,
			"title":        recipe.Title,
			"steps":        recipe.Steps,
			"lastModified": neo4j.LocalDateTime(*recipe.LastModified),
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}

		node, found := TypedGet[neo4j.Node](record, "r")
		if !found {
			return nil, fmt.Errorf("could not find column")
		}

		return ParseRecipeNode(node)
	})
}

func DeleteRecipe(id string) (string, error) {
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.NoAuth()) // todo implement auth
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (string, error) {
		query := fmt.Sprintf("%s SET r.deleted = $deleted\n"+
			"RETURN r.id AS id",
			matchRecipeById)
		params := map[string]any{
			"id":      id,
			"deleted": neo4j.LocalDateTime(time.Now()),
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return "", err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return "", err
		}

		deletedId, found := TypedGet[string](record, "id")
		if !found {
			return "", errors.New("missing id")
		}

		return deletedId, nil
	})
}

var matchRecipeById string = "MATCH (r:Recipe {id: $id})"

func ParseRecipeNode(node dbtype.Node) (*model.Recipe, error) {
	id, err := neo4j.GetProperty[string](node, "id")
	if err != nil {
		return nil, err
	}

	title, err := neo4j.GetProperty[string](node, "title")
	if err != nil {
		return nil, err
	}

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

	rawLastModified, err := neo4j.GetProperty[neo4j.LocalDateTime](node, "lastModified")
	lastModified := new(time.Time)
	if err != nil {
		lastModified = nil
	} else {
		*lastModified = rawLastModified.Time()
	}

	return &model.Recipe{Id: id, Title: title, Steps: steps, Created: created, LastModified: lastModified}, nil
}
