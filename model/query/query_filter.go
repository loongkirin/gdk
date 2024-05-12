package query

type QueryFilter struct {
	FieldName       string        `json:"field_name"`
	FilterValues    []interface{} `json:"filter_values"`
	FilterOperation string        `json:"filter_operation"`
	Connector       string        `json:"connector"`
	FieldType       string        `json:"field_type"`
}

func NewQueryFilter(field string, values []interface{}, op string, ft string) *QueryFilter {
	return &QueryFilter{
		FieldName:       field,
		FilterValues:    values,
		FilterOperation: op,
		Connector:       "AND",
		FieldType:       ft,
	}
}
