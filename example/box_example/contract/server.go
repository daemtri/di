package contract

import (
	"context"
	"net/http"
)

type Server interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

type Service interface {
	AddRoute(mux *http.ServeMux)
}
