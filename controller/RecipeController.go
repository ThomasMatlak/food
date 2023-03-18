package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ThomasMatlak/food/controller/request"
	"github.com/ThomasMatlak/food/controller/response"
	"github.com/ThomasMatlak/food/model"
	"github.com/ThomasMatlak/food/repository"
	"github.com/gorilla/mux"
)

func allRecipes(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func getRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var recipe model.Recipe
	recipe.Id = id

	json.NewEncoder(w).Encode(recipe)
}

func createRecipe(w http.ResponseWriter, r *http.Request) {
	var createRecipeRequest request.CreateRecipeRequest
	json.NewDecoder(r.Body).Decode(&createRecipeRequest)

	var newRecipe model.Recipe

	newRecipe.Title = createRecipeRequest.Title
	newRecipe.IngredientIds = createRecipeRequest.IngredientIds
	newRecipe.Steps = createRecipeRequest.Steps

	newRecipe.Created = new(time.Time)
	*newRecipe.Created = time.Now()

	recipe, err := repository.CreateRecipe(newRecipe)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(recipe)
}

func replaceRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var replaceRecipeRequest request.CreateRecipeRequest
	json.NewDecoder(r.Body).Decode(&replaceRecipeRequest)

	var recipe model.Recipe
	recipe.Id = id

	recipe.Title = replaceRecipeRequest.Title
	recipe.IngredientIds = replaceRecipeRequest.IngredientIds
	recipe.Steps = replaceRecipeRequest.Steps

	recipe.LastModified = new(time.Time)
	*recipe.LastModified = time.Now()

	json.NewEncoder(w).Encode(recipe)
}

func updateRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var updateRecipeRequest request.UpdateRecipeRequest
	json.NewDecoder(r.Body).Decode(&updateRecipeRequest)

	var recipe model.Recipe
	recipe.Id = id

	if updateRecipeRequest.Title != nil {
		recipe.Title = *updateRecipeRequest.Title
	}

	if updateRecipeRequest.IngredientIds != nil {
		recipe.IngredientIds = *updateRecipeRequest.IngredientIds
	}

	if updateRecipeRequest.Steps != nil {
		recipe.Steps = *updateRecipeRequest.Steps
	}

	recipe.LastModified = new(time.Time)
	*recipe.LastModified = time.Now()

	json.NewEncoder(w).Encode(recipe)
}

func deleteRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var deleteRecipeResponse response.DeleteRecipeResponse
	deleteRecipeResponse.Id = id

	json.NewEncoder(w).Encode(deleteRecipeResponse)
}

func RecipeRoutes(router *mux.Router) {
	reciperouter := router.PathPrefix("/recipe").Subrouter()

	reciperouter.HandleFunc("", allRecipes).Methods("GET")
	reciperouter.HandleFunc("/{id}", getRecipe).Methods("GET")
	reciperouter.HandleFunc("", createRecipe).Methods("POST")
	reciperouter.HandleFunc("/{id}", replaceRecipe).Methods("PUT")
	reciperouter.HandleFunc("/{id}", updateRecipe).Methods("PATCH")
	reciperouter.HandleFunc("/{id}", deleteRecipe).Methods("DELETE")
}
