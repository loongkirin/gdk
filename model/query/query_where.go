package query

type QueryWhere struct {
	QueryFilters []QueryFilter `json:"query_filters"`
	Connector    string        `json:"connector"`
}

func NewDbQueryWhere(filters []QueryFilter, connector string) *QueryWhere {
	return &QueryWhere{
		QueryFilters: filters,
		Connector:    connector,
	}
}
