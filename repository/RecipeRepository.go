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

func checkIngredientsExist(ctx context.Context, tx neo4j.ManagedTransaction, ingredientIds util.Set[string]) (bool, error) {
	query := fmt.Sprintf("MATCH (i:`%s`) WHERE i.id IN $ids AND i.deleted IS NULL\n"+ // TODO IN, or UNWIND? any performance difference?
		"RETURN count(DISTINCT i) AS c",
		IngredientLabel,
	)
	params := map[string]any{"ids": util.SetToArray(ingredientIds)}

	record, err := RunAndReturnSingleRecord(ctx, tx, query, params)
	if err != nil {
		return false, err
	}

	ingredientCount, found := TypedGet[int64](record, "c")
	if !found {
		return false, errors.New("could not find column c")
	}

	return ingredientCount == int64(len(ingredientIds)), nil
}

func (r *RecipeRepository) Create(ctx context.Context, recipe model.Recipe) (*model.Recipe, error) {
	work := func(ctx context.Context, session neo4j.SessionWithContext, query *string, params map[string]any) (*model.Recipe, error) {
		return neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (*model.Recipe, error) {
			labels := []string{RecipeLabel, ResourceLabel}
			id, err := model.ResourceId(labels)
			if err != nil {
				return nil, err
			}

			// TODO is this possible to do in the same query as creating the relationships without getting super ugly?
			ingredientIds := util.ArrayToSet(util.MapArray(recipe.Ingredients, model.ExtractIngredientId))
			ingredientsExist, err := checkIngredientsExist(ctx, tx, ingredientIds)
			if err != nil {
				return nil, err
			}
			if !ingredientsExist {
				return nil, errors.New("tried to create a recipe with non-existent ingredient(s)")
			}

			// TODO fail the query if any 1 of the ingredients is not found
			*query = fmt.Sprintf("CREATE (r:`%s`) SET r = {id: $id, title: $title, description: $description, steps: $steps, created: $created}\n"+
				"WITH r UNWIND $ingredients AS ingredient\n"+
				"MATCH (i:`%s` {id: ingredient.id}) WHERE i.deleted IS NULL\n"+
				"CREATE (r)-[ci:`%s` {unit: ingredient.unit, amount: ingredient.amount, created: $created}]->(i)\n"+
				"RETURN r AS recipe, collect({ingredient: i, rel: ci}) AS ingredients",
				strings.Join(labels, "`:`"),
				IngredientLabel,
				ContainsIngredientLabel,
			)

			ingredientParams := []map[string]any{}
			for _, ci := range recipe.Ingredients {
				ingredient := map[string]any{}
				ingredient["id"] = ci.IngredientId
				ingredient["unit"] = ci.Unit
				ingredient["amount"] = ci.Amount
				ingredientParams = append(ingredientParams, ingredient)
			}
			params = map[string]any{
				"id":          id,
				"title":       recipe.Title,
				"description": recipe.Description,
				"steps":       recipe.Steps,
				"ingredients": ingredientParams,
				"created":     neo4j.LocalDateTime(time.Now()),
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

			newIngredientIds := util.ArrayToSet(util.MapArray(recipe.Ingredients, model.ExtractIngredientId))
			existingIngredientIds := util.ArrayToSet(util.MapArray(existingRecipe.Ingredients, model.ExtractIngredientId))

			removedIngredientIds := util.Difference(existingIngredientIds, newIngredientIds)
			addedIngredientIds := util.Difference(newIngredientIds, existingIngredientIds)
			updatedIngredientIds := util.Intersection(existingIngredientIds, newIngredientIds)

			// check that newly added ingredients exist
			// probably no need to check removed or updated ingredients
			ingredientsExist, err := checkIngredientsExist(ctx, tx, addedIngredientIds)
			if err != nil {
				return nil, err
			}
			if !ingredientsExist {
				return nil, errors.New("tried to create a recipe with non-existent ingredient(s)")
			}

			removedIngredientParams := []map[string]string{}
			for _, ingredientId := range util.SetToArray(removedIngredientIds) {
				ingredient := map[string]string{"id": ingredientId}
				removedIngredientParams = append(removedIngredientParams, ingredient)
			}
			addedIngredientParams := []map[string]any{}
			for _, ingredientId := range util.SetToArray(addedIngredientIds) {
				ingredient := map[string]any{"id": ingredientId, "unit": newIngredients[ingredientId].Unit, "amount": newIngredients[ingredientId].Amount}
				addedIngredientParams = append(addedIngredientParams, ingredient)
			}
			updatedIngredientParams := []map[string]any{}
			for _, ingredientId := range util.SetToArray(updatedIngredientIds) {
				ingredient := map[string]any{"id": ingredientId, "unit": newIngredients[ingredientId].Unit, "amount": newIngredients[ingredientId].Amount}
				updatedIngredientParams = append(updatedIngredientParams, ingredient)
			}

			// TODO It may be possilbe to always include all 3 cases in the query:
			// https://neo4j.com/docs/cypher-manual/current/clauses/unwind/#unwind-using-unwind-with-an-empty-list
			// Using a plain UNWIND does not return any rows and will end the query execution early if e.g. there are no removed ingredients
			removeIngredientsStatement := fmt.Sprintf("WITH r UNWIND $removedIngredients AS ingredient\n"+
				"MATCH (r)-[ci:`%s`]->(:`%s` {id: ingredient.id}) SET ci.deleted = $lastModified\n",
				ContainsIngredientLabel, IngredientLabel,
			)
			addIngredientsStatement := fmt.Sprintf("WITH r UNWIND $addedIngredients AS ingredient\n"+
				"MATCH (i:`%s` {id: ingredient.id})\n"+
				"CREATE (r)-[:`%s` {unit: ingredient.unit, amount: ingredient.amount, created: $lastModified}]->(i)\n",
				IngredientLabel,
				ContainsIngredientLabel,
			)
			updateIngredientsStatement := fmt.Sprintf("WITH r UNWIND $updatedIngredients AS ingredient\n"+
				"MATCH (r)-[ci:`%s`]->(:`%s` {id: ingredient.id}) SET ci += {unit: ingredient.unit, amount: ingredient.amount, lastModified: $lastModified}\n",
				ContainsIngredientLabel, IngredientLabel,
			)

			*query = fmt.Sprintf("MATCH (r:`%s` {id: $id}) SET r += {title: $title, description: $description, steps: $steps, lastModified: $lastModified}\n",
				RecipeLabel,
			)
			if len(removedIngredientParams) > 0 {
				*query = *query + removeIngredientsStatement
			}
			if len(addedIngredientParams) > 0 {
				*query = *query + addIngredientsStatement
			}
			if len(updatedIngredientParams) > 0 {
				*query = *query + updateIngredientsStatement
			}
			*query = *query + fmt.Sprintf("WITH r MATCH (r)-[ci:`%s`]->(i:`%s`) WHERE ci.deleted IS NULL\n"+
				"RETURN r AS recipe, collect({ingredient: i, rel: ci}) AS ingredients",
				ContainsIngredientLabel, IngredientLabel,
			)

			params = map[string]any{
				"id":                 recipe.Id,
				"description":        recipe.Description,
				"title":              recipe.Title,
				"steps":              recipe.Steps,
				"removedIngredients": removedIngredientParams,
				"addedIngredients":   addedIngredientParams,
				"updatedIngredients": updatedIngredientParams,
				"lastModified":       neo4j.LocalDateTime(time.Now()),
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
