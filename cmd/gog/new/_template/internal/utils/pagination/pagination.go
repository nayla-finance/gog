package pagination

import (
	"math"
)

type PaginatedResponse[T any] struct {
	Status     string     `json:"status"`
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Page         int  `json:"page"`
	PageSize     int  `json:"page_size"`
	HasNext      bool `json:"has_next"`
	HasPrevious  bool `json:"has_previous"`
	TotalRecords int  `json:"total_records"`
	TotalPages   int  `json:"total_pages"`
}

func ToPaginatedResponse[T any](data []T, totalRecords int, page int, pageSize int) PaginatedResponse[T] {
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))

	pagination := Pagination{
		Page:         page,
		PageSize:     pageSize,
		HasNext:      page < totalPages,
		HasPrevious:  page > 1,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
	}

	return PaginatedResponse[T]{
		Status:     "success",
		Data:       data,
		Pagination: pagination,
	}
}

func GetPaginationParams(p, l int) (page int, limit int) {
	limit = l
	if limit <= 0 {
		limit = 25
	}

	page = p
	if page <= 0 {
		page = 1
	}

	return page, limit
}
