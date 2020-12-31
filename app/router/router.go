package router

import (
	"github.com/gorilla/mux"
	"net/http"
	"scheduler-microservice/app/controllers"
	"scheduler-microservice/pkg/logs"
)

//Route Create a single route object
type Route struct {
	Name    string
	Path    string
	Method  string
	Handler http.HandlerFunc
}

//Routes Create an object of different routes
type Routes []Route

// InitRoutes Set up routes
func InitRoutes() *mux.Router {

	router := mux.NewRouter()

	routes := Routes{
		// health check
		Route{
			Name:    "Health Check",
			Path:    "/",
			Method:  http.MethodGet,
			Handler: controllers.IndexHandler,
		},
		Route{
			Name:    "Get Schedule",
			Path:    "/schedule",
			Method:  http.MethodPost,
			Handler: controllers.GetSchedule,
		},
	}

	for _, route := range routes {
		var handler http.Handler

		handler = route.Handler
		handler = logs.HandlerLog(handler, route.Name)

		router.Name(route.Name).
			Methods(route.Method).
			Path(route.Path).
			Handler(handler)
	}

	return router
}
