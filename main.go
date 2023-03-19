package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThomasMatlak/food/controller"
	"github.com/ThomasMatlak/food/repository"
	"github.com/gorilla/mux"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	dbUri := "bolt://localhost:7687"                                 // TODO get from configuration
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.NoAuth()) // TODO implement auth
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	defer driver.Close(ctx)

	recipeRepository := repository.NewRecipeRepository(ctx, driver)
	recipeController := controller.NewRecipeController(recipeRepository)

	ingredientRepository := repository.NewIngredientRepository(ctx, driver)
	ingredientController := controller.NewIngredientController(ingredientRepository)

	router := mux.NewRouter().StrictSlash(true)

	recipeController.RecipeRoutes(router)
	ingredientController.IngredientRoutes(router)

	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println(pathTemplate)
		}

		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println(methods)
		}

		return nil
	})

	http.ListenAndServe(":8080", router)
}
