package controller

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ThomasMatlak/food/controller/request"
	"github.com/ThomasMatlak/food/controller/response"
	"github.com/ThomasMatlak/food/model"
	"github.com/gorilla/mux"
)

type RecipeController struct {
	recipeRepository model.RecipeRepository
}

func NewRecipeController(recipeRepository model.RecipeRepository) *RecipeController {
	return &RecipeController{recipeRepository: recipeRepository}
}

func (rc *RecipeController) RecipeRoutes(router *mux.Router) {
	reciperouter := router.PathPrefix("/recipe").Subrouter()

	reciperouter.HandleFunc("", rc.allRecipes).Methods("GET")
	reciperouter.HandleFunc("/{id}", rc.getRecipe).Methods("GET")
	reciperouter.HandleFunc("", rc.createRecipe).Methods("POST")
	reciperouter.HandleFunc("/{id}", rc.replaceRecipe).Methods("PUT")
	reciperouter.HandleFunc("/{id}", rc.updateRecipe).Methods("PATCH")
	reciperouter.HandleFunc("/{id}", rc.deleteRecipe).Methods("DELETE")
}

func (rc *RecipeController) allRecipes(w http.ResponseWriter, r *http.Request) {
	recipes, err := rc.recipeRepository.GetAll()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	response := response.GetRecipesResponse{Recipes: recipes}
	json.NewEncoder(w).Encode(response)
}

func (rc *RecipeController) getRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	recipe, err := rc.recipeRepository.GetById(id)
	if err != nil && err.Error() == "Result contains no more records" { // todo is there a better way to do this?
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(recipe)
}

func (rc *RecipeController) createRecipe(w http.ResponseWriter, r *http.Request) {
	var createRecipeRequest request.CreateRecipeRequest
	json.NewDecoder(r.Body).Decode(&createRecipeRequest)

	var newRecipe model.Recipe

	newRecipe.Title = strings.TrimSpace(createRecipeRequest.Title)
	*newRecipe.Description = strings.TrimSpace(*createRecipeRequest.Description)
	newRecipe.IngredientIds = createRecipeRequest.IngredientIds
	newRecipe.Steps = createRecipeRequest.Steps

	newRecipe.Created = new(time.Time)
	*newRecipe.Created = time.Now()

	recipe, err := rc.recipeRepository.Create(newRecipe)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(recipe)
}

func (rc *RecipeController) replaceRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	recipe, err := rc.recipeRepository.GetById(id)
	if err != nil && err.Error() == "Result contains no more records" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var replaceRecipeRequest request.CreateRecipeRequest
	json.NewDecoder(r.Body).Decode(&replaceRecipeRequest)

	recipe.Title = strings.TrimSpace(replaceRecipeRequest.Title)
	*recipe.Description = strings.TrimSpace(*replaceRecipeRequest.Description)
	recipe.IngredientIds = replaceRecipeRequest.IngredientIds
	recipe.Steps = replaceRecipeRequest.Steps

	recipe.LastModified = new(time.Time)
	*recipe.LastModified = time.Now()

	updatedRecipe, err := rc.recipeRepository.Update(*recipe)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(updatedRecipe)
}

func (rc *RecipeController) updateRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	recipe, err := rc.recipeRepository.GetById(id)
	if err != nil && err.Error() == "Result contains no more records" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var updateRecipeRequest request.UpdateRecipeRequest
	json.NewDecoder(r.Body).Decode(&updateRecipeRequest)

	if updateRecipeRequest.Title != nil {
		recipe.Title = strings.TrimSpace(*updateRecipeRequest.Title)
	}

	if updateRecipeRequest.Description != nil {
		*recipe.Description = strings.TrimSpace(*updateRecipeRequest.Description)
	}

	if updateRecipeRequest.IngredientIds != nil {
		recipe.IngredientIds = *updateRecipeRequest.IngredientIds
	}

	if updateRecipeRequest.Steps != nil {
		recipe.Steps = *updateRecipeRequest.Steps
	}

	recipe.LastModified = new(time.Time)
	*recipe.LastModified = time.Now()

	updatedRecipe, err := rc.recipeRepository.Update(*recipe)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(updatedRecipe)
}

func (rc *RecipeController) deleteRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	recipe, err := rc.recipeRepository.GetById(id)
	if err != nil && err.Error() == "Result contains no more records" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	deletedId, err := rc.recipeRepository.Delete(recipe.Id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	deleteRecipeResponse := response.DeleteRecipeResponse{Id: deletedId}
	json.NewEncoder(w).Encode(deleteRecipeResponse)
}
