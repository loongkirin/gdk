package request

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
