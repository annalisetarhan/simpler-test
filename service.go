package main

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

const (
	DefaultPage     = 1
	DefaultPageSize = 10
)

func (s *ProductService) CreateProduct(req ProductCreateRequest) (*Product, error) {
	product := Product{
		Name:        req.Name,
		Description: req.Description,
		SKU:         req.SKU,
		Price:       req.Price,
		Quantity:    req.Quantity,
		Category:    req.Category,
	}

	err := s.db.Create(&product).Error
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateSKU, req.SKU)
		}
		return nil, err
	}

	return &product, nil
}

func (s *ProductService) GetProduct(id int) (*Product, error) {
	var product Product

	err := s.db.Where("id = ?", id).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &product, nil
}

func (s *ProductService) GetProducts(requestedPage, requestedSize *int) (*BulkProductResponse, error) {
	var products []Product
	var total int64

	limit, offset, page := CalculatePagination(requestedPage, requestedSize)

	err := s.db.Model(&Product{}).Count(&total).Error
	if err != nil {
		return nil, err
	}

	err = s.db.Order("id ASC").Offset(offset).Limit(limit).Find(&products).Error
	if err != nil {
		return nil, err
	}

	if total > 0 && len(products) == 0 {
		return nil, ErrOutOfRange
	}

	totalPages := CalculateTotalPages(total, limit)

	response := BulkProductResponse{
		Products:   products,
		Page:       page,
		Size:       limit,
		TotalPages: totalPages,
		TotalCount: total,
	}

	return &response, nil
}

func (s *ProductService) UpdateProduct(id int, req ProductUpdateRequest) (*Product, error) {
	product, err := s.GetProduct(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.SKU != nil {
		product.SKU = *req.SKU
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Quantity != nil {
		product.Quantity = *req.Quantity
	}
	if req.Category != nil {
		product.Category = *req.Category
	}

	err = s.db.Where("id = ?", id).Save(product).Error
	if err != nil {
		if isUniqueConstraintError(err) && req.SKU != nil {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateSKU, *req.SKU)
		}
		return nil, err
	}

	return product, nil
}

func (s *ProductService) DeleteProduct(id int) error {
	result := s.db.Delete(&Product{}, id)

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return result.Error
}

func CalculatePagination(page, size *int) (limit, offset, actualPage int) {
	actualPage = DefaultPage
	limit = DefaultPageSize

	if page != nil {
		actualPage = *page
	}
	if size != nil {
		limit = *size
	}

	offset = (actualPage - 1) * limit
	return limit, offset, actualPage
}

func CalculateTotalPages(total int64, limit int) int64 {
	if limit == 0 {
		return 0
	}
	return (total + int64(limit) - 1) / int64(limit)
}

func isUniqueConstraintError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
