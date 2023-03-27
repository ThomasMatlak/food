package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ThomasMatlak/food/model"
	"github.com/ThomasMatlak/food/repository"
	"github.com/ThomasMatlak/food/util"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
)

func TestRecipeRepository(t *testing.T) {
	ctx := context.Background()

	neo4jContainer, err := startNeo4j(ctx, t)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	neo4jDriver, err := neo4jDriver(ctx, t, neo4jContainer)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	repo := repository.NewRecipeRepository(*neo4jDriver)

	t.Run("Get One", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testGetOneRecipe(ctx, neo4jDriver, repo, t)
	})
	t.Run("Get One (does not exist)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testGetOneDoesNotExistRecipe(ctx, neo4jDriver, repo, t)
	})
	t.Run("Get All", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testGetAllRecipes(ctx, neo4jDriver, repo, t)
	})
	t.Run("Get All (empty)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testGetAllEmptyRecipe(ctx, neo4jDriver, repo, t)
	})
	t.Run("Create", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testCreateRecipe(ctx, neo4jDriver, repo, t)
	})
	t.Run("Create (no description)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testCreateRecipeNoDescription(ctx, neo4jDriver, repo, t)
	})
	t.Run("Create (one ingredient not found)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testCreateRecipeOneIngredientNotFound(ctx, neo4jDriver, repo, t)
	})
	t.Run("Create (no ingredients found)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testCreateRecipeNoIngredientsFound(ctx, neo4jDriver, repo, t)
	})
	t.Run("Update (ingredient list unchanged)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testUpdateRecipeNoIngredientsChanged(ctx, neo4jDriver, repo, t)
	})
	t.Run("Update (ingredient added)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testUpdateRecipeIngredientAdded(ctx, neo4jDriver, repo, t)
	})
	t.Run("Update (ingredient removed)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testUpdateRecipeIngredientRemoved(ctx, neo4jDriver, repo, t)
	})
	t.Run("Update (one ingredient kept, one added, one removed)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testUpdateRecipeIngredientsKeptAddedRemoved(ctx, neo4jDriver, repo, t)
	})
	t.Run("Update (ingredient that was previously deleted added again)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testUpdateRecipeReaddIngredient(ctx, neo4jDriver, repo, t)
	})
	t.Run("Update (one ingredient not found)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testUpdateRecipeOneIngredientNotFound(ctx, neo4jDriver, repo, t)
	})
	t.Run("Update (no ingredients found)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testUpdateRecipeNoIngredientsFound(ctx, neo4jDriver, repo, t)
	})
	// TODO when it is a deleted ingredient that cannot be found, should there be an error?
	t.Run("Delete", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testDeleteRecipe(ctx, neo4jDriver, repo, t)
	})
}

var seedIngredientsAndRecipes string = `UNWIND $ingredients AS ingredient
MERGE (i:Ingredient {id: ingredient.id}) SET i+= {name: ingredient.name, created: $created}
WITH '' AS throwaway
UNWIND $recipes AS recipe
MERGE (r:Recipe {id : recipe.id}) SET r += {title: recipe.title, description: recipe.description, steps: recipe.steps, created: $created}
WITH recipe, r UNWIND recipe.ingredients AS ingredient
MATCH (i:Ingredient {id: ingredient.ingredient_id})
MERGE (r)-[rel:CONTAINS_INGREDIENT]->(i) SET rel = {unit: ingredient.unit, amount: ingredient.amount, created: $created}`

func extractIngredientId(ci model.ContainsIngredient) string { return ci.IngredientId }

func testGetOneRecipe(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// seed data
	recipeId := "test recipe id"
	recipeTitle := "onion"
	recipeDesc := "just a raw onion"
	ingredientId := "test ingredient id"
	steps := []string{"peel the onion", "eat the onion like an apple", "enjoy! :)"}

	query := "CREATE (:Recipe {id: $id, title: $title, description: $description, steps: $steps, created: $created})-[:CONTAINS_INGREDIENT {unit: $unit, amount: $amount, created: $created}]->(:Ingredient {id: $ingredientId, name: 'onion', created: $created})"
	createdTime := time.Now()
	params := map[string]any{
		"id":           recipeId,
		"title":        recipeTitle,
		"description":  recipeDesc,
		"steps":        steps,
		"unit":         "whole",
		"amount":       1,
		"ingredientId": ingredientId,
		"created":      neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, query, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	recipe, found, err := repo.GetById(ctx, recipeId)

	expetedIngredientIds := []string{ingredientId}
	actualIngredientIds := util.MapArray(recipe.Ingredients, func(ci model.ContainsIngredient) string { return ci.IngredientId })

	assert := assert.New(t)
	assert.NoError(err)
	assert.True(found)
	assert.Equal(recipeId, recipe.Id)
	assert.Equal(recipeTitle, recipe.Title)
	assert.Equal(recipeDesc, *recipe.Description)
	assert.Equal(expetedIngredientIds, actualIngredientIds)
	assert.Equal(len(expetedIngredientIds), len(actualIngredientIds))
	assert.Equal(steps, recipe.Steps)
	assert.WithinDuration(createdTime, *recipe.Created, 0)
	assert.Nil(recipe.LastModified)
	assert.Nil(recipe.Deleted)
}

func testGetOneDoesNotExistRecipe(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// no seed data

	// test
	recipe, found, err := repo.GetById(ctx, "test id")

	assert := assert.New(t)
	assert.NoError(err)
	assert.False(found)
	assert.Nil(recipe)
}

func testGetAllRecipes(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// seed data
	seedIngredients := []map[string]any{
		{"id": "123", "name": "test ingredient 1"},
		{"id": "456", "name": "test ingredient 2"},
		{"id": "789", "name": "test ingredient 3"},
	}

	seedRecipes := []map[string]any{
		{"id": "1", "title": "test recipe 1", "description": "tastes alright", "steps": []string{"cook it"},
			"ingredients": []map[string]any{{"unit": "g", "amount": 15, "ingredient_id": "123"}},
		},
		{"id": "2", "title": "test recipe 2", "steps": []string{"mix", "cook"}, "ingredients": []map[string]any{
			{"unit": "cup", "amount": 1, "ingredient_id": "456"},
			{"unit": "oz", "amount": 42, "ingredient_id": "789"},
		}},
	}

	createdTime := time.Now()
	params := map[string]any{
		"ingredients": seedIngredients,
		"recipes":     seedRecipes,
		"created":     neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, seedIngredientsAndRecipes, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	recipes, err := repo.GetAll(ctx)

	assert := assert.New(t)
	assert.NoError(err)
	// comparing time stamps is tricky
	fmt.Println(recipes)
}

func testGetAllEmptyRecipe(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// no seed data

	// test
	recipes, err := repo.GetAll(ctx)

	assert := assert.New(t)
	assert.NoError(err)
	assert.Empty([]model.Ingredient{}, recipes)
}

func testCreateRecipe(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// seed data
	query := "UNWIND $ingredients AS i CREATE (:Ingredient {id: i.id, name: i.name, created: $created})"
	ingredientParams := []map[string]string{
		{"id": "asdf", "name": "rice"},
		{"id": "zxcv", "name": "beans"},
	}
	createdTime := time.Now()
	params := map[string]any{
		"ingredients": ingredientParams,
		"created":     neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, query, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	title := "test recipe"
	description := "test description"
	ingredients := []model.ContainsIngredient{
		{Unit: "cup", Amount: 1, IngredientId: "asdf"},
		{Unit: "cup", Amount: 1, IngredientId: "zxcv"},
	}
	steps := []string{"cook beans", "cook rice", "combine cooked beans and rice"}
	recipe := model.Recipe{Title: title, Description: &description, Ingredients: ingredients, Steps: steps}
	createdRecipe, err := repo.Create(ctx, recipe)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.NoError(err)
	assert.NotEmpty(createdRecipe.Id)
	assert.Equal(title, createdRecipe.Title)
	assert.Equal(description, *createdRecipe.Description)
	assert.ElementsMatch(util.MapArray(ingredients, extractIngredientId), util.MapArray(createdRecipe.Ingredients, extractIngredientId))
	assert.WithinDuration(time.Now(), *createdRecipe.Created, time.Duration(1_000_000_000))
	assert.Nil(createdRecipe.LastModified)
	assert.Nil(createdRecipe.Deleted)
}

func testCreateRecipeNoDescription(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// seed data
	query := "UNWIND $ingredients AS i CREATE (:Ingredient {id: i.id, name: i.name, created: $created})"
	ingredientParams := []map[string]string{
		{"id": "asdf", "name": "rice"},
		{"id": "zxcv", "name": "beans"},
	}
	createdTime := time.Now()
	params := map[string]any{
		"ingredients": ingredientParams,
		"created":     neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, query, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	title := "test recipe"
	ingredients := []model.ContainsIngredient{
		{Unit: "cup", Amount: 1, IngredientId: "asdf"},
		{Unit: "cup", Amount: 1, IngredientId: "zxcv"},
	}
	steps := []string{"cook beans", "cook rice", "combine cooked beans and rice"}
	recipe := model.Recipe{Title: title, Ingredients: ingredients, Steps: steps}
	createdRecipe, err := repo.Create(ctx, recipe)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.NoError(err)
	assert.NotEmpty(createdRecipe.Id)
	assert.Equal(title, createdRecipe.Title)
	assert.Nil(createdRecipe.Description)
	assert.ElementsMatch(util.MapArray(ingredients, extractIngredientId), util.MapArray(createdRecipe.Ingredients, extractIngredientId))
	assert.WithinDuration(time.Now(), *createdRecipe.Created, time.Duration(1_000_000_000))
	assert.Nil(createdRecipe.LastModified)
	assert.Nil(createdRecipe.Deleted)
}

func testCreateRecipeOneIngredientNotFound(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// seed data
	query := "UNWIND $ingredients AS i CREATE (:Ingredient {id: i.id, name: i.name, created: $created})"
	ingredientParams := []map[string]string{
		{"id": "asdf", "name": "rice"},
	}
	createdTime := time.Now()
	params := map[string]any{
		"ingredients": ingredientParams,
		"created":     neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, query, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	title := "test recipe"
	description := "test description"
	ingredients := []model.ContainsIngredient{
		{Unit: "cup", Amount: 1, IngredientId: "asdf"},
		{Unit: "cup", Amount: 1, IngredientId: "zxcv"},
	}
	steps := []string{"cook beans", "cook rice", "combine cooked beans and rice"}
	recipe := model.Recipe{Title: title, Description: &description, Ingredients: ingredients, Steps: steps}
	createdRecipe, err := repo.Create(ctx, recipe)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.Error(err)
	assert.Nil(createdRecipe)
}

func testCreateRecipeNoIngredientsFound(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// no seed data since we don't want to match any ingredients
	// test
	title := "test recipe"
	description := "test description"
	ingredients := []model.ContainsIngredient{
		{Unit: "cup", Amount: 1, IngredientId: "asdf"},
		{Unit: "cup", Amount: 1, IngredientId: "zxcv"},
	}
	steps := []string{"cook beans", "cook rice", "combine cooked beans and rice"}
	recipe := model.Recipe{Title: title, Description: &description, Ingredients: ingredients, Steps: steps}
	createdRecipe, err := repo.Create(ctx, recipe)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.Error(err)
	assert.Nil(createdRecipe)
}

func testUpdateRecipeNoIngredientsChanged(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// seed data
	id := "1"

	seedIngredients := []map[string]any{
		{"id": "123", "name": "test ingredient"},
	}

	seedRecipes := []map[string]any{
		{"id": id, "title": "test recipe", "description": "tastes alright", "steps": []string{"cook it"},
			"ingredients": []map[string]any{{"unit": "g", "amount": 15, "ingredient_id": "123"}}},
	}

	createdTime := time.Now()
	params := map[string]any{
		"ingredients": seedIngredients,
		"recipes":     seedRecipes,
		"created":     neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, seedIngredientsAndRecipes, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	desc := "tastes okay"
	recipe := model.Recipe{Id: id, Title: "test recipe updated", Description: &desc, Steps: []string{"do prep work", "cook it"}, Ingredients: []model.ContainsIngredient{
		{Unit: "g", Amount: 15, IngredientId: "123"},
	}}
	updatedRecipe, err := repo.Update(ctx, recipe)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(id, updatedRecipe.Id)
	assert.Equal("test recipe updated", updatedRecipe.Title)
	assert.Equal("tastes okay", *updatedRecipe.Description)
	assert.Equal([]string{"do prep work", "cook it"}, updatedRecipe.Steps)
	assert.ElementsMatch([]string{"123"}, util.MapArray(updatedRecipe.Ingredients, extractIngredientId))
	assert.WithinDuration(createdTime, *updatedRecipe.Created, 0)
	assert.True((*updatedRecipe.LastModified).After(createdTime))
	assert.Nil(updatedRecipe.Deleted)
}

func testUpdateRecipeIngredientAdded(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// seed data
	id := "1"

	seedIngredients := []map[string]any{
		{"id": "123", "name": "test ingredient 1"},
		{"id": "456", "name": "test ingredient 2"},
	}

	seedRecipes := []map[string]any{
		{"id": id, "title": "test recipe", "description": "tastes alright", "steps": []string{"cook it"},
			"ingredients": []map[string]any{{"unit": "g", "amount": 15, "ingredient_id": "123"}}},
	}

	createdTime := time.Now()
	params := map[string]any{
		"ingredients": seedIngredients,
		"recipes":     seedRecipes,
		"created":     neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, seedIngredientsAndRecipes, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	desc := "tastes okay"
	recipe := model.Recipe{Id: id, Title: "test recipe updated", Description: &desc, Steps: []string{"do prep work", "cook it"}, Ingredients: []model.ContainsIngredient{
		{Unit: "g", Amount: 15, IngredientId: "123"},
		{Unit: "cup", Amount: 1, IngredientId: "456"},
	}}
	updatedRecipe, err := repo.Update(ctx, recipe)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(id, updatedRecipe.Id)
	assert.Equal("test recipe updated", updatedRecipe.Title)
	assert.Equal("tastes okay", *updatedRecipe.Description)
	assert.Equal([]string{"do prep work", "cook it"}, updatedRecipe.Steps)
	assert.ElementsMatch([]string{"123", "456"}, util.MapArray(updatedRecipe.Ingredients, extractIngredientId))
	assert.WithinDuration(createdTime, *updatedRecipe.Created, 0)
	assert.True((*updatedRecipe.LastModified).After(createdTime))
	assert.Nil(updatedRecipe.Deleted)
}

func testUpdateRecipeIngredientRemoved(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// seed data
	id := "1"

	seedIngredients := []map[string]any{
		{"id": "123", "name": "test ingredient 1"},
		{"id": "456", "name": "test ingredient 2"},
	}

	seedRecipes := []map[string]any{
		{"id": id, "title": "test recipe", "description": "tastes alright", "steps": []string{"cook it"},
			"ingredients": []map[string]any{{"unit": "g", "amount": 15, "ingredient_id": "123"}, {"unit": "cup", "amount": 1, "ingredient_id": "456"}}},
	}

	createdTime := time.Now()
	params := map[string]any{
		"ingredients": seedIngredients,
		"recipes":     seedRecipes,
		"created":     neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, seedIngredientsAndRecipes, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	desc := "tastes okay"
	recipe := model.Recipe{Id: id, Title: "test recipe updated", Description: &desc, Steps: []string{"do prep work", "cook it"}, Ingredients: []model.ContainsIngredient{
		{Unit: "g", Amount: 15, IngredientId: "123"},
	}}
	updatedRecipe, err := repo.Update(ctx, recipe)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(id, updatedRecipe.Id)
	assert.Equal("test recipe updated", updatedRecipe.Title)
	assert.Equal("tastes okay", *updatedRecipe.Description)
	assert.Equal([]string{"do prep work", "cook it"}, updatedRecipe.Steps)
	assert.ElementsMatch([]string{"123"}, util.MapArray(updatedRecipe.Ingredients, extractIngredientId))
	assert.WithinDuration(createdTime, *updatedRecipe.Created, 0)
	assert.True((*updatedRecipe.LastModified).After(createdTime))
	assert.Nil(updatedRecipe.Deleted)
}

func testUpdateRecipeIngredientsKeptAddedRemoved(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// seed data
	id := "1"

	seedIngredients := []map[string]any{
		{"id": "123", "name": "test ingredient 1"},
		{"id": "456", "name": "test ingredient 2"},
		{"id": "789", "name": "test ingredient 3"},
	}

	seedRecipes := []map[string]any{
		{"id": id, "title": "test recipe", "description": "tastes alright", "steps": []string{"cook it"},
			"ingredients": []map[string]any{{"unit": "g", "amount": 15, "ingredient_id": "123"}, {"unit": "cup", "amount": 1, "ingredient_id": "456"}}},
	}

	createdTime := time.Now()
	params := map[string]any{
		"ingredients": seedIngredients,
		"recipes":     seedRecipes,
		"created":     neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, seedIngredientsAndRecipes, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	desc := "tastes okay"
	recipe := model.Recipe{Id: id, Title: "test recipe updated", Description: &desc, Steps: []string{"do prep work", "cook it"}, Ingredients: []model.ContainsIngredient{
		{Unit: "g", Amount: 15, IngredientId: "123"}, {Unit: "oz", Amount: 12, IngredientId: "789"},
	}}
	updatedRecipe, err := repo.Update(ctx, recipe)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(id, updatedRecipe.Id)
	assert.Equal("test recipe updated", updatedRecipe.Title)
	assert.Equal("tastes okay", *updatedRecipe.Description)
	assert.Equal([]string{"do prep work", "cook it"}, updatedRecipe.Steps)
	assert.ElementsMatch([]string{"123", "789"}, util.MapArray(updatedRecipe.Ingredients, extractIngredientId))
	assert.WithinDuration(createdTime, *updatedRecipe.Created, 0)
	assert.True((*updatedRecipe.LastModified).After(createdTime))
	assert.Nil(updatedRecipe.Deleted)
}

func testUpdateRecipeReaddIngredient(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// seed data
	id := "1"

	seedIngredients := []map[string]any{
		{"id": "123", "name": "test ingredient 1"},
		{"id": "456", "name": "test ingredient 2"},
	}

	seedRecipes := []map[string]any{
		{"id": id, "title": "test recipe", "description": "tastes alright", "steps": []string{"cook it"},
			"ingredients": []map[string]any{{"unit": "g", "amount": 15, "ingredient_id": "123"}, {"unit": "cup", "amount": 1, "ingredient_id": "456"}}},
	}

	createdTime := time.Now()
	params := map[string]any{
		"ingredients": seedIngredients,
		"recipes":     seedRecipes,
		"created":     neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, seedIngredientsAndRecipes, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	_, err = neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			query := "MATCH (:Recipe {id: $recipeId})-[rel:CONTAINS_INGREDIENT]->(:Ingredient {id: $ingredientId}) SET rel.deleted = $deleted"
			params := map[string]any{"recipeId": id, "ingredientId": "456", "deleted": neo4j.LocalDateTime(time.Now())}
			return tx.Run(ctx, query, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	desc := "tastes okay"
	recipe := model.Recipe{Id: id, Title: "test recipe updated", Description: &desc, Steps: []string{"do prep work", "cook it"}, Ingredients: []model.ContainsIngredient{
		{Unit: "g", Amount: 15, IngredientId: "123"},
		{Unit: "cup", Amount: 1, IngredientId: "456"},
	}}
	updatedRecipe, err := repo.Update(ctx, recipe)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(id, updatedRecipe.Id)
	assert.Equal("test recipe updated", updatedRecipe.Title)
	assert.Equal("tastes okay", *updatedRecipe.Description)
	assert.Equal([]string{"do prep work", "cook it"}, updatedRecipe.Steps)
	assert.ElementsMatch([]string{"123", "456"}, util.MapArray(updatedRecipe.Ingredients, extractIngredientId))
	assert.WithinDuration(createdTime, *updatedRecipe.Created, 0)
	assert.True((*updatedRecipe.LastModified).After(createdTime))
	assert.Nil(updatedRecipe.Deleted)
}

func testUpdateRecipeOneIngredientNotFound(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// seed data
	id := "1"

	seedIngredients := []map[string]any{
		{"id": "123", "name": "test ingredient 1"},
	}

	seedRecipes := []map[string]any{
		{"id": id, "title": "test recipe", "description": "tastes alright", "steps": []string{"cook it"},
			"ingredients": []map[string]any{{"unit": "g", "amount": 15, "ingredient_id": "123"}}},
	}

	createdTime := time.Now()
	params := map[string]any{
		"ingredients": seedIngredients,
		"recipes":     seedRecipes,
		"created":     neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, seedIngredientsAndRecipes, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	desc := "tastes okay"
	recipe := model.Recipe{Id: id, Title: "test recipe updated", Description: &desc, Steps: []string{"do prep work", "cook it"}, Ingredients: []model.ContainsIngredient{
		{Unit: "g", Amount: 15, IngredientId: "123"},
		{Unit: "cup", Amount: 1, IngredientId: "456"},
	}}
	updatedRecipe, err := repo.Update(ctx, recipe)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.Error(err)
	assert.Nil(updatedRecipe)
}

func testUpdateRecipeNoIngredientsFound(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// seed data
	id := "1"

	seedIngredients := []map[string]any{
		{"id": "123", "name": "test ingredient 1"},
	}

	seedRecipes := []map[string]any{
		{"id": id, "title": "test recipe", "description": "tastes alright", "steps": []string{"cook it"},
			"ingredients": []map[string]any{{"unit": "g", "amount": 15, "ingredient_id": "123"}}},
	}

	createdTime := time.Now()
	params := map[string]any{
		"ingredients": seedIngredients,
		"recipes":     seedRecipes,
		"created":     neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, seedIngredientsAndRecipes, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	desc := "tastes okay"
	recipe := model.Recipe{Id: id, Title: "test recipe updated", Description: &desc, Steps: []string{"do prep work", "cook it"}, Ingredients: []model.ContainsIngredient{
		{Unit: "cup", Amount: 1, IngredientId: "456"},
		{Unit: "oz", Amount: 1, IngredientId: "789"},
	}}
	updatedRecipe, err := repo.Update(ctx, recipe)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.Error(err)
	assert.Nil(updatedRecipe)
}

func testDeleteRecipe(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.RecipeRepository, t *testing.T) {
	// seed data
	id := "123"

	query := "CREATE (:Recipe {id: $id, title: $title, created: $created})-[:CONTAINS_INGREDIENT {created: $created}]->(:Ingredient:Resource {created: $created})"
	createdTime := time.Now()
	params := map[string]any{
		"id":      id,
		"title":   "test recipe",
		"created": neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, query, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	deletedId, err := repo.Delete(ctx, id)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(id, deletedId)
}
