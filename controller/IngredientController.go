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

type IngredientController struct {
	ingredientRepository model.IngredientRepository
}

func NewIngredientController(ingredientRepository model.IngredientRepository) *IngredientController {
	return &IngredientController{ingredientRepository: ingredientRepository}
}

func (rc *IngredientController) IngredientRoutes(router *mux.Router) {
	ingredientRouter := router.PathPrefix("/ingredient").Subrouter()

	// TODO search
	ingredientRouter.HandleFunc("", rc.allIngredients).Methods("GET")
	ingredientRouter.HandleFunc("/{id}", rc.getIngredient).Methods("GET")
	ingredientRouter.HandleFunc("", rc.createIngredient).Methods("POST")
	ingredientRouter.HandleFunc("/{id}", rc.replaceIngredient).Methods("PUT")
	ingredientRouter.HandleFunc("/{id}", rc.updateIngredient).Methods("PATCH")
	ingredientRouter.HandleFunc("/{id}", rc.deleteIngredient).Methods("DELETE")
}

func (ic *IngredientController) allIngredients(w http.ResponseWriter, r *http.Request) {
	ingredients, err := ic.ingredientRepository.GetAll()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// TODO pagination
	response := response.GetIngredientsResponse{Ingredients: ingredients}
	json.NewEncoder(w).Encode(response)
}

func (ic *IngredientController) getIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	ingredient, found, err := ic.ingredientRepository.GetById(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(ingredient)
}

func (ic *IngredientController) createIngredient(w http.ResponseWriter, r *http.Request) {
	var createIngredientRequest request.CreateIngredientRequest
	json.NewDecoder(r.Body).Decode(&createIngredientRequest)

	var newIngredient model.Ingredient
	newIngredient.Name = strings.TrimSpace(createIngredientRequest.Name)

	newIngredient.Created = new(time.Time)
	*newIngredient.Created = time.Now()

	ingredient, err := ic.ingredientRepository.Create(newIngredient)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(ingredient)
}

func (ic *IngredientController) replaceIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	ingredient, found, err := ic.ingredientRepository.GetById(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	var replaceIngredientRequest request.CreateIngredientRequest
	json.NewDecoder(r.Body).Decode(&replaceIngredientRequest)

	ingredient.Name = strings.TrimSpace(replaceIngredientRequest.Name)

	ingredient.LastModified = new(time.Time)
	*ingredient.LastModified = time.Now()

	updatedIngredient, err := ic.ingredientRepository.Update(*ingredient)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(updatedIngredient)
}

func (ic *IngredientController) updateIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	ingredient, found, err := ic.ingredientRepository.GetById(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	var replaceIngredientRequest request.UpdateIngredientRequest
	json.NewDecoder(r.Body).Decode(&replaceIngredientRequest)

	if replaceIngredientRequest.Name != nil {
		ingredient.Name = strings.TrimSpace(*replaceIngredientRequest.Name)
	}

	ingredient.LastModified = new(time.Time)
	*ingredient.LastModified = time.Now()

	updatedIngredient, err := ic.ingredientRepository.Update(*ingredient)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(updatedIngredient)
}

func (ic *IngredientController) deleteIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	ingredient, found, err := ic.ingredientRepository.GetById(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	deletedId, err := ic.ingredientRepository.Delete(ingredient.Id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	deleteIngredientResponse := response.DeleteIngredientResponse{Id: deletedId}
	json.NewEncoder(w).Encode(deleteIngredientResponse)
}
