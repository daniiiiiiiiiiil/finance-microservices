package pagination

type Pagination struct {
	Total       int  `json:"total"`
	Limit       int  `json:"limit"`
	Offset      int  `json:"offset"`
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
}

func NewPagination(total, limit, offset int) Pagination {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	currentPage := (offset / limit) + 1
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return Pagination{
		Total:       total,
		Limit:       limit,
		Offset:      offset,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		HasNext:     currentPage < totalPages,
		HasPrev:     currentPage > 1,
	}
}

func FromRequest(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func LimitOffset(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
