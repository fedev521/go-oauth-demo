package mw

import (
	"log"
	"net/http"
	"os"
	"strings"

	"example.com/go-oauth/api"
	"github.com/gorilla/mux"
)

const BasePath = "/api/v1"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		IndexHandler,
	},
	Route{
		"Products",
		"GET",
		BasePath + "/products",
		api.GetProducts,
	},
}

// NewRouter creates a router and adds middleware and routes.
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	// add middleware
	router.Use(
		clockMiddleware,
		loggingMiddleware(os.Stdout),
		prepareTokenValidationMiddleware(),
	)
	// add specific routes
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	logRoutes(router)

	return router
}

func logRoutes(r *mux.Router) {
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		methods, err1 := route.GetMethods()
		pathTemplate, err2 := route.GetPathTemplate()
		if err1 == nil && err2 == nil {
			log.Println(strings.Join(methods, " "), pathTemplate)
		}

		return nil
	})
}
