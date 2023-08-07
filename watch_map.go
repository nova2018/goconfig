package goconfig

import (
	"fmt"
	"reflect"
)

type mapKeyWatch struct {
	*keyWatch
	mapWatchItem map[string]watchItem
	notify       []func(e ConfigUpdateEvent)
}

func (w *mapKeyWatch) init() {
	w.reload()
}

func (w *mapKeyWatch) reload() {
	w.keyWatch.reload()
	listKey := w.keys(w.lastVal)

	w.mapWatchItem = make(map[string]watchItem, len(listKey))
	if len(listKey) > 0 {
		for _, k := range listKey {
			item := watchItem{
				key:    fmt.Sprintf("%s.%s", w.key, k),
				config: w.config,
			}
			item.reload()
			w.mapWatchItem[k] = item
		}
	}
}

func (w *mapKeyWatch) keys(v interface{}) []string {
	var listKey []string
	if v == nil {
		return listKey
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Map:
		keys := reflect.ValueOf(v).MapKeys()
		listKey = make([]string, 0, len(keys))
		for _, k := range keys {
			listKey = append(listKey, fmt.Sprintf("%v", k.Interface()))
		}
	case reflect.Slice, reflect.Array:
		maxLen := reflect.ValueOf(v).Len()
		listKey = make([]string, 0, maxLen)
		for i := 0; i < maxLen; i++ {
			listKey = append(listKey, fmt.Sprintf("%d", i))
		}
	}
	return listKey
}

func (w *mapKeyWatch) checkAndNotify() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.isChange() {
		listFn := w.checkKeyAndNotify()
		w.reload()
		for _, fn := range listFn {
			fn()
		}
	}
}

func (w *mapKeyWatch) checkKeyAndNotify() []func() {
	config := w.config.GetConfig().Get(w.key)
	keys := w.keys(config)
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
			key := fmt.Sprintf("%s.%s", w.key, k)
			listFn = append(listFn, func() {
				notify(key, k, EventOpAdd)
			})
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
