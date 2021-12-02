package filter

import (
	"fmt"
	"strings"

	"github.com/go-bread/components/configurable/types"
)

type fType string

type Relation string
type Operator string

const (
	Standard    fType = "standard"
	WhiteList   fType = "white_list"
	AccountType fType = "account_type"
	Category    fType = "category"

	And Relation = "And"
	Or  Relation = "Or"

	NotIn    Operator = "not_in"
	In       Operator = "in"
	Equal    Operator = "equal"
	NotEqual Operator = "not_equal"
)

func (r *fType) GetType() string {
	return string(*r)
}

type Filters struct {
	Fs       []IFilter
	Relation Relation
}

type IFilter interface {
	GetType() string
	Hash() string
	Filter(UserInfo, []string) bool
}

func TransferOperator(op string) Operator {
	switch strings.ToLower(op) {
	case string(NotIn):
		return NotIn
	case string(In):
		return In
	case string(Equal):
		return Equal
	case string(NotEqual):
		return NotEqual
	default:
		panic(fmt.Sprintf("invliad operator: %s", op))
	}
}

func (f *Filters) Filter(info UserInfo, values []string, action *types.Action) bool {
	if f.Relation == And {
		for _, ff := range f.Fs {
			// 更新场景下回显时, 如果标准过滤器通过的话, 就不进行其他filter()的校验
			if ff.GetType() == string(Standard) {
				if !ff.Filter(info, values) {
					return false
				}
				if len(values) > 0 && action != nil && *action == types.ActionUpdate {
					return true
				}
			} else if !ff.Filter(info, values) {
				return false
			}
		}
		return true
	}

	result := false
	for _, ff := range f.Fs {
		// 标准过滤器未通过, 直接终止
		if ff.GetType() == string(Standard) {
			if !ff.Filter(info, values) {
				return false
			}
			// 更新场景下回显时, 如果标准过滤器通过的话, 就不进行其他filter()的校验
			if len(values) > 0 && action != nil && *action == types.ActionUpdate {
				return true
			}
		} else if ff.Filter(info, values) {
			result = true
		}
	}
	return result
}

type UserInfo struct {
	UserId      uint64
	Category    string
	AccountType string
	Permissions map[int]struct{}
}
