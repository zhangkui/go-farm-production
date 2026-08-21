package domain

// Pagination holds list-query parameters shared across all collection APIs.
type Pagination struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	OrderBy  string `json:"order_by" form:"order_by"`
	Order    string `json:"order" form:"order"` // asc | desc
	Search   string `json:"search" form:"search"`
}

// Defaults applied by transport layer before validation.
const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Normalize coerces invalid pagination values into safe defaults.
func (p *Pagination) Normalize() {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.PageSize < 1 {
		p.PageSize = DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}
	if p.Order == "" {
		p.Order = "desc"
	}
	if p.Order != "asc" && p.Order != "desc" {
		p.Order = "desc"
	}
}

// Offset returns the SQL OFFSET value for the current page.
func (p *Pagination) Offset() int { return (p.Page - 1) * p.PageSize }

// PageResult wraps a page of items with total count and pagination metadata.
type PageResult[T any] struct {
	Items     []T   `json:"items"`
	Total     int64 `json:"total"`
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	TotalPage int   `json:"total_page"`
}

// NewPageResult assembles a page result, computing total pages.
func NewPageResult[T any](items []T, total int64, p Pagination) PageResult[T] {
	tp := 0
	if p.PageSize > 0 {
		tp = int((total + int64(p.PageSize) - 1) / int64(p.PageSize))
	}
	if items == nil {
		items = []T{}
	}
	return PageResult[T]{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPage: tp}
}
