package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func TestEmptyDatabase(t *testing.T) {
	router, logger := initRouter()
	defer logger.Sync()

	server := httptest.NewServer(router)
	defer server.Close()

	e := httpexpect.Default(t, server.URL)

	response := e.GET("/api/v1/products").
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	response.ContainsKey("page").Value("page").Number().IsEqual(1)
	response.ContainsKey("size").Value("size").Number().IsEqual(10)
	response.ContainsKey("total_pages").Value("total_pages").Number().IsEqual(0)
	response.ContainsKey("total_count").Value("total_count").Number().IsEqual(0)
}

func TestCreateProducts(t *testing.T) {
	router, logger := initRouter()
	defer logger.Sync()

	server := httptest.NewServer(router)
	defer server.Close()

	e := httpexpect.Default(t, server.URL)

	testCases := []struct {
		name           string
		product        ProductCreateRequest
		expectedStatus int
	}{
		{
			name: "Valid product",
			product: ProductCreateRequest{
				Name:        "first product",
				Description: "this describes the first product",
				SKU:         "1234",
				Price:       99.99,
				Quantity:    1,
				Category:    "product > subtype",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid product - missing name",
			product: ProductCreateRequest{
				Description: "missing name",
				SKU:         "1235",
				Price:       50.00,
				Quantity:    2,
				Category:    "product > subtype",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid product - invalid price",
			product: ProductCreateRequest{
				Name:        "third product",
				Description: "invalid price",
				SKU:         "1236",
				Price:       -10.0,
				Quantity:    3,
				Category:    "product > subtype",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid product - duplicate sku",
			product: ProductCreateRequest{
				Name:        "fourth product",
				Description: "duplicate sku product",
				SKU:         "1234",
				Price:       10.0,
				Quantity:    3,
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Invalid product - missing sku",
			product: ProductCreateRequest{
				Name:        "fifth product",
				Description: "missing sku product",
				Price:       10.0,
				Quantity:    3,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid product - invalid quantity",
			product: ProductCreateRequest{
				Name:        "sixth product",
				Description: "bad quantity product",
				SKU:         "12348",
				Price:       10.0,
				Quantity:    -1,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Valid product - barely valid quantity",
			product: ProductCreateRequest{
				Name:        "seventh product",
				Description: "barely valid quantity product",
				SKU:         "12349",
				Price:       10.0,
				Quantity:    0,
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := e.POST("/api/v1/products").WithJSON(tc.product).
				Expect().
				Status(tc.expectedStatus)

			if tc.expectedStatus == http.StatusCreated {
				object := response.JSON().Object()
				object.Value("id").Number().Gt(0)
				object.Value("name").String().IsEqual(tc.product.Name)
				object.Value("description").String().IsEqual(tc.product.Description)
				object.Value("sku").String().IsEqual(tc.product.SKU)
				object.Value("price").Number().IsEqual(tc.product.Price)
				object.Value("quantity").Number().IsEqual(tc.product.Quantity)
				object.Value("category").String().IsEqual(tc.product.Category)
				object.Value("created_at").String().NotEmpty()
				object.Value("updated_at").String().NotEmpty()
				object.Value("deleted_at").IsNull()
			}
		})
	}
}

func TestGetProduct(t *testing.T) {
	router, logger := initRouter()
	defer logger.Sync()

	server := httptest.NewServer(router)
	defer server.Close()

	e := httpexpect.Default(t, server.URL)

	// insert sample products
	productData := getSampleProductRequests()

	// store created product objects
	productObjects := []*httpexpect.Object{}
	for _, product := range productData {
		response := e.POST("/api/v1/products").WithJSON(product).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()
		productObjects = append(productObjects, response)
	}

	// get each product, compare all fields
	for _, object := range productObjects {
		id := int(object.Value("id").Number().Raw())
		response := e.GET("/api/v1/products/" + strconv.Itoa(id)).Expect().Status(http.StatusOK)
		product := response.JSON().Object()

		product.Value("id").Number().IsEqual(id)
		product.Value("name").String().IsEqual(object.Value("name").String().Raw())
		product.Value("description").String().IsEqual(object.Value("description").String().Raw())
		product.Value("sku").String().IsEqual(object.Value("sku").String().Raw())
		product.Value("price").Number().IsEqual(object.Value("price").Number().Raw())
		product.Value("quantity").Number().IsEqual(object.Value("quantity").Number().Raw())
		product.Value("category").String().IsEqual(object.Value("category").String().Raw())
		product.Value("created_at").String().IsEqual(object.Value("created_at").String().Raw())
		product.Value("updated_at").String().IsEqual(object.Value("updated_at").String().Raw())
		product.Value("deleted_at").IsNull()
	}
}

func TestGetProducts(t *testing.T) {
	router, logger := initRouter()
	defer logger.Sync()

	server := httptest.NewServer(router)
	defer server.Close()

	e := httpexpect.Default(t, server.URL)

	// insert sample products
	products := getSampleProductRequests()
	for _, product := range products {
		e.POST("/api/v1/products").WithJSON(product).
			Expect().
			Status(http.StatusCreated)
	}

	numProducts := len(products)

	testCases := []struct {
		name               string
		queryParams        map[string]string
		expectedStatus     int
		expectedPage       int
		expectedSize       int
		expectedTotalPages int
	}{
		{
			name:               "Defaults",
			expectedStatus:     http.StatusOK,
			expectedPage:       1,
			expectedSize:       10,
			expectedTotalPages: 1,
		},
		{
			name:               "Default page, custom size",
			queryParams:        map[string]string{"size": "20"},
			expectedStatus:     http.StatusOK,
			expectedPage:       1,
			expectedSize:       20,
			expectedTotalPages: 1,
		},
		{
			name:               "Custom page, custom size",
			queryParams:        map[string]string{"page": "2", "size": "2"},
			expectedStatus:     http.StatusOK,
			expectedPage:       2,
			expectedSize:       2,
			expectedTotalPages: 2,
		},
		{
			name:           "Invalid - page provided without size",
			queryParams:    map[string]string{"page": "2"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid - page number out of range",
			queryParams:    map[string]string{"page": "2", "size": "10"},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "Invalid - negative page number",
			queryParams:    map[string]string{"page": "-2", "size": "2"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid - negative size",
			queryParams:    map[string]string{"page": "2", "size": "-2"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := e.GET("/api/v1/products")

			for key, val := range tc.queryParams {
				request = request.WithQuery(key, val)
			}

			response := request.Expect().Status(tc.expectedStatus)

			if tc.expectedStatus == http.StatusOK {
				body := response.JSON().Object()
				body.ContainsKey("products").Value("products").Array().Length().Gt(0)
				body.ContainsKey("page").Value("page").Number().IsEqual(tc.expectedPage)
				body.ContainsKey("size").Value("size").Number().IsEqual(tc.expectedSize)
				body.ContainsKey("total_pages").Value("total_pages").Number().IsEqual(tc.expectedTotalPages)
				body.ContainsKey("total_count").Value("total_count").Number().IsEqual(numProducts)
			}
		})
	}
}

func TestUpdateProducts(t *testing.T) {
	router, logger := initRouter()
	defer logger.Sync()

	server := httptest.NewServer(router)
	defer server.Close()

	e := httpexpect.Default(t, server.URL)

	// insert sample products
	for _, product := range getSampleProductRequests() {
		e.POST("/api/v1/products").WithJSON(product).
			Expect().
			Status(http.StatusCreated)
	}

	duplicatedSKU := getSampleProductRequests()[1].SKU

	testCases := []struct {
		name           string
		productID      string
		request        ProductUpdateRequest
		expectedStatus int
	}{
		{
			name:      "Valid updates",
			productID: "1",
			request: ProductUpdateRequest{
				Name:        strPtr("new name"),
				Description: strPtr("new description"),
				SKU:         strPtr("9876"),
				Price:       floatPtr(1.11),
				Quantity:    intPtr(1),
				Category:    strPtr("new > category"),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Invalid product ID",
			productID: "abc",
			request: ProductUpdateRequest{
				Name: strPtr("new name"),
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "Missing product ID",
			productID: "",
			request: ProductUpdateRequest{
				Name: strPtr("new name"),
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "Product doesn't exist",
			productID: "1000000",
			request: ProductUpdateRequest{
				Name: strPtr("new name"),
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "Duplicated SKU",
			productID: "1",
			request: ProductUpdateRequest{
				SKU: &duplicatedSKU,
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:      "Invalid quantity",
			productID: "1",
			request: ProductUpdateRequest{
				Quantity: intPtr(-1),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Invalid price",
			productID: "1",
			request: ProductUpdateRequest{
				Price: floatPtr(-1),
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := "/api/v1/products/" + tc.productID
			response := e.PATCH(path).WithJSON(tc.request).
				Expect().
				Status(tc.expectedStatus)
			if tc.expectedStatus == http.StatusOK {
				object := response.JSON().Object()
				object.Value("id").Number().IsEqual(1)
				if tc.request.Name != nil {
					object.Value("name").String().IsEqual(*tc.request.Name)
				}
				if tc.request.Description != nil {
					object.Value("description").String().IsEqual(*tc.request.Description)
				}
				if tc.request.SKU != nil {
					object.Value("sku").String().IsEqual(*tc.request.SKU)
				}
				if tc.request.Price != nil {
					object.Value("price").Number().IsEqual(*tc.request.Price)
				}
				if tc.request.Quantity != nil {
					object.Value("quantity").Number().IsEqual(*tc.request.Quantity)
				}
				if tc.request.Category != nil {
					object.Value("category").String().IsEqual(*tc.request.Category)
				}
			}
		})
	}
}

func TestDeleteProducts(t *testing.T) {
	router, logger := initRouter()
	defer logger.Sync()

	server := httptest.NewServer(router)
	defer server.Close()

	e := httpexpect.Default(t, server.URL)

	// insert sample products
	productData := getSampleProductRequests()

	// store created product IDs
	productIDs := []int{}
	for _, product := range productData {
		response := e.POST("/api/v1/products").WithJSON(product).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		productID := uint(response.Value("id").Number().Raw())
		productIDs = append(productIDs, int(productID))
	}

	for _, productID := range productIDs {

		// first delete request should succeed
		e.DELETE("/api/v1/products/" + strconv.Itoa(productID)).Expect().Status(http.StatusNoContent)

		// second delete request should fail
		e.DELETE("/api/v1/products/" + strconv.Itoa(productID)).Expect().Status(http.StatusNotFound)

		// get requests should fail after deletion
		e.GET("/api/v1/products/" + strconv.Itoa(productID)).Expect().Status(http.StatusNotFound)
	}

	// requests to create the same products again (with the same SKUs) should succeed
	for _, product := range productData {
		e.POST("/api/v1/products").WithJSON(product).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()
	}
}

func initRouter() (*mux.Router, *zap.Logger) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	zap.ReplaceGlobals(logger)

	db := InitDatabase()
	service := NewProductService(db)
	validator := validator.New()
	handler := NewProductHandler(service, validator)

	return InitRouter(handler), logger
}

func getSampleProductRequests() []ProductCreateRequest {
	return []ProductCreateRequest{
		{
			Name:        "first product",
			Description: "this describes the first product",
			SKU:         "1234",
			Price:       99.99,
			Quantity:    1,
			Category:    "product > subtype",
		},
		{
			Name:        "second product",
			Description: "this describes the second product",
			SKU:         "5678",
			Price:       9.99,
			Quantity:    10,
			Category:    "product > other_type",
		},
		{
			Name:        "third product",
			Description: "this describes the third product",
			SKU:         "91011",
			Price:       19.99,
			Quantity:    100,
			Category:    "product > subtype > another_type",
		},
	}
}

func strPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}
