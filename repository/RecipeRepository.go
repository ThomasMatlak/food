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

// TODO don't store context in structs https://pkg.go.dev/context#section-documentation
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

	return neo4j.ExecuteRead(r.ctx, session, func(tx neo4j.ManagedTransaction) ([]model.Recipe, error) {
		query := fmt.Sprintf("MATCH (r:`%s`) WHERE r.deleted IS NULL\n"+
			"RETURN r",
			RecipeLabel)
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
				// TODO return error?
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

func (r *RecipeRepository) GetById(id string) (*model.Recipe, bool, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	recipe, err := neo4j.ExecuteRead(r.ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
		query := fmt.Sprintf("%s WHERE r.deleted IS NULL\n"+
			"RETURN r",
			MatchNodeById("r", []string{RecipeLabel}))
		params := map[string]any{
			"id": id,
		}

		record, err := RunAndReturnSingleRecord(r.ctx, tx, query, params)
		if err != nil {
			return nil, err
		}

		node, found := TypedGet[neo4j.Node](record, "r")
		if !found {
			return nil, errors.New("could not find column r")
		}

		return ParseRecipeNode(node)
	})

	if err != nil && err.Error() == "Result contains no more records" {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	return recipe, true, nil
}

func (r *RecipeRepository) Create(recipe model.Recipe) (*model.Recipe, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	return neo4j.ExecuteWrite(r.ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
		// TODO different id function?
		query := fmt.Sprintf("CREATE (r:`%s`:`%s`) SET r.id = toString(id(r)), r.title = $title, r.description = $description, r.steps = $steps, r.created = $created\n"+
			"RETURN r",
			RecipeLabel, ResourceLabel)
		params := map[string]any{
			"title":       recipe.Title,
			"description": recipe.Description,
			"steps":       recipe.Steps,
			"created":     neo4j.LocalDateTime(*recipe.Created),
		}

		record, err := RunAndReturnSingleRecord(r.ctx, tx, query, params)
		if err != nil {
			return nil, err
		}

		node, found := TypedGet[neo4j.Node](record, "r")
		if !found {
			return nil, errors.New("could not find column r")
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
			MatchNodeById("r", []string{RecipeLabel}))
		params := map[string]any{
			"id":           recipe.Id,
			"description":  recipe.Description,
			"title":        recipe.Title,
			"steps":        recipe.Steps,
			"lastModified": neo4j.LocalDateTime(*recipe.LastModified),
		}

		record, err := RunAndReturnSingleRecord(r.ctx, tx, query, params)
		if err != nil {
			return nil, err
		}

		node, found := TypedGet[neo4j.Node](record, "r")
		if !found {
			return nil, errors.New("could not find column r")
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
			MatchNodeById("r", []string{RecipeLabel}))
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

	resource, err := ParseResourceNode(node)
	if err != nil {
		return nil, err
	}

	return &model.Recipe{Id: id, Title: title, Description: description, Steps: steps, Resource: *resource}, nil
}
