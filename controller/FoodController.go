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

type FoodController struct {
	foodRepository model.FoodRepository
}

func NewFoodController(foodRepository model.FoodRepository) *FoodController {
	return &FoodController{foodRepository: foodRepository}
}

func (rc *FoodController) FoodRoutes(router *mux.Router) {
	foodRouter := router.PathPrefix("/food").Subrouter()

	// TODO search
	foodRouter.HandleFunc("", rc.allFoods).Methods("GET")
	foodRouter.HandleFunc("/{id}", rc.getFood).Methods("GET")
	foodRouter.HandleFunc("", rc.createFood).Methods("POST")
	foodRouter.HandleFunc("/{id}", rc.replaceFood).Methods("PUT")
	foodRouter.HandleFunc("/{id}", rc.updateFood).Methods("PATCH")
	foodRouter.HandleFunc("/{id}", rc.deleteFood).Methods("DELETE")
}

func (ic *FoodController) allFoods(w http.ResponseWriter, r *http.Request) {
	foods, err := ic.foodRepository.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// TODO pagination
	response := response.GetFoodsResponse{Foods: foods}
	json.NewEncoder(w).Encode(response)
}

func (ic *FoodController) getFood(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	food, found, err := ic.foodRepository.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(food)
}

func (ic *FoodController) createFood(w http.ResponseWriter, r *http.Request) {
	var createFoodRequest request.CreateFoodRequest
	json.NewDecoder(r.Body).Decode(&createFoodRequest)

	var newFood model.Food
	newFood.Name = strings.TrimSpace(createFoodRequest.Name)

	food, err := ic.foodRepository.Create(r.Context(), newFood)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(food)
}

func (ic *FoodController) replaceFood(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	food, found, err := ic.foodRepository.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	var replaceFoodRequest request.CreateFoodRequest
	json.NewDecoder(r.Body).Decode(&replaceFoodRequest)

	food.Name = strings.TrimSpace(replaceFoodRequest.Name)

	updatedFood, err := ic.foodRepository.Update(r.Context(), *food)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(updatedFood)
}

func (ic *FoodController) updateFood(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	food, found, err := ic.foodRepository.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	var replaceFoodRequest request.UpdateFoodRequest
	json.NewDecoder(r.Body).Decode(&replaceFoodRequest)

	if replaceFoodRequest.Name != nil {
		food.Name = strings.TrimSpace(*replaceFoodRequest.Name)
	}

	updatedFood, err := ic.foodRepository.Update(r.Context(), *food)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(updatedFood)
}

func (ic *FoodController) deleteFood(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	food, found, err := ic.foodRepository.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	deletedId, err := ic.foodRepository.Delete(r.Context(), food.Id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	deleteFoodResponse := response.DeleteFoodResponse{Id: deletedId}
	json.NewEncoder(w).Encode(deleteFoodResponse)
}
