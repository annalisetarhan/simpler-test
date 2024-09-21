package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	db := InitDatabase()
	service := NewProductService(db)
	validator := validator.New()
	handler := NewProductHandler(service, validator)
	router := InitRouter(handler)

	zap.L().Info("Server is running on port 8080")
	http.ListenAndServe(":8080", router)

}

func InitDatabase() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		zap.S().Fatalf("Failed to load .env file: %v", err)
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		dbUser, dbPassword, dbName, dbHost, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		zap.S().Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Product{}); err != nil {
		zap.S().Fatalf("Failed to migrate database schema: %v", err)
	}

	zap.L().Info("Database connection initialized successfully")
	return db
}
