package main

import (
	"fmt"
	"net/http"

	"github.com/ThomasMatlak/food/controller"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)

	controller.RecipeRoutes(router)

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
