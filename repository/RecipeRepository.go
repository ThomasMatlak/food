package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ThomasMatlak/food/model"
	"github.com/ThomasMatlak/food/util"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/rs/zerolog/log"
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

	// TODO wrapper for logging query results
	var query string
	var params map[string]any

	recipes, err := neo4j.ExecuteRead(r.ctx, session, func(tx neo4j.ManagedTransaction) ([]model.Recipe, error) {
		query = fmt.Sprintf("MATCH (r:`%s`) WHERE r.deleted IS NULL\n"+
			"MATCH (r)-[ci:`%s`]->(i:`%s`) WHERE ci.deleted IS NULL AND i.deleted IS NULL\n"+
			"RETURN r AS recipe, collect(ci) AS rels",
			RecipeLabel,
			ContainsIngredientLabel, IngredientLabel)
		params = map[string]any{}

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
			recipeNode, found := TypedGet[neo4j.Node](records[i], "recipe")
			if !found {
				// TODO return error?
				continue
			}

			recipe, err := ParseRecipeNode(recipeNode)
			if err != nil {
				return nil, err
			}

			rawIngredients, found := TypedGet[[]any](records[i], "rels")
			if !found {
				// TODO return error?
				continue
			}
			ingredients := util.UnpackArray[dbtype.Relationship](rawIngredients)

			err = setIngredients(recipe, &ingredients)
			if err != nil {
				return nil, err
			}

			recipes[i] = *recipe
		}

		return recipes, nil
	})

	log.Debug().Str("query", query).Any("params", params).Any("result", recipes).Err(err).Msg("get all recipes")
	return recipes, err
}

func (r *RecipeRepository) GetById(id string) (*model.Recipe, bool, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	var query string
	var params map[string]any

	recipe, err := neo4j.ExecuteRead(r.ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
		query = fmt.Sprintf("%s WHERE r.deleted IS NULL\n"+
			"MATCH (r)-[ci:`%s`]->(i:`%s`) WHERE ci.deleted IS NULL AND i.deleted IS NULL\n"+
			"RETURN r AS recipe, collect(ci) AS rels",
			MatchNodeById("r", []string{RecipeLabel}),
			ContainsIngredientLabel, IngredientLabel)
		params = map[string]any{
			"rId": id,
		}

		record, err := RunAndReturnSingleRecord(r.ctx, tx, query, params)
		if err != nil {
			return nil, err
		}

		recipeNode, found := TypedGet[neo4j.Node](record, "recipe")
		if !found {
			return nil, errors.New("could not find column recipe")
		}

		recipe, err := ParseRecipeNode(recipeNode)
		if err != nil {
			return nil, err
		}

		rawIngredients, found := TypedGet[[]any](record, "rels")
		if !found {
			return nil, errors.New("could not find column rels")
		}
		ingredients := util.UnpackArray[dbtype.Relationship](rawIngredients)

		err = setIngredients(recipe, &ingredients)
		if err != nil {
			return nil, err
		}
		return recipe, nil
	})

	log.Debug().Str("query", query).Any("params", params).Any("result", recipe).Err(err).Msg("get recipe")

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

	var query string
	var params map[string]any

	createdRecipe, err := neo4j.ExecuteWrite(r.ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
		relateIngredientStmts := []string{}
		ingredientIdParams := map[string]any{}
		for i, containsIngredient := range recipe.Ingredients {
			nodeVar := fmt.Sprintf("i%d", i)
			unitVar := fmt.Sprintf("i%dUnit", i)
			amountVar := fmt.Sprintf("i%dAmount", i)
			// TODO fail if ingredient does not exist or has been deleted
			statement := fmt.Sprintf("%s CREATE (r)-[:`%s` {unit: $%s, amount: $%s, created: $created}]->(%s)",
				MatchNodeById(nodeVar, []string{IngredientLabel}), ContainsIngredientLabel, unitVar, amountVar, nodeVar)

			relateIngredientStmts = append(relateIngredientStmts, statement)
			ingredientIdParams[fmt.Sprintf("%sId", nodeVar)] = containsIngredient.TargetId
			ingredientIdParams[fmt.Sprintf("%sUnit", nodeVar)] = containsIngredient.Unit
			ingredientIdParams[fmt.Sprintf("%sAmount", nodeVar)] = containsIngredient.Amount
		}

		// TODO different id function?
		query = fmt.Sprintf("CREATE (r:`%s`:`%s`) SET r = {id: toString(id(r)), title: $title, description: $description, steps: $steps, created: $created}\n"+
			"WITH r %s\n"+
			"WITH r MATCH (r)-[ci:`%s`]->(:`%s`)\n"+ // the node and relationships have just been created, so no need to check they are not deleted
			"RETURN r AS recipe, collect(ci) AS rels",
			RecipeLabel, ResourceLabel,
			strings.Join(relateIngredientStmts, "\nWITH r "),
			ContainsIngredientLabel, IngredientLabel)
		params = map[string]any{
			"title":       recipe.Title,
			"description": recipe.Description,
			"steps":       recipe.Steps,
			"created":     neo4j.LocalDateTime(*recipe.Created),
		}
		for k, v := range ingredientIdParams {
			params[k] = v
		}

		record, err := RunAndReturnSingleRecord(r.ctx, tx, query, params)
		if err != nil {
			return nil, err
		}

		recipeNode, found := TypedGet[neo4j.Node](record, "recipe")
		if !found {
			return nil, errors.New("could not find column recipe")
		}

		recipe, err := ParseRecipeNode(recipeNode)
		if err != nil {
			return nil, err
		}

		rawIngredients, found := TypedGet[[]any](record, "rels")
		if !found {
			return nil, errors.New("could not find column rels")
		}
		ingredients := util.UnpackArray[dbtype.Relationship](rawIngredients)

		err = setIngredients(recipe, &ingredients)
		if err != nil {
			return nil, err
		}
		return recipe, nil
	})

	log.Debug().Str("query", query).Any("params", params).Any("result", createdRecipe).Err(err).Msg("create recipe")
	return createdRecipe, err
}

func (r *RecipeRepository) Update(recipe model.Recipe) (*model.Recipe, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	var query string
	var params map[string]any

	updatedRecipe, err := neo4j.ExecuteWrite(r.ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
		// TODO do on the db, if possible; very large ingredient lists could cause performance issues in the application

		existingRecipe, found, err := r.GetById(recipe.Id)
		if !found && err != nil {
			return nil, err
		} else if !found {
			return nil, errors.New("recipe disappeared :(")
		}

		newIngredients := map[string]model.ContainsIngredient{}
		for _, ci := range recipe.Ingredients {
			newIngredients[ci.TargetId] = ci
		}
		existingIngredients := map[string]model.ContainsIngredient{}
		for _, ci := range existingRecipe.Ingredients {
			existingIngredients[ci.TargetId] = ci
		}

		newIngredientIds := util.ArrayToSet(util.Map(recipe.Ingredients, func(ci model.ContainsIngredient) string { return ci.TargetId }))
		existingIngredientIds := util.ArrayToSet(util.Map(existingRecipe.Ingredients, func(ci model.ContainsIngredient) string { return ci.TargetId }))

		removedIngredientIds := util.Difference(existingIngredientIds, newIngredientIds)
		addedIngredientIds := util.Difference(newIngredientIds, existingIngredientIds)
		updatedIngredientIds := util.Intersection(existingIngredientIds, newIngredientIds)

		unRelateIngredientStms := []string{}
		ingredientIdParams := map[string]any{}
		for i, ingredientId := range util.SetToArray(removedIngredientIds) {
			nodeVar := fmt.Sprintf("i%d", i)
			relVar := fmt.Sprintf("ci%d", i)
			statement := fmt.Sprintf("%s<-[`%s`:`%s`]-(r) SET `%s`.deleted = $lastModified",
				MatchNodeById(nodeVar, []string{IngredientLabel}), relVar, ContainsIngredientLabel, relVar)
			unRelateIngredientStms = append(unRelateIngredientStms, statement)
			ingredientIdParams[fmt.Sprintf("%sId", nodeVar)] = ingredientId
		}

		offset := len(unRelateIngredientStms)
		addIngredientStms := []string{}
		for i, ingredientId := range util.SetToArray(addedIngredientIds) {
			nodeVar := fmt.Sprintf("i%d", offset+i)
			unitVar := fmt.Sprintf("i%dUnit", offset+i)
			amountVar := fmt.Sprintf("i%dAmount", offset+i)
			statement := fmt.Sprintf("%s CREATE (r)-[:`%s` {unit: $%s, amount: $%s, created: $lastModified}]->(`%s`)",
				MatchNodeById(nodeVar, []string{IngredientLabel}), ContainsIngredientLabel, unitVar, amountVar, nodeVar)
			addIngredientStms = append(addIngredientStms, statement)
			ingredientIdParams[fmt.Sprintf("%sId", nodeVar)] = ingredientId
			ingredientIdParams[fmt.Sprintf("%sUnit", nodeVar)] = newIngredients[ingredientId].Unit
			ingredientIdParams[fmt.Sprintf("%sAmount", nodeVar)] = newIngredients[ingredientId].Amount
		}

		offset += len(addIngredientStms)
		updateIngredientStms := []string{}
		for i, ingredientId := range util.SetToArray(updatedIngredientIds) {
			nodeVar := fmt.Sprintf("i%d", offset+i)
			relVar := fmt.Sprintf("ci%d", offset+i)
			unitVar := fmt.Sprintf("i%dUnit", offset+i)
			amountVar := fmt.Sprintf("i%dAmount", offset+i)
			statement := fmt.Sprintf("%s MATCH (r)-[`%s`:`%s`]->(`%s`) SET `%s` += {unit: $%s, amount: $%s, lastModified: $lastModified}",
				MatchNodeById(nodeVar, []string{IngredientLabel}), relVar, ContainsIngredientLabel, nodeVar, relVar, unitVar, amountVar)
			updateIngredientStms = append(updateIngredientStms, statement)
			ingredientIdParams[fmt.Sprintf("%sId", nodeVar)] = ingredientId
			ingredientIdParams[fmt.Sprintf("%sUnit", nodeVar)] = newIngredients[ingredientId].Unit
			ingredientIdParams[fmt.Sprintf("%sAmount", nodeVar)] = newIngredients[ingredientId].Amount
		}

		relStmts := append(append(unRelateIngredientStms, addIngredientStms...), updateIngredientStms...)

		query = fmt.Sprintf("%s SET r += {title: $title, description: $description, steps: $steps, lastModified: $lastModified}\n"+
			"WITH r %s\n"+
			"WITH r MATCH (r)-[ci:`%s`]->(i:`%s`) WHERE ci.deleted IS NULL AND i.deleted IS NULL\n"+
			"RETURN r AS recipe, collect(ci) AS rels",
			MatchNodeById("r", []string{RecipeLabel}),
			strings.Join(relStmts, "\nWITH r "),
			ContainsIngredientLabel, IngredientLabel)
		params = map[string]any{
			"rId":          recipe.Id,
			"description":  recipe.Description,
			"title":        recipe.Title,
			"steps":        recipe.Steps,
			"lastModified": neo4j.LocalDateTime(*recipe.LastModified),
		}
		for k, v := range ingredientIdParams {
			params[k] = v
		}

		record, err := RunAndReturnSingleRecord(r.ctx, tx, query, params)
		if err != nil {
			return nil, err
		}

		recipeNode, found := TypedGet[neo4j.Node](record, "recipe")
		if !found {
			return nil, errors.New("could not find column recipe")
		}

		recipe, err := ParseRecipeNode(recipeNode)
		if err != nil {
			return nil, err
		}

		rawIngredients, found := TypedGet[[]any](record, "rels")
		if !found {
			return nil, errors.New("could not find column rels")
		}
		ingredients := util.UnpackArray[dbtype.Relationship](rawIngredients)

		err = setIngredients(recipe, &ingredients)
		if err != nil {
			return nil, err
		}
		return recipe, nil
	})

	log.Debug().Str("query", query).Any("params", params).Any("result", updatedRecipe).Err(err).Msg("update recipe")
	return updatedRecipe, err
}

func (r *RecipeRepository) Delete(id string) (string, error) {
	session := r.driver.NewSession(r.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(r.ctx)

	var query string
	var params map[string]any

	deletedId, err := neo4j.ExecuteWrite(r.ctx, session, func(tx neo4j.ManagedTransaction) (string, error) {
		// TODO apply a Deleted label (and filter that :Resources are not also :Deleted)?
		query = fmt.Sprintf("%s MATCH (r)-[rel]-(:`%s`) WHERE rel.deleted IS NULL\n"+
			"SET r.deleted = $deleted, rel.deleted = $deleted\n"+
			"WITH r RETURN r.id AS id",
			MatchNodeById("r", []string{RecipeLabel}), ResourceLabel)
		params = map[string]any{
			"rId":     id,
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

	// TODO log deleted rels?
	log.Debug().Str("query", query).Any("params", params).Any("result", deletedId).Err(err).Msg("delete recipe")
	return deletedId, err
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

	resource, err := ParseResourceEntity(node)
	if err != nil {
		return nil, err
	}

	return &model.Recipe{Id: id, Title: title, Description: description, Steps: steps, Resource: *resource}, nil
}

func setIngredients(recipe *model.Recipe, containsIngredientRels *[]dbtype.Relationship) error {
	recipeIngredients := []model.ContainsIngredient{}
	for i := range *containsIngredientRels {
		containsIngredientRel := (*containsIngredientRels)[i]
		containsIngredient, err := ParseContainsIngredientRelationship(containsIngredientRel)
		if err != nil {
			return err
		}

		recipeIngredients = append(recipeIngredients, *containsIngredient)
	}

	recipe.Ingredients = recipeIngredients
	return nil
}
