package filter

import (
	"strings"
)

type StandardFilter struct {
	fType
	Op         Operator
	Conditions []string
	Strict  bool
}

func (s *StandardFilter) Filter(_ UserInfo, value []string) bool {
	if s.Strict {
		if len(s.Conditions) != 0 && len(value) == 0 {
			if s.Op == In {
				return false
			}
			if s.Op == NotIn {
				return true
			}
		}
	}
	if len(value) == 0 {
		if len(s.Conditions) == 0 {
			if s.Op == In {
				return true
			}
			if s.Op == NotIn {
				return false
			}
		}

		return true
	}
	for _, v := range value {
		if calculate(s.Op, v, s.Conditions) {
			return true
		}
	}
	return false
}

func (s *StandardFilter) Hash() string {
	var buf strings.Builder
	// 减少扩容次数,降低gc压力
	buf.Grow(64)
	buf.WriteString(strings.Join(s.Conditions, ","))
	buf.WriteString(string(s.Op))
	return buf.String()
}

func NewStandardFilter(operator Operator, conditions []string, strict bool) *StandardFilter {
	return &StandardFilter{
		Standard,
		operator,
		conditions,
		strict,
	}
}
