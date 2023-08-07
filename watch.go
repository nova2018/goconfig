package goconfig

import (
	"sync"
)

type keyWatch struct {
	*watchItem
	lock   *sync.Mutex
	notify []func() // 通知变更
	key    string
	config *Config
}

func (w *keyWatch) init() {
	w.reload()
}

func (w *keyWatch) reload() {
	w.watchItem.key = w.key
	w.watchItem.config = w.config
	w.watchItem.reload()
}

func (w *keyWatch) isChange() bool {
	return w.watchItem.isChange()
}

func (w *keyWatch) checkAndNotify() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.isChange() {
		w.reload()
		for _, fn := range w.notify {
			go fn()
		}
	}
}
