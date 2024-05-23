package query

import (
	"fmt"
	"strings"
)

type Operator string

const (
	EQ      Operator = "EQ"
	NEQ     Operator = "NEQ"
	LT      Operator = "LT"
	LTE     Operator = "LTE"
	GT      Operator = "GT"
	GTE     Operator = "GTE"
	LIKE    Operator = "LIKE"
	IN      Operator = "IN"
	BETWEEN Operator = "BETWEEN"
)

type Connector string

const (
	AND Connector = "AND"
	OR  Connector = "OR"
)

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

func (q *Query) GetWhereClause() (whereClause string, values []interface{}, order string) {
	whereClause = " 1=1 "
	order = "id"
	if len(q.QueryWheres) < 1 {
		return whereClause, values, order
	}
	whereClause += "AND "
	var sb strings.Builder

	for _, where := range q.QueryWheres {
		var subClause strings.Builder
		for _, filter := range where.QueryFilters {
			fieldName := filter.FieldName
			var op string
			switch filter.Operator {
			case EQ:
				op = " = ? "
				values = append(values, filter.FilterValues[0])
			case NEQ:
				op = " <> ? "
				values = append(values, filter.FilterValues[0])
			case LT:
				op = " < ? "
				values = append(values, filter.FilterValues[0])
			case LTE:
				op = " <= ? "
				values = append(values, filter.FilterValues[0])
			case GT:
				op = " > ? "
				values = append(values, filter.FilterValues[0])
			case GTE:
				op = " >= ? "
				values = append(values, filter.FilterValues[0])
			case LIKE:
				op = " LIKE ? "
				values = append(values, "%"+fmt.Sprint(filter.FilterValues[0])+"%")
			case IN:
				op = " IN ? "
				values = append(values, filter.FilterValues)
			case BETWEEN:
				op = " BETWEEN ? AND ? "
				values = append(values, filter.FilterValues[0], filter.FilterValues[1])
			}
			subClause.WriteString(fmt.Sprintf(" %s %s AND ", fieldName, op))
		}
		subWhere := subClause.String()
		subWhere = strings.Trim(subWhere, "AND ")
		sb.WriteString(subWhere)
		sb.WriteString(fmt.Sprintf(" %s ", where.Connector))
	}
	whereClause = whereClause + sb.String()
	whereClause = strings.Trim(whereClause, "AND ")
	whereClause = strings.Trim(whereClause, "OR ")

	var sbOrder strings.Builder
	for _, order := range q.OrderBy {
		if order.IsAsc {
			sbOrder.WriteString(fmt.Sprintf("%s ASC,", order.FieldName))
		} else {
			sbOrder.WriteString(fmt.Sprintf("%s DESC,", order.FieldName))
		}
	}
	order = sbOrder.String()
	order = strings.Trim(order, ",")
	if len(order) < 1 {
		order = "id"
	}
	return whereClause, values, order
}
