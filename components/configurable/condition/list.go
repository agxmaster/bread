package condition

import (
	"container/list"
)

func NewConditionList() *List {
	return &List{
		list.New(),
		make(map[string]bool),
		make(map[string]string),
	}
}

type ListPair struct {
	Front    *List
	End      *List
	IsOrient bool
}

type List struct {
	*list.List

	intersectFields map[string]bool
	intersectRefer  map[string]string
}

func (cl *List) AddNode(key string, value []string) {
	cl.PushBack(newConditionNode(cl, key, value))
}

func (cl *List) HasIntersectRefer() bool {
	return len(cl.intersectFields) > 0
}

func (cl *List) AddIntersectRefer(refers map[string]string) {
	for k, v := range refers {
		cl.intersectFields[v] = true
		cl.intersectRefer[k] = v
	}
}

func (cl *List) GetAllIntersectRefers() map[string]string {
	return cl.intersectRefer
}

type Node struct {
	key   string
	value []string
	list  *List
}

func (n *Node) Key() string {
	return n.key
}

func (n *Node) Value() []string {
	return n.value
}

func (n *Node) IsIntersectRefer() (string, bool) {
	field, ok := n.list.intersectRefer[n.key]
	return field, ok
}

func newConditionNode(list *List, key string, value []string) *Node {
	return &Node{
		key:   key,
		value: value,
		list:  list,
	}
}

func FillListWithCondition(list *List, conditions map[string][]string) *ListPair {
	if list == nil || list.Len() == 0 || len(conditions) == 0 {
		return nil
	}

	l := NewConditionList()
	refers := list.GetAllIntersectRefers()
	if refers != nil {
		l.AddIntersectRefer(refers)
	}
	cur := list.Front()
	for cur != nil {
		key := cur.Value.(*Node).Key()
		if v, ok := conditions[key]; ok {
			l.AddNode(key, v)
		} else {
			l.AddNode(key, cur.Value.(*Node).Value())
		}
		cur = cur.Next()
	}
	return &ListPair{
		Front: l,
		End:   nil,
	}
}

func ReFillListPairWithCondition(pair *ListPair, conditions map[string][]string) (*ListPair, bool) {
	if pair.Front == nil || pair.Front.Len() == 0 || len(conditions) == 0 {
		return nil, false
	}
	fill := false
	l := NewConditionList()
	refers := pair.Front.GetAllIntersectRefers()
	if refers != nil {
		l.AddIntersectRefer(refers)
	}
	cur := pair.Front.Front()
	for cur != nil {
		key := cur.Value.(*Node).Key()
		if v, ok := conditions[key]; ok && len(cur.Value.(*Node).Value()) == 0 {
			fill = true
			l.AddNode(key, v)
		} else {
			l.AddNode(key, cur.Value.(*Node).Value())
		}
		cur = cur.Next()
	}
	return &ListPair{
		Front: l,
		End:   pair.End,
	}, fill
}
