package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func InitRouter(handler *ProductHandler) *mux.Router {
	router := mux.NewRouter().PathPrefix("/api/v1").Subrouter()

	router.HandleFunc("/products", handler.CreateProduct).Methods(http.MethodPost)
	router.HandleFunc("/products", handler.GetProducts).Methods(http.MethodGet)
	router.HandleFunc("/products/{id:[0-9]+}", handler.GetProduct).Methods(http.MethodGet)
	router.HandleFunc("/products/{id:[0-9]+}", handler.UpdateProduct).Methods(http.MethodPatch)
	router.HandleFunc("/products/{id:[0-9]+}", handler.DeleteProduct).Methods(http.MethodDelete)

	zap.L().Info("Router initialized successfully")
	return router
}
