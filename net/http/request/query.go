package request

type Query struct {
	QueryWheres []*QueryWhere   `json:"query_wheres"`
	OrderBy     []*QueryOrderBy `json:"order_by"`
	PageSize    int             `json:"page_size"`
	PageNumber  int             `json:"page_number"`
}

func NewQuery(wheres []*QueryWhere, ps int, pn int, order []*QueryOrderBy) *Query {
	return &Query{
		QueryWheres: wheres,
		PageSize:    ps,
		PageNumber:  pn,
		OrderBy:     order,
	}
}
