package goconfig

import (
	"github.com/spf13/viper"
	"reflect"
	"sync"
)

type watchItem struct {
	key       string       // 监控的键
	lastViper *viper.Viper // 上次的配置viper格式
	lastHash  string       // 上次配置的viper哈希
	lastVal   interface{}  // 上次的配置
	config    *Config
}

func (w *watchItem) reset() {
	w.key = ""
	w.lastViper = nil
	w.lastHash = ""
	w.lastVal = nil
	w.config = nil
}

func (w *watchItem) reload() {
	w.lastViper = w.config.GetConfig().Sub(w.key)
	w.lastVal = w.config.GetConfig().Get(w.key)
	if w.lastViper != nil {
		w.lastHash = genHash(w.lastViper)
	}
}

func (w *watchItem) isChange() bool {
	newVal := w.config.GetConfig().Get(w.key)
	newViper := w.config.GetConfig().Sub(w.key)
	if (newViper == nil || w.lastViper == nil) && newViper != w.lastViper {
		// 如果均为空，可能是由于key对应的配置不是map，因此不能认为全为空表示不变
		return true
	}
	if newViper != nil && w.lastViper != nil {
		newHash := genHash(newViper)
		return w.lastHash != newHash
	}
	// 处理均不是map的情况
	return !reflect.DeepEqual(newVal, w.lastVal)
}

var (
	watchItemPool = sync.Pool{New: func() any {
		return &watchItem{}
	}}
)

func acquireWatchItem() *watchItem {
	item := watchItemPool.Get().(*watchItem)
	item.reset()
	return item
}

func freeWatchItem(item *watchItem) {
	watchItemPool.Put(item)
}

func freeWatchItemMap(m map[string]*watchItem) {
	for _, x := range m {
		freeWatchItem(x)
	}
}
