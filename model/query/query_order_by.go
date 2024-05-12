package query

type QueryOrderBy struct {
	FieldName string `json:"field_name"`
	IsAsc     bool   `json:"is_asc"`
}

func NewDDbQueryOrderBy(field string, asc bool) *QueryOrderBy {
	return &QueryOrderBy{
		FieldName: field,
		IsAsc:     asc,
	}
}
