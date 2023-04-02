package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThomasMatlak/food/controller"
	"github.com/ThomasMatlak/food/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.StripSlashes)

	recipeController.RecipeRoutes(router)
	foodController.FoodRoutes(router)

	err = chi.Walk(router, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		fmt.Printf("%s %s\n", method, route)
		return nil
	})
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", router)
}
