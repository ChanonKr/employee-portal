package shared

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PageRequest struct {
	Page     int // 1-based
	PageSize int
}

// NewPageRequest applies defaults for zero values and rejects out-of-range input.
func NewPageRequest(page, pageSize int) (PageRequest, error) {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}
	if page < 1 {
		return PageRequest{}, NewValidationError("page", "must be 1 or more")
	}
	if pageSize < 1 || pageSize > MaxPageSize {
		return PageRequest{}, NewValidationError("pageSize", "must be from 1 to 100")
	}
	return PageRequest{Page: page, PageSize: pageSize}, nil
}

func (p PageRequest) Offset() int { return (p.Page - 1) * p.PageSize }
