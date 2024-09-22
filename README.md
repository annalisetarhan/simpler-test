# Simpler Test Microservice

This is a Go-based microservice backed by a PostgreSQL database, providing CRUD operations for products. The project uses Docker for containerization.

## .env File Setup

You need to create a `.env` file in the root of the project to configure the database credentials:
```
DB_USER={user_name}
DB_PASSWORD={password}
DB_NAME={database_name}
```

## Running the Application

1. Build and run the application using Docker Compose:
```
docker compose up --build
```


2. Access the service at `http://localhost:8080`.

## Running Tests
Included are a complete set of end-to-end tests, as well as unit tests for pagination functions. To run the tests within the Docker container, use the following command. Note: this will drop all data in the products table!
```
docker compose --profile test run test
```


## Stopping the Application

Stop and remove the containers:

```
docker compose down
```

To remove the database data as well:

```
docker compose down -v
```


## Health Check

The service includes a health check, available at:

```
http://localhost:8080/health
```

## API Documentation

Detailed API documentation can be found in the `api.yaml` file, formatted according to the OpenAPI 3.0 specification.. It includes information on the available endpoints, request parameters, and response structures.

## Usage

Here are some example `cURL` commands to interact with the API:

### Create a Product (POST /api/v1/products)

```
curl -X POST http://localhost:8080/api/v1/products \
-H "Content-Type: application/json" \
-d '{
  "name": "New Product",
  "description": "A new test product",
  "sku": "sku12345",
  "price": 99.99,
  "quantity": 10,
  "category": "Test Category"
}'
```

### Get all products with default pagination (GET /api/v1/products)
This returns the first ten products in the database, ordered by `ID`
```
curl -X GET http://localhost:8080/api/v1/products
```

### Get all products with custom pagination (GET /api/v1/products?page={page}&size={size})
Replace `{page}` and `{size}` with appropriate values. `page` is the page number, starting from 1, and `size` is the number of results per page. Requests where `page` is specified but `size` is not will be rejected. If the requested page is out of range, a `422 Unprocessable Entity` error will be returned.
```
curl -X GET "http://localhost:8080/api/v1/products?page=1&size=10"
```

### Get a specific product by ID (GET /api/v1/products/{id})
```
curl -X GET http://localhost:8080/api/v1/products/1
```

### Update a product (PATCH /api/v1/products/{id})
```
curl -X PATCH http://localhost:8080/api/v1/products/1 \
-H "Content-Type: application/json" \
-d '{
  "name": "Updated Product",
  "description": "Updated description",
  "price": 89.99,
  "quantity": 20
}'
```

### Delete a product (DELETE /api/v1/products/{id})
```
curl -X DELETE http://localhost:8080/api/v1/products/1
```