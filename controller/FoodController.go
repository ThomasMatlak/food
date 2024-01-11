package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ThomasMatlak/food/controller/request"
	"github.com/ThomasMatlak/food/controller/response"
	"github.com/ThomasMatlak/food/model"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

type FoodController struct {
	foodRepository model.FoodRepository
}

func NewFoodController(foodRepository model.FoodRepository) *FoodController {
	return &FoodController{foodRepository: foodRepository}
}

func (ic *FoodController) FoodRoutes(router chi.Router) {
	router.Route("/food", func(r chi.Router) {
		// TODO search
		r.Post("/", ic.createFood)
		r.Get("/", ic.allFoods)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", ic.getFood)
			r.Put("/", ic.replaceFood)
			r.Patch("/", ic.updateFood)
			r.Delete("/", ic.deleteFood)
		})

		r.Route("/create", func(r chi.Router) {
			r.Get("/", ic.createFoodForm)
		})
		r.Route("/{id}/edit", func(r chi.Router) {
			r.Get("/", ic.editFoodForm)
		})
	})
}

func (ic *FoodController) createFoodForm(w http.ResponseWriter, r *http.Request) {
	component := response.CreateFood()
	templ.Handler(component).ServeHTTP(w, r)
}

func (ic *FoodController) editFoodForm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	food, found, err := ic.foodRepository.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	component := response.EditFoodForm(food)
	templ.Handler(component).ServeHTTP(w, r)
}

func (ic *FoodController) allFoods(w http.ResponseWriter, r *http.Request) {
	foods, err := ic.foodRepository.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// TODO pagination
	if r.Header.Get("Accept") == "application/json" {
		response := response.GetFoodsResponse{Foods: foods}
		json.NewEncoder(w).Encode(response)
	} else {
		templ.Handler(response.ViewFoods(foods)).ServeHTTP(w, r)
	}
}

func (ic *FoodController) getFood(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	food, found, err := ic.foodRepository.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		json.NewEncoder(w).Encode(food)
	} else {
		templ.Handler(response.GetFood(food)).ServeHTTP(w, r)
	}
}

func (ic *FoodController) createFood(w http.ResponseWriter, r *http.Request) {
	var createFoodRequest request.CreateFoodRequest

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	form := r.Form

	if len(form) == 0 {
		json.NewDecoder(r.Body).Decode(&createFoodRequest)
	} else {
		createFoodRequest.Name = form.Get("name")
	}

	var newFood model.Food
	newFood.Name = strings.TrimSpace(createFoodRequest.Name)

	food, err := ic.foodRepository.Create(r.Context(), newFood)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		json.NewEncoder(w).Encode(food)
	} else {
		http.Redirect(w, r, fmt.Sprint("/food/", food.Id), http.StatusSeeOther)
	}
}

func (ic *FoodController) replaceFood(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	food, found, err := ic.foodRepository.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} else if !found {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	var replaceFoodRequest request.CreateFoodRequest

	err = r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	form := r.Form

	if len(form) == 0 {
		json.NewDecoder(r.Body).Decode(&replaceFoodRequest)
	} else {
		replaceFoodRequest.Name = form.Get("name")
	}

	food.Name = strings.TrimSpace(replaceFoodRequest.Name)

	updatedFood, err := ic.foodRepository.Update(r.Context(), *food)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		json.NewEncoder(w).Encode(updatedFood)
	} else {
		templ.Handler(response.GetFood(updatedFood)).ServeHTTP(w, r)
	}
}

func (ic *FoodController) updateFood(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

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

	if r.Header.Get("Accept") == "application/json" {
		json.NewEncoder(w).Encode(updatedFood)
	}
}

func (ic *FoodController) deleteFood(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

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

	if r.Header.Get("Accept") == "application/json" {
		deleteFoodResponse := response.DeleteFoodResponse{Id: deletedId}
		json.NewEncoder(w).Encode(deleteFoodResponse)
	}
}
