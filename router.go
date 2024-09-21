package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func InitRouter(handler *ProductHandler) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/health", HealthCheckHandler).Methods(http.MethodGet)

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	apiRouter.HandleFunc("/products", handler.CreateProduct).Methods(http.MethodPost)
	apiRouter.HandleFunc("/products", handler.GetProducts).Methods(http.MethodGet)
	apiRouter.HandleFunc("/products/{id:[0-9]+}", handler.GetProduct).Methods(http.MethodGet)
	apiRouter.HandleFunc("/products/{id:[0-9]+}", handler.UpdateProduct).Methods(http.MethodPatch)
	apiRouter.HandleFunc("/products/{id:[0-9]+}", handler.DeleteProduct).Methods(http.MethodDelete)

	zap.L().Info("Router initialized successfully")
	return router
}

func HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
