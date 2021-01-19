package router

import (
	"github.com/gorilla/mux"
	"net/http"
	"scheduler-microservice/app/controllers"
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
		Route {
			Name: "Schedule status websocket handler",
			Path: "/schedule-status",
			Method: http.MethodPost,
			Handler: controllers.ScheduleStatus,
		},
	}

	for _, route := range routes {
		var handler http.Handler

		handler = route.Handler

		router.Name(route.Name).
			Methods(route.Method).
			Path(route.Path).
			Handler(handler)
	}

	return router
}
