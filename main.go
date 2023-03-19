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
	dbUri := "bolt://localhost:7687"                                 // todo get from configuration
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.NoAuth()) // todo implement auth
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	defer driver.Close(ctx)

	recipeRepository := repository.NewRecipeRepository(ctx, driver)
	recipeController := controller.NewRecipeController(recipeRepository)

	router := mux.NewRouter().StrictSlash(true)

	recipeController.RecipeRoutes(router)

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
