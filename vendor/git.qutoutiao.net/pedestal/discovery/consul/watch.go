package consul

import (
	"context"
	"fmt"
	"os"

	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/logger"
	"git.qutoutiao.net/pedestal/discovery/logger/hclog"
	"git.qutoutiao.net/pedestal/discovery/rolling"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

type watchKey struct {
	dc   string
	name string
	tags []string
}

type watchResult struct {
	key     watchKey
	index   uint64
	entries []*api.ServiceEntry
}

type Watch struct {
	client     *api.Client
	opts       *option
	key        *watchKey
	resultChan chan *watchResult

	// internal
	plan      *watch.Plan
	filter    Filter
	backoff   *rolling.Backoff
	lastIndex uint64
}

func NewWatch(client *api.Client, opts *option, key *watchKey, resultChan chan *watchResult) *Watch {
	watcher := &Watch{
		client:     client,
		opts:       opts,
		key:        key,
		filter:     newPassingOnlyFilter(key, opts),
		resultChan: resultChan,
	}

	return watcher
}

func (w *Watch) option() *option {
	return w.opts
}

func (w *Watch) isDebug() bool {
	return w.opts.debug
}

func (w *Watch) consul() *api.Client {
	return w.client
}

func (w *Watch) Watch() error {
	logger.Infof("consul.Watch(%+v) ...", w.key)

	params := map[string]interface{}{
		"type":    "service",
		"service": w.key.name,
		"stale":   w.option().stale,
	}
	if len(w.key.tags) > 0 {
		params["tag"] = w.key.tags
	}

	plan, err := watch.Parse(params)
	if err != nil {
		logger.Errorf("watch.Parse(%+v): %v", params, err)

		return errors.Wrap(err)
	}

	plan.Datacenter = w.key.dc
	plan.Handler = w.Handler
	plan.Watcher = w.Watcher

	hlog := hclog.New(os.Stderr)

	w.plan = plan
	w.backoff = rolling.NewBackoffWithLogger(w.opts.watchTimeout, w.opts.watchLatency, logger.NewWithHclog(hlog))

	err = plan.RunWithClientAndHclog(w.consul(), hlog)
	if err != nil {
		logger.Errorf("consul.Watch(%+v): %v", w.key, err)

		return errors.Wrap(err)
	}

	return nil
}

func (w *Watch) Handler(idx uint64, value interface{}) {
	entries, ok := value.([]*api.ServiceEntry)
	if !ok {
		return
	}

	if len(entries) == 0 {
		logger.Warnf("consul.Handler(%+v, %d): empty entry, ignored!", w.key, idx)
		return
	}

	var addrs = make([]string, len(entries))
	for i, entry := range entries {
		addrs[i] = fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port)
	}
	logger.Infof("consul.Handler(%+v, %d): total=%d, addresses=%v", w.key, idx, len(addrs), addrs)

	if w.filter != nil {
		entries = w.filter.Apply(entries)
	}

	w.resultChan <- &watchResult{
		key:     *w.key,
		index:   idx,
		entries: entries,
	}

	w.backoff.Delay("consul.Watch", w.key.name)

	return
}

func (w *Watch) Watcher(plan *watch.Plan) (watch.BlockingParamVal, interface{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := &api.QueryOptions{
		Datacenter: w.key.dc,
		AllowStale: w.option().stale,
		WaitIndex:  w.lastIndex,
		UseCache:   w.option().agentCache,
		WaitTime:   w.option().watchTimeout,
	}
	opts = opts.WithContext(ctx)

	entries, meta, err := w.consul().Health().ServiceMultipleTags(w.key.name, w.key.tags, false, opts)
	if err != nil {
		return nil, nil, err
	}

	if w.isDebug() {
		logger.Debugf("consul.Watcher(%+v): prev index: %v, last index: %v, entries: %d",
			w.key, w.lastIndex, meta.LastIndex, len(entries))
	}

	if w.lastIndex <= 0 {
		logger.Infof("consul.Watcher(%+v): prev index: %v, last index: %v, entries: %d",
			w.key, w.lastIndex, meta.LastIndex, len(entries))

		w.Handler(meta.LastIndex, entries)
	}

	// update lastIndex
	w.lastIndex = meta.LastIndex

	return watch.WaitIndexVal(w.lastIndex), entries, err
}

func (w *Watch) Stop() {
	if w.plan == nil {
		return
	}

	w.plan.Stop()
}
