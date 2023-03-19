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

type RecipeRepository struct {
	ctx    context.Context
	driver neo4j.DriverWithContext
}

func NewRecipeRepository(ctx context.Context, driver neo4j.DriverWithContext) *RecipeRepository {
	return &RecipeRepository{ctx: ctx, driver: driver}
}

func (r *RecipeRepository) GetAll() ([]model.Recipe, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	return neo4j.ExecuteWrite(r.ctx, session, func(tx neo4j.ManagedTransaction) ([]model.Recipe, error) {
		query := "MATCH (r:Recipe) WHERE r.deleted IS NULL\n" +
			"RETURN r"
		params := map[string]any{}

		result, err := tx.Run(r.ctx, query, params)
		if err != nil {
			return nil, err
		}

		records, err := result.Collect(r.ctx)
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

func (r *RecipeRepository) GetById(id string) (*model.Recipe, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	return neo4j.ExecuteWrite(r.ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
		query := fmt.Sprintf("%s WHERE r.deleted IS NULL\n"+
			"RETURN r",
			matchRecipeById)
		params := map[string]any{
			"id": id,
		}

		result, err := tx.Run(r.ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := result.Single(r.ctx)
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

func (r *RecipeRepository) Create(recipe model.Recipe) (*model.Recipe, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	return neo4j.ExecuteWrite(r.ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
		// todo different id function?
		query := "CREATE (r:Recipe) SET r.id = toString(id(r)), r.title = $title, description = $description, r.steps = $steps, r.created = $created\n" +
			"RETURN r"
		params := map[string]any{
			"title":       recipe.Title,
			"description": recipe.Description,
			"steps":       recipe.Steps,
			"created":     neo4j.LocalDateTime(*recipe.Created),
		}

		result, err := tx.Run(r.ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := result.Single(r.ctx)
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

func (r *RecipeRepository) Update(recipe model.Recipe) (*model.Recipe, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	return neo4j.ExecuteWrite(r.ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
		query := fmt.Sprintf("%s SET r.title = $title, description = $description, r.steps = $steps, r.lastModified = $lastModified\n"+
			"RETURN r",
			matchRecipeById)
		params := map[string]any{
			"id":           recipe.Id,
			"description":  recipe.Description,
			"title":        recipe.Title,
			"steps":        recipe.Steps,
			"lastModified": neo4j.LocalDateTime(*recipe.LastModified),
		}

		result, err := tx.Run(r.ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := result.Single(r.ctx)
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

func (r *RecipeRepository) Delete(id string) (string, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	return neo4j.ExecuteWrite(r.ctx, session, func(tx neo4j.ManagedTransaction) (string, error) {
		query := fmt.Sprintf("%s SET r.deleted = $deleted\n"+
			"RETURN r.id AS id",
			matchRecipeById)
		params := map[string]any{
			"id":      id,
			"deleted": neo4j.LocalDateTime(time.Now()),
		}

		result, err := tx.Run(r.ctx, query, params)
		if err != nil {
			return "", err
		}

		record, err := result.Single(r.ctx)
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

	description := new(string)
	rawDescription, err := neo4j.GetProperty[string](node, "description")
	if err != nil {
		description = nil
	} else {
		*description = rawDescription
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

	return &model.Recipe{Id: id, Title: title, Description: description, Steps: steps, Created: created, LastModified: lastModified}, nil
}
