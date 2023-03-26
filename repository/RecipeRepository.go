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
)

type RecipeRepository struct {
	driver neo4j.DriverWithContext
}

func NewRecipeRepository(driver neo4j.DriverWithContext) *RecipeRepository {
	return &RecipeRepository{driver: driver}
}

func (r *RecipeRepository) GetAll(ctx context.Context) ([]model.Recipe, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) ([]model.Recipe, error) {
		return neo4j.ExecuteRead(ctx, session, func(tx neo4j.ManagedTransaction) ([]model.Recipe, error) {
			*query = fmt.Sprintf("MATCH (r:`%s`) WHERE r.deleted IS NULL\n"+
				"MATCH (r)-[ci:`%s`]->(i:`%s`) WHERE ci.deleted IS NULL AND i.deleted IS NULL\n"+
				"RETURN r AS recipe, collect({ingredient: i, rel: ci}) AS ingredients",
				RecipeLabel,
				ContainsIngredientLabel, IngredientLabel)
			params = map[string]any{}

			result, err := tx.Run(ctx, *query, params)
			if err != nil {
				return nil, err
			}

			records, err := result.Collect(ctx)
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

				rawIngredients, found := TypedGet[[]any](records[i], "ingredients")
				if !found {
					// TODO return error?
					continue
				}
				ingredients := util.UnpackArray[map[string]any](rawIngredients)

				err = setIngredients(recipe, ingredients)
				if err != nil {
					return nil, err
				}

				recipes[i] = *recipe
			}

			return recipes, nil
		})
	}

	return RunQuery(ctx, r.driver, "get all recipes", neo4j.AccessModeRead, work)
}

func (r *RecipeRepository) GetById(ctx context.Context, id string) (*model.Recipe, bool, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) (*model.Recipe, error) {
		return neo4j.ExecuteRead(ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
			*query = fmt.Sprintf("%s WHERE r.deleted IS NULL\n"+
				"MATCH (r)-[ci:`%s`]->(i:`%s`) WHERE ci.deleted IS NULL AND i.deleted IS NULL\n"+
				"RETURN r AS recipe, collect({ingredient: i, rel: ci}) AS ingredients",
				MatchNodeById("r", []string{RecipeLabel}),
				ContainsIngredientLabel, IngredientLabel)
			params = map[string]any{
				"rId": id,
			}

			record, err := RunAndReturnSingleRecord(ctx, tx, *query, params)
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

			rawIngredients, found := TypedGet[[]any](record, "ingredients")
			if !found {
				return nil, err
			}
			ingredients := util.UnpackArray[map[string]any](rawIngredients)

			err = setIngredients(recipe, ingredients)
			if err != nil {
				return nil, err
			}
			return recipe, nil
		})
	}

	recipe, err := RunQuery(ctx, r.driver, "get recipe", neo4j.AccessModeRead, work)

	if err != nil && err.Error() == "Result contains no more records" {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	return recipe, true, nil
}

func (r *RecipeRepository) Create(ctx context.Context, recipe model.Recipe) (*model.Recipe, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) (*model.Recipe, error) {
		return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
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
				ingredientIdParams[fmt.Sprintf("%sId", nodeVar)] = containsIngredient.IngredientId
				ingredientIdParams[fmt.Sprintf("%sUnit", nodeVar)] = containsIngredient.Unit
				ingredientIdParams[fmt.Sprintf("%sAmount", nodeVar)] = containsIngredient.Amount
			}

			labels := []string{RecipeLabel, ResourceLabel}
			id, err := model.ResourceId(labels)
			if err != nil {
				return nil, err
			}

			*query = fmt.Sprintf("CREATE (r:`%s`) SET r = {id: $id, title: $title, description: $description, steps: $steps, created: $created}\n"+
				"WITH r %s\n"+
				"WITH r MATCH (r)-[ci:`%s`]->(i:`%s`)\n"+ // the node and relationships have just been created, so no need to check they are not deleted
				"RETURN r AS recipe, collect({ingredient: i, rel: ci}) AS ingredients",
				strings.Join(labels, "`:`"),
				strings.Join(relateIngredientStmts, "\nWITH r "),
				ContainsIngredientLabel, IngredientLabel)
			params = map[string]any{
				"id":          id,
				"title":       recipe.Title,
				"description": recipe.Description,
				"steps":       recipe.Steps,
				"created":     neo4j.LocalDateTime(time.Now()),
			}
			for k, v := range ingredientIdParams {
				params[k] = v
			}

			record, err := RunAndReturnSingleRecord(ctx, tx, *query, params)
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

			rawIngredients, found := TypedGet[[]any](record, "ingredients")
			if !found {
				return nil, err
			}
			ingredients := util.UnpackArray[map[string]any](rawIngredients)

			err = setIngredients(recipe, ingredients)
			if err != nil {
				return nil, err
			}
			return recipe, nil
		})
	}

	return RunQuery(ctx, r.driver, "create recipe", neo4j.AccessModeWrite, work)
}

func (r *RecipeRepository) Update(ctx context.Context, recipe model.Recipe) (*model.Recipe, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) (*model.Recipe, error) {
		return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
			existingRecipe, found, err := r.GetById(ctx, recipe.Id)
			if !found && err != nil {
				return nil, err
			} else if !found {
				return nil, errors.New("recipe disappeared :(")
			}

			// TODO do this diffing on the db, if possible; very large ingredient lists could cause performance issues in the application
			newIngredients := map[string]model.ContainsIngredient{}
			for _, ci := range recipe.Ingredients {
				newIngredients[ci.IngredientId] = ci
			}
			existingIngredients := map[string]model.ContainsIngredient{}
			for _, ci := range existingRecipe.Ingredients {
				existingIngredients[ci.IngredientId] = ci
			}

			newIngredientIds := util.ArrayToSet(util.MapArray(recipe.Ingredients, func(ci model.ContainsIngredient) string { return ci.IngredientId }))
			existingIngredientIds := util.ArrayToSet(util.MapArray(existingRecipe.Ingredients, func(ci model.ContainsIngredient) string { return ci.IngredientId }))

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

			*query = fmt.Sprintf("%s SET r += {title: $title, description: $description, steps: $steps, lastModified: $lastModified}\n"+
				"WITH r %s\n"+
				"WITH r MATCH (r)-[ci:`%s`]->(i:`%s`) WHERE ci.deleted IS NULL AND i.deleted IS NULL\n"+
				"RETURN r AS recipe, collect({ingredient: i, rel: ci}) AS ingredients",
				MatchNodeById("r", []string{RecipeLabel}),
				strings.Join(relStmts, "\nWITH r "),
				ContainsIngredientLabel, IngredientLabel)
			params = map[string]any{
				"rId":          recipe.Id,
				"description":  recipe.Description,
				"title":        recipe.Title,
				"steps":        recipe.Steps,
				"lastModified": neo4j.LocalDateTime(time.Now()),
			}
			for k, v := range ingredientIdParams {
				params[k] = v
			}

			record, err := RunAndReturnSingleRecord(ctx, tx, *query, params)
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

			rawIngredients, found := TypedGet[[]any](record, "ingredients")
			if !found {
				return nil, err
			}
			ingredients := util.UnpackArray[map[string]any](rawIngredients)

			err = setIngredients(recipe, ingredients)
			if err != nil {
				return nil, err
			}
			return recipe, nil
		})
	}

	return RunQuery(ctx, r.driver, "update recipe", neo4j.AccessModeWrite, work)
}

// TODO return *string?
func (r *RecipeRepository) Delete(ctx context.Context, id string) (string, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) (string, error) {
		return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (string, error) {
			// TODO apply a Deleted label (and filter that :Resources are not also :Deleted)?
			*query = fmt.Sprintf("%s MATCH (r)-[rel]-(:`%s`) WHERE rel.deleted IS NULL\n"+
				"SET r.deleted = $deleted, rel.deleted = $deleted\n"+
				"WITH r RETURN r.id AS id",
				MatchNodeById("r", []string{RecipeLabel}), ResourceLabel)
			params = map[string]any{
				"rId":     id,
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

	return RunQuery(ctx, r.driver, "delete recipe", neo4j.AccessModeWrite, work)
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

func setIngredients(recipe *model.Recipe, ingredients []map[string]any) error {
	recipeIngredients := []model.ContainsIngredient{}
	for i := range ingredients {
		ingredient := ingredients[i]["ingredient"].(neo4j.Node)
		containsIngredientRel := ingredients[i]["rel"].(neo4j.Relationship)

		containsIngredient, err := ParseContainsIngredientRelationship(&ingredient, &containsIngredientRel)
		if err != nil {
			return err
		}

		recipeIngredients = append(recipeIngredients, *containsIngredient)
	}

	recipe.Ingredients = recipeIngredients
	return nil
}
