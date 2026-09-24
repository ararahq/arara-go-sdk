package arara

import (
	"net/url"
	"strconv"
)

// Pagination is the pagination block of a {data, pagination} response.
type Pagination struct {
	Page          int   `json:"page"`
	Size          int   `json:"size"`
	TotalElements int64 `json:"totalElements"`
	TotalPages    int   `json:"totalPages"`
}

// Paginated is a page of results in the {data, pagination} shape.
type Paginated[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// PageParams selects a page. Size <= 0 uses the endpoint default.
type PageParams struct {
	Page int
	Size int
}

func (p PageParams) values(defaultPageSize int) url.Values {
	return url.Values{
		"page": {strconv.Itoa(p.Page)},
		"size": {strconv.Itoa(defaultSize(p.Size, defaultPageSize))},
	}
}
