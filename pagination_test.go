package main

import "testing"

// validation guarantees:
// 1. page and size are either nil or integers > 0
// 2. if size is nil, page is also nil
func TestCalculatePagination(t *testing.T) {
	var tests = []struct {
		name       string
		page       *int
		size       *int
		limit      int
		offset     int
		actualPage int
	}{
		{"default values", nil, nil, 10, 0, 1},
		{"simple input values", intPtr(1), intPtr(10), 10, 0, 1},
		{"second page", intPtr(2), intPtr(10), 10, 10, 2},
		{"long second page", intPtr(2), intPtr(100), 100, 100, 2},
		{"size but no page", nil, intPtr(100), 100, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit, offset, actualPage := CalculatePagination(tt.page, tt.size)
			if limit != tt.limit {
				t.Errorf("limit incorrect. got %d, want %d", limit, tt.limit)
			}
			if offset != tt.offset {
				t.Errorf("offset incorrect. got %d, want %d", offset, tt.offset)
			}
			if actualPage != tt.actualPage {
				t.Errorf("actualPage incorrect. got %d, want %d", actualPage, tt.actualPage)
			}
		})
	}
}

// guarantees:
// 1. total >= 0
// 2. limit > 0
func TestTotalPages(t *testing.T) {
	var tests = []struct {
		name       string
		total      int64
		limit      int
		totalPages int64
	}{
		{"even values", int64(100), 10, 10},
		{"zero total", int64(0), 10, 0},
		{"zero limit", int64(100), 0, 0},
		{"odd limit", int64(100), 11, 10},
		{"odd total", int64(101), 10, 11},
		{"odd values", int64(101), 11, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pages := CalculateTotalPages(tt.total, tt.limit)
			if pages != tt.totalPages {
				t.Errorf("pages incorrect. got %d, want %d", pages, tt.totalPages)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
