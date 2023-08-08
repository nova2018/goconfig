package goconfig

import (
	"regexp"
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
	w.reload(w.keys())
}

func (w *matchKeyWatch) reload(listKey []string) {
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

func (w *matchKeyWatch) keys() []string {
	keys := w.config.keys()

	listKey := make([]string, 0, len(keys))
	for _, k := range keys {
		if w.key.MatchString(k) {
			listKey = append(listKey, k)
		}
	}
	return listKey
}

func (w *matchKeyWatch) checkAndNotify() {
	w.lock.Lock()
	defer w.lock.Unlock()
	listKey := w.keys()

	listFn := w.checkKeyAndNotify(listKey)
	if len(listFn) > 0 {
		w.reload(listKey)
	}
	for _, fn := range listFn {
		fn()
	}
}

func (w *matchKeyWatch) checkKeyAndNotify(keys []string) []func() {
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
