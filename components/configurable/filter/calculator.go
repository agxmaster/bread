package filter

import (
	"fmt"

	"github.com/thoas/go-funk"
)

func calculate(operator Operator, value string, conditions []string) bool {
	switch operator {
	case In:
		if len(conditions) == 0 {
			return true
		}

		return funk.ContainsString(conditions, value)
	case NotIn:
		if len(conditions) == 0 {
			return false
		}
		return !funk.ContainsString(conditions, value)
	default:
		panic(fmt.Sprintf("invalid calculator operator: %s", operator))
	}
}
