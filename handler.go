package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type ProductHandler struct {
	productService *ProductService
	validator      *validator.Validate
}

func NewProductHandler(service *ProductService, validator *validator.Validate) *ProductHandler {
	return &ProductHandler{productService: service, validator: validator}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	zap.L().Info("Create product")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		zap.L().Error("Failed to create product because request body could not be read", zap.Error(err))
		http.Error(w, "unexpected error occurred", http.StatusInternalServerError)
		return
	}

	var request ProductCreateRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		zap.L().Info("Failed to create product because request could not be unmarshalled", zap.Error(err))
		http.Error(w, "failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	err = h.validator.Struct(request)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			zap.L().Info("Failed to create product because request failed validation", zap.Any("validationErrors", validationErrors))
			httpBadRequest(w, formatValidationErrors(validationErrors))
		} else {
			zap.L().Error("Unexpected error occurred during ProductCreateRequest validation", zap.Error(err))
			http.Error(w, "unexpected error occurred", http.StatusInternalServerError)
		}
		return
	}

	product, err := h.productService.CreateProduct(request)
	if err != nil {
		if errors.Is(err, ErrDuplicateSKU) {
			zap.L().Info("Failed to create product", zap.Error(err))
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		zap.L().Error("Failed to create product", zap.Error(err))
		http.Error(w, "unexpected error occurred", http.StatusInternalServerError)
		return
	}

	zap.L().Info("Product created successfully", zap.Uint("product ID", product.ID))
	httpCreated(w, &product)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	zap.L().Info("Get product", zap.String("path", r.URL.Path))

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		zap.L().Info("Failed to get product because product ID was invalid", zap.String("path", r.URL.Path))
		http.Error(w, "invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.productService.GetProduct(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			zap.L().Info("Failed to retrieve product because product was not found", zap.Int("product ID", id))
			http.Error(w, "product not found", http.StatusNotFound)
		} else {
			zap.L().Error("Failed to retrieve product", zap.Int("product ID", id), zap.Error(err))
			http.Error(w, "failed to retrieve product", http.StatusInternalServerError)
		}
		return
	}

	zap.L().Info("Product retrieved successfully", zap.Int("product ID", id))
	httpOK(w, product)
}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	zap.L().Info("Get products", zap.String("path", r.URL.Path))

	var page, size *int

	pageStr := r.URL.Query().Get("page")

	if pageStr != "" {
		pageInt, err := strconv.Atoi(pageStr)
		if err != nil || pageInt <= 0 {
			zap.L().Info("Failed to get products because page param was invalid", zap.String("path", r.URL.Path))
			http.Error(w, "invalid page param", http.StatusBadRequest)
			return
		}
		page = &pageInt
	}

	sizeStr := r.URL.Query().Get("size")
	if sizeStr != "" {
		sizeInt, err := strconv.Atoi(sizeStr)
		if err != nil || sizeInt <= 0 {
			zap.L().Info("Failed to get products because size param was invalid", zap.String("path", r.URL.Path))
			http.Error(w, "invalid size param", http.StatusBadRequest)
			return
		}
		size = &sizeInt
	}

	if page != nil && size == nil {
		zap.L().Info("Failed to get products because page was specified but size wasn't", zap.String("path", r.URL.Path))
		http.Error(w, "must specify size if page is included", http.StatusBadRequest)
		return
	}

	response, err := h.productService.GetProducts(page, size)
	if err != nil {
		if errors.Is(err, ErrOutOfRange) {
			zap.L().Info("Failed to get products", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		zap.L().Error("Failed to get products", zap.Error(err))
		http.Error(w, "unexpected error occurred", http.StatusInternalServerError)
		return
	}

	httpOK(w, response)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	zap.L().Info("Update product", zap.String("path", r.URL.Path))

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		zap.L().Info("Failed to update product because product ID was invalid", zap.String("path", r.URL.Path))
		http.Error(w, "invalid product ID", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		zap.L().Error("Failed to update product because request body could not be read", zap.Error(err))
		http.Error(w, "unexpected error occurred", http.StatusInternalServerError)
		return
	}

	var request ProductUpdateRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		zap.L().Info("Failed to update product because request could not be unmarshalled", zap.Error(err))
		http.Error(w, "failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	err = h.validator.Struct(request)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			zap.L().Info("Failed to update product because request failed validation", zap.Any("validationErrors", validationErrors))
			httpBadRequest(w, formatValidationErrors(validationErrors))
		} else {
			zap.L().Error("Unexpected error occurred during ProductUpdateRequest validation", zap.Error(err))
			http.Error(w, "unexpected error occurred", http.StatusInternalServerError)
		}
		return
	}

	product, err := h.productService.UpdateProduct(id, request)
	if err != nil {
		if errors.Is(err, ErrDuplicateSKU) {
			zap.L().Info("Failed to update product", zap.Error(err))
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		zap.L().Error("Failed to update product", zap.Error(err))
		http.Error(w, "unexpected error occurred", http.StatusInternalServerError)
		return
	}

	zap.L().Info("Product updated successfully", zap.Uint("product ID", product.ID))
	httpOK(w, &product)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	zap.L().Info("Delete product", zap.String("path", r.URL.Path))

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		zap.L().Info("Failed to delete product because product ID was invalid", zap.String("path", r.URL.Path))
		http.Error(w, "invalid product ID", http.StatusBadRequest)
		return
	}

	err = h.productService.DeleteProduct(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			zap.L().Info("Failed to delete product because product was not found", zap.Int("product ID", id))
			http.Error(w, "invalid product ID", http.StatusNotFound)
			return
		}
		zap.L().Error("Failed to delete product", zap.Error(err))
		http.Error(w, "unexpected error occurred", http.StatusInternalServerError)
		return
	}

	zap.L().Info("Product deleted successfully", zap.Int("product ID", id))
	w.WriteHeader(http.StatusNoContent)
}

func httpOK(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(body)
}

func httpCreated(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func httpBadRequest(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(body)
}

func formatValidationErrors(validationErrors validator.ValidationErrors) map[string]string {
	errorsMap := make(map[string]string)
	for _, ve := range validationErrors {
		errorsMap[ve.Field()] = fmt.Sprintf("failed validation on the '%s' rule", ve.Tag())
	}
	return errorsMap
}
