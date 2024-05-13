package api

import (
	"github.com/loongkirin/gdk/model/query"
)

type GetDataListRequest struct {
	SearchQuery query.Query      `json:"search_query"`
	Pagination  query.Pagination `json:"page_info"`
}

type DataListResponse[T any] struct {
	DataList   []T              `json:"data_list"`
	Pagination query.Pagination `json:"page_info"`
}

type DataRequest[T any] struct {
	Data T `json:"data"`
}

type DataResponse[T any] struct {
	Data T `json:"data"`
}
