package goconfig

import (
	"github.com/spf13/viper"
	"sync"
)

type Config struct {
	config       *viper.Viper
	cLock        *sync.RWMutex
	kLock        *sync.RWMutex
	listWatchKey []*keyWatch
}

var _c *Config

func init() {
	_c = New()
}

func New() *Config {
	return &Config{
		config:       viper.New(),
		cLock:        &sync.RWMutex{},
		kLock:        &sync.RWMutex{},
		listWatchKey: make([]*keyWatch, 0),
	}
}

func OnKeyChange(key string, fn func()) {
	_c.OnKeyChange(key, fn)
}

func (c *Config) OnKeyChange(key string, fn func()) {
	c.kLock.Lock()
	defer c.kLock.Unlock()
	isHit := false
	for _, watch := range c.listWatchKey {
		if watch.key == key {
			watch.notify = append(watch.notify, fn)
			isHit = true
		}
	}
	if !isHit {
		watch := &keyWatch{
			key:    key,
			notify: []func(){fn},
			config: c,
			lock:   &sync.Mutex{},
		}
		watch.init()
		c.listWatchKey = append(c.listWatchKey, watch)
	}
}

func (c *Config) GetConfig() *viper.Viper {
	c.cLock.RLock()
	defer c.cLock.RUnlock()
	return c.config
}

// SetConfig 重置全局viper
func (c *Config) SetConfig(v *viper.Viper) {
	c.cLock.Lock()
	c.config = v
	c.cLock.Unlock()

	if len(c.listWatchKey) > 0 {
		c.notifyKeyUpdate()
	}
}

func (c *Config) notifyKeyUpdate() {
	c.kLock.RLock()
	defer c.kLock.RUnlock()
	for _, watch := range c.listWatchKey {
		watch.checkAndNotify()
	}
}
