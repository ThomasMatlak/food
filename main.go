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
	dbUri := "bolt://localhost:7687" // TODO get from configuration
	useConsoleLogger := func(level neo4j.LogLevel) func(config *neo4j.Config) {
		return func(config *neo4j.Config) {
			config.Log = neo4j.ConsoleLogger(level)
		}
	}
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.NoAuth(), useConsoleLogger(neo4j.DEBUG)) // TODO implement auth
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	defer driver.Close(ctx)

	recipeRepository := repository.NewRecipeRepository(driver)
	recipeController := controller.NewRecipeController(recipeRepository)

	foodRepository := repository.NewFoodRepository(driver)
	foodController := controller.NewFoodController(foodRepository)

	router := mux.NewRouter().StrictSlash(true)

	recipeController.RecipeRoutes(router)
	foodController.FoodRoutes(router)

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
