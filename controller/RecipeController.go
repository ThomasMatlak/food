package controller

import (
	"encoding/json"
	"net/http"
	"strings"

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
	recipeRouter := router.PathPrefix("/recipe").Subrouter()

	// TODO search
	recipeRouter.HandleFunc("", rc.allRecipes).Methods("GET")
	recipeRouter.HandleFunc("/{id}", rc.getRecipe).Methods("GET")
	recipeRouter.HandleFunc("", rc.createRecipe).Methods("POST")
	recipeRouter.HandleFunc("/{id}", rc.replaceRecipe).Methods("PUT")
	recipeRouter.HandleFunc("/{id}", rc.updateRecipe).Methods("PATCH")
	recipeRouter.HandleFunc("/{id}", rc.deleteRecipe).Methods("DELETE")
}

func (rc *RecipeController) allRecipes(w http.ResponseWriter, r *http.Request) {
	recipes, err := rc.recipeRepository.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// TODO pagination
	response := response.GetRecipesResponse{Recipes: recipes}
	json.NewEncoder(w).Encode(response)

}

func (rc *RecipeController) getRecipe(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]

	recipe, found, err := rc.recipeRepository.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(recipe)
}

func (rc *RecipeController) createRecipe(w http.ResponseWriter, r *http.Request) {
	var createRecipeRequest request.CreateRecipeRequest
	json.NewDecoder(r.Body).Decode(&createRecipeRequest)

	if !request.CanCreateRecipe(&createRecipeRequest) {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	var newRecipe model.Recipe

	newRecipe.Title = strings.TrimSpace(createRecipeRequest.Title)
	if createRecipeRequest.Description != nil {
		*newRecipe.Description = strings.TrimSpace(*createRecipeRequest.Description)
	}
	newRecipe.Ingredients = createRecipeRequest.Ingredients
	newRecipe.Steps = createRecipeRequest.Steps

	recipe, err := rc.recipeRepository.Create(r.Context(), newRecipe)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(recipe)
}

func (rc *RecipeController) replaceRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	recipe, found, err := rc.recipeRepository.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	var replaceRecipeRequest request.CreateRecipeRequest
	json.NewDecoder(r.Body).Decode(&replaceRecipeRequest)

	if !request.CanCreateRecipe(&replaceRecipeRequest) {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	recipe.Title = strings.TrimSpace(replaceRecipeRequest.Title)
	if replaceRecipeRequest.Description != nil {
		*recipe.Description = strings.TrimSpace(*replaceRecipeRequest.Description)
	} else {
		recipe.Description = nil
	}
	recipe.Ingredients = replaceRecipeRequest.Ingredients
	recipe.Steps = replaceRecipeRequest.Steps

	updatedRecipe, err := rc.recipeRepository.Update(r.Context(), *recipe)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(updatedRecipe)
}

func (rc *RecipeController) updateRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	recipe, found, err := rc.recipeRepository.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	var updateRecipeRequest request.UpdateRecipeRequest
	json.NewDecoder(r.Body).Decode(&updateRecipeRequest)

	if !request.CanUpdateRecipe(&updateRecipeRequest) {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	if updateRecipeRequest.Title != nil {
		recipe.Title = strings.TrimSpace(*updateRecipeRequest.Title)
	}

	if updateRecipeRequest.Description != nil {
		*recipe.Description = strings.TrimSpace(*updateRecipeRequest.Description)
	}

	if updateRecipeRequest.Ingredients != nil {
		recipe.Ingredients = *updateRecipeRequest.Ingredients
	}

	if updateRecipeRequest.Steps != nil {
		recipe.Steps = *updateRecipeRequest.Steps
	}

	updatedRecipe, err := rc.recipeRepository.Update(r.Context(), *recipe)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(updatedRecipe)
}

func (rc *RecipeController) deleteRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	recipe, found, err := rc.recipeRepository.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	deletedId, err := rc.recipeRepository.Delete(r.Context(), recipe.Id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	deleteRecipeResponse := response.DeleteRecipeResponse{Id: deletedId}
	json.NewEncoder(w).Encode(deleteRecipeResponse)
}
