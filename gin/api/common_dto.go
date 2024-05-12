package api

import (
	"github.com/loongkirin/gdk/model/query"
)

type GetDataListRequest struct {
	SearchQuery query.Query `json:"search_query"`
	Pagination  Pagination  `json:"page_info"`
}

type DataListResponse[T any] struct {
	DataList   []T        `json:"data_list"`
	Pagination Pagination `json:"page_info"`
}

type DataRequest[T any] struct {
	Data T `json:"data"`
}

type DataResponse[T any] struct {
	Data T `json:"data"`
}

type Pagination struct {
	PageSize    int  `json:"page_size"`
	PageNumber  int  `json:"page_number"`
	HasNextPage bool `json:"has_next_page"`
}
