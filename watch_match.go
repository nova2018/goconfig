package goconfig

import (
	"regexp"
	"strings"
	"sync"
)

type matchKeyWatch struct {
	lock         *sync.Mutex
	key          *regexp.Regexp
	config       *Config
	mapWatchItem map[string]*watchItem
	notify       []func(e ConfigUpdateEvent)
}

func (w *matchKeyWatch) init() {
	w.reload()
}

func (w *matchKeyWatch) keys() []string {
	listKey := make([]string, 0)
	cfg := w.config.GetConfig()
	if cfg != nil {
		keys := cfg.AllKeys()
		for _, k := range keys {
			kk := keySlice(cfg, k)
			listKey = append(listKey, kk...)
		}
	}
	// 展开+去重
	mapUnique := make(map[string]bool)
	newListKey := make([]string, 0)
	for _, k := range listKey {
		p := strings.Split(k, ".")
		l := len(p)
		for i := 0; i < l; i++ {
			sub := p[0 : i+1]
			newKey := strings.Join(sub, ".")
			if !mapUnique[newKey] {
				mapUnique[newKey] = true
				if w.key.MatchString(newKey) {
					newListKey = append(newListKey, newKey)
				}
			}
		}
	}
	return newListKey
}

func (w *matchKeyWatch) reload() {
	listKey := w.keys()

	mapWatchItem := make(map[string]*watchItem, len(listKey))
	if len(listKey) > 0 {
		for _, k := range listKey {
			item := acquireWatchItem()
			item.key = k
			item.config = w.config
			item.reload()
			mapWatchItem[k] = item
		}
	}
	oldMapWatchItem := w.mapWatchItem
	w.mapWatchItem = mapWatchItem
	freeWatchItemMap(oldMapWatchItem)
}

func (w *matchKeyWatch) checkAndNotify() {
	w.lock.Lock()
	defer w.lock.Unlock()

	listFn := w.checkKeyAndNotify()
	if len(listFn) > 0 {
		w.reload()
	}
	for _, fn := range listFn {
		fn()
	}
}

func (w *matchKeyWatch) checkKeyAndNotify() []func() {
	keys := w.keys()
	hitKey := make(map[string]bool, len(keys))
	notify := func(key string, subKey string, op int8) {
		event := ConfigUpdateEvent{
			fullKey: key,
			key:     subKey,
			op:      op,
		}
		for _, fn := range w.notify {
			go fn(event)
		}
	}
	listFn := make([]func(), 0)
	for _, k := range keys {
		hitKey[k] = true
		if item, ok := w.mapWatchItem[k]; ok {
			// 存在，检查变更
			if item.isChange() {
				listFn = append(listFn, func() {
					notify(item.key, k, EventOpUpdate)
				})
			}
		} else {
			// 不存在，则新增
			listFn = append(listFn, func(key string) func() {
				return func() {
					notify(key, key, EventOpAdd)
				}
			}(k))
		}
	}
	for k, item := range w.mapWatchItem {
		if hitKey[k] {
			continue
		}
		// 没有命中，则移除
		listFn = append(listFn, func(subKey string, key string) func() {
			return func() {
				notify(key, subKey, EventOpDelete)
			}
		}(k, item.key))
	}
	return listFn
}
