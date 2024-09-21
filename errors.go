package main

import "errors"

var (
	ErrNotFound     = errors.New("product not found")
	ErrDuplicateSKU = errors.New("product with this SKU already exists")
	ErrOutOfRange   = errors.New("page number out of range")
)
