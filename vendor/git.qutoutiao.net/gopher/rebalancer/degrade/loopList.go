package degrade

import (
	"container/list"
	"time"
)

type Elem struct {
	UnixNano int64 // 存入时的时间戳
	List     []*Node
}

type loopList struct {
	list *list.List
}

func newLoopList() *loopList {
	return &loopList{
		list: list.New(),
	}
}

func (l *loopList) push(elnow *Elem, opts DegradeOpts) {
	// 删除过期数据
	l.delTimeout(elnow.UnixNano, opts)
	// 当 elnow 的 len(List) 为空的时候，不存放历史数据
	if len(elnow.List) <= 0 {
		return
	}
	// 距离上次插入数据间隔大于 EndpointsSaveInterval 才允许追加写入
	el := l.list.Front()
	if el != nil {
		elv := el.Value.(*Elem)
		if elnow.UnixNano-elv.UnixNano < int64(opts.EndpointsSaveInterval) {
			return
			//l.list.Remove(el)
		}
	}
	l.list.PushFront(elnow)
}

func (l *loopList) back(opts DegradeOpts) *Elem {
	now := time.Now().UnixNano()
	// 删除过期数据
	l.delTimeout(now, opts)
	if l.list.Len() > 0 {
		el := l.list.Back()
		if el == nil {
			return nil
		}
		return el.Value.(*Elem)
	}
	return nil
}

func (l *loopList) reset() {
	l.list = list.New()
}

func (l *loopList) delTimeout(now int64, opts DegradeOpts) {
	for {
		el := l.list.Back()
		if el == nil {
			break
		}
		pre := el.Prev()
		if pre == nil {
			break
		}
		// 至少保留一个15m前数据, 后一个插入的数据必须已经超过 15 分钟才能删除当前的
		if now-pre.Value.(*Elem).UnixNano <= int64(opts.ThresholdContrastInterval) {
			break
		}
		l.list.Remove(el)
	}
}
