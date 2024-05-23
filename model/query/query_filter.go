package query

type QueryFilter struct {
	FieldName    string        `json:"field_name"`
	FilterValues []interface{} `json:"filter_values"`
	Operator     Operator      `json:"operator"`
	Connector    Connector     `json:"connector"`
	FieldType    string        `json:"field_type"`
}

func NewQueryFilter(field string, values []interface{}, op Operator, ft string) *QueryFilter {
	return &QueryFilter{
		FieldName:    field,
		FilterValues: values,
		Operator:     op,
		Connector:    AND,
		FieldType:    ft,
	}
}
