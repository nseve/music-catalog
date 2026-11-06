package server

import (
	"fmt"
	"net/http"
)

type ApiVersion string

const (
	ApiVersion1 = ApiVersion("v1")
)

type APIVersionRouter struct {
	*http.ServeMux
	apiVesion ApiVersion
}

func NewAPIVersionRouter(apiVersion ApiVersion) *APIVersionRouter {
	return &APIVersionRouter{
		ServeMux:  http.NewServeMux(),
		apiVesion: apiVersion,
	}
}

func (r *APIVersionRouter) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		pattern := fmt.Sprintf("%s %s", route.Method, route.Path)
		r.Handle(pattern, route.Handler)
	}
}
