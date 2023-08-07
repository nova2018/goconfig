package goconfig

import (
	"github.com/spf13/viper"
	"regexp"
	"sync"
)

type Config struct {
	config            *viper.Viper
	cLock             *sync.RWMutex
	kLock             *sync.RWMutex
	listWatchKey      []*keyWatch
	listWatchMapKey   []*mapKeyWatch
	listWatchMatchKey []*matchKeyWatch
}

var _c *Config

func init() {
	_c = New()
}

func New() *Config {
	return &Config{
		config: viper.New(),
		cLock:  &sync.RWMutex{},
		kLock:  &sync.RWMutex{},
	}
}

func OnKeyChange(key string, fn func()) {
	_c.OnKeyChange(key, fn)
}

func OnMapKeyChange(key string, fn func(e ConfigUpdateEvent)) {
	_c.OnMapKeyChange(key, fn)
}

func OnMatchKeyChange(key string, fn func(e ConfigUpdateEvent)) {
	_c.OnMapKeyChange(key, fn)
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
			watchItem: acquireWatchItem(),
			key:       key,
			notify:    []func(){fn},
			config:    c,
			lock:      &sync.Mutex{},
		}
		watch.init()
		c.listWatchKey = append(c.listWatchKey, watch)
	}
}

func (c *Config) OnMapKeyChange(key string, fn func(e ConfigUpdateEvent)) {
	c.kLock.Lock()
	defer c.kLock.Unlock()
	isHit := false
	for _, watch := range c.listWatchMapKey {
		if watch.key == key {
			watch.notify = append(watch.notify, fn)
			isHit = true
		}
	}
	if !isHit {
		mapWatch := &mapKeyWatch{
			keyWatch: &keyWatch{
				watchItem: acquireWatchItem(),
				key:       key,
				config:    c,
				lock:      &sync.Mutex{},
			},
			mapWatchItem: make(map[string]*watchItem, 1),
			notify:       []func(e ConfigUpdateEvent){fn},
		}
		mapWatch.init()
		c.listWatchMapKey = append(c.listWatchMapKey, mapWatch)
	}
}

func (c *Config) OnMatchKeyChange(key *regexp.Regexp, fn func(e ConfigUpdateEvent)) {
	c.kLock.Lock()
	defer c.kLock.Unlock()
	isHit := false
	for _, watch := range c.listWatchMatchKey {
		if watch.key == key || key.String() == watch.key.String() {
			watch.notify = append(watch.notify, fn)
			isHit = true
		}
	}
	if !isHit {
		matchWatch := &matchKeyWatch{
			key:          key,
			config:       c,
			lock:         &sync.Mutex{},
			mapWatchItem: make(map[string]*watchItem, 1),
			notify:       []func(e ConfigUpdateEvent){fn},
		}
		matchWatch.init()
		c.listWatchMatchKey = append(c.listWatchMatchKey, matchWatch)
	}
}

func GetConfig() *viper.Viper {
	return _c.GetConfig()
}

func (c *Config) GetConfig() *viper.Viper {
	c.cLock.RLock()
	defer c.cLock.RUnlock()
	return c.config
}

func SetConfig(v *viper.Viper) {
	_c.SetConfig(v)
}

// SetConfig 重置全局viper
func (c *Config) SetConfig(v *viper.Viper) {
	c.cLock.Lock()
	c.config = v
	c.cLock.Unlock()

	if len(c.listWatchKey) > 0 ||
		len(c.listWatchMapKey) > 0 ||
		len(c.listWatchMatchKey) > 0 {
		c.notifyKeyUpdate()
	}
}

func (c *Config) notifyKeyUpdate() {
	c.kLock.RLock()
	defer c.kLock.RUnlock()
	for _, watch := range c.listWatchKey {
		watch.checkAndNotify()
	}
	for _, watch := range c.listWatchMapKey {
		watch.checkAndNotify()
	}
	for _, watch := range c.listWatchMatchKey {
		watch.checkAndNotify()
	}
}
