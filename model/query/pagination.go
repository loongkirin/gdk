package query

type Pagination struct {
	PageSize    int  `json:"page_size"`
	PageNumber  int  `json:"page_number"`
	HasNextPage bool `json:"has_next_page"`
}
