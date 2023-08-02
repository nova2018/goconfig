package goconfig

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"log"
	"reflect"
	"sync"
	"time"
)

type Config struct {
	config         *viper.Viper
	lock           *sync.RWMutex
	isWatch        bool
	listWatch      []*configWatch
	onConfigUpdate func() // 变更通知
	listWatchKey   []*keyWatch
}

var c *Config

func init() {
	c = New()
}

type configWatch struct {
	viper  *viper.Viper
	watch  uint8 // 监控类型
	prefix string
	inited bool // 是否初始化
	enable bool
}

type keyWatch struct {
	lock     *sync.Mutex
	key      string
	update   []func()
	valViper *viper.Viper
	val      interface{}
	valHash  string
	config   *Config
}

func (w *keyWatch) init() {
	w.reload()
}

func (w *keyWatch) reload() {
	w.valViper = w.config.config.Sub(w.key)
	w.val = w.config.config.Get(w.key)
	if w.valViper != nil {
		w.valHash = genHash(w.valViper)
	}
}

func (w *keyWatch) isChange() bool {
	newVal := w.config.config.Get(w.key)
	newViper := w.config.config.Sub(w.key)
	if (newViper == nil || w.valViper == nil) && newViper != w.valViper {
		return true
	}
	if newViper != nil && w.valViper != nil {
		newHash := genHash(newViper)
		return w.valHash != newHash
	}
	return !reflect.DeepEqual(newVal, w.val)
}

func (w *keyWatch) checkAndNotify() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.isChange() {
		w.reload()
		for _, fn := range w.update {
			go fn()
		}
	}
}

const (
	WatchNone   uint8 = 0
	WatchFile   uint8 = 1
	WatchRemote uint8 = 2
)

var (
	watchHandler = map[uint8]func(*Config, *configWatch, uint8, bool){
		WatchFile:   watchViper,
		WatchRemote: watchViper,
	}
)

func New() *Config {
	return &Config{
		config:    viper.New(),
		lock:      &sync.RWMutex{},
		listWatch: make([]*configWatch, 0, 1),
		isWatch:   false,
	}
}

// Start 启动配置监控
func Start(watch ...bool) { c.Start(watch...) }

func StartWait() { c.StartWait() }

func Stop() { c.Stop() }

// Start 启动配置监控
func (c *Config) Start(watch ...bool) {
	if len(watch) == 0 {
		watch = []bool{false}
	}
	c.watch(watch[0])
	c.flush()
}

// StartWait 启动并等待加载完成
func (c *Config) StartWait() {
	c.Start(true)
}

func (c *Config) Stop() {
	c.watch(false)
	c.flush()
}

// 启动监控
func (c *Config) watch(isStart bool) {
	c.isWatch = isStart
	for _, v := range c.listWatch {
		for w, h := range watchHandler {
			if hasOp(v.watch, w) {
				h(c, v, w, isStart)
			}
		}
	}
}

func OnKeyChange(key string, fn func()) {
	c.OnKeyChange(key, fn)
}

func (c *Config) OnKeyChange(key string, fn func()) {
	c.lock.Lock()
	defer c.lock.Unlock()
	isHit := false
	for _, watch := range c.listWatchKey {
		if watch.key == key {
			watch.update = append(watch.update, fn)
			isHit = true
		}
	}
	if !isHit {
		watch := &keyWatch{
			key:    key,
			update: []func(){fn},
			config: c,
			lock:   &sync.Mutex{},
		}
		watch.init()
		c.listWatchKey = append(c.listWatchKey, watch)
	}
}

func genHash(v *viper.Viper) string {
	c := v.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("unable to marshal config to YAML: %v", err)
	}
	hash := md5.Sum(bs)
	return string(hash[:])
}

func GetConfig() *Config { return c }

func (c *Config) GetConfig() *viper.Viper {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.config
}

func AddWatchViper(watch uint8, v *viper.Viper, prefix ...string) {
	c.AddWatchViper(watch, v, prefix...)
}

func (c *Config) AddWatchViper(watch uint8, v *viper.Viper, prefix ...string) {
	if len(prefix) == 0 {
		prefix = []string{""}
	}
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.listWatch = append(c.listWatch, &configWatch{
		viper:  v,
		watch:  watch,
		prefix: prefix[0],
		enable: true,
	})
}

// AddViper 添加子viper
func AddViper(v *viper.Viper, prefix ...string) { c.AddViper(v, prefix...) }

// AddViper 添加子viper
func (c *Config) AddViper(v *viper.Viper, prefix ...string) {
	c.AddWatchViper(WatchFile, v, prefix...)
}

func AddNoWatchViper(v *viper.Viper, prefix ...string) { c.AddNoWatchViper(v, prefix...) }

func (c *Config) AddNoWatchViper(v *viper.Viper, prefix ...string) {
	c.AddWatchViper(WatchNone, v, prefix...)
}

func DelViper(v *viper.Viper) {
	c.DelViper(v)
}

func (c *Config) DelViper(v *viper.Viper) {
	c.lock.Lock()
	defer c.lock.Unlock()

	listWatch := make([]*configWatch, 0, len(c.listWatch)-1)
	for _, watch := range c.listWatch {
		if watch.viper != v {
			listWatch = append(listWatch, watch)
		} else {
			watch.enable = false
		}
	}
	c.listWatch = listWatch
}

// AddConfig 初始化并添加一个viper
func AddConfig(configType, configName string, configPath ...string) {
	c.AddConfig(configType, configName, configPath...)
}

// AddConfig 初始化并添加一个viper
func (c *Config) AddConfig(configType, configName string, configPath ...string) {
	v := viper.New()
	v.SetConfigType(configType)
	v.SetConfigName(configName)
	for _, p := range configPath {
		v.AddConfigPath(p)
	}
	c.AddViper(v)
}

// 重置全局viper
func (c *Config) setConfig(v *viper.Viper) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.config = v
}

func (c *Config) onViperChange(e fsnotify.Event) {
	if c.isWatch {
		c.flush()
	}
}

func (c *Config) notifyKeyUpdate() {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for _, watch := range c.listWatchKey {
		watch.checkAndNotify()
	}
}

func mergeConfig(viper, sourceViper *viper.Viper, prefix ...string) {
	if len(prefix) == 0 {
		prefix = []string{""}
	}
	viper.SetConfigType("yml")
	cfg := sourceViper.AllSettings()
	if prefix[0] != "" {
		cfg = map[string]interface{}{
			prefix[0]: cfg,
		}
	}
	bs, err := yaml.Marshal(cfg)
	if err != nil {
		log.Fatalf("unable to marshal config to YAML: %v", err)
	}
	_ = viper.MergeConfig(bytes.NewBuffer(bs))
}

// 更新配置
func (c *Config) flush() {
	newViper := viper.New()

	for _, w := range c.listWatch {
		if w.viper != nil {
			mergeConfig(newViper, w.viper, w.prefix)
		}
	}

	hash := genHash(c.config)
	newHash := genHash(newViper)
	if hash == newHash {
		// 无变化，则跳过
		return
	}

	c.setConfig(newViper)

	if c.onConfigUpdate != nil {
		c.onConfigUpdate()
	}

	if len(c.listWatchKey) > 0 {
		c.notifyKeyUpdate()
	}
}

func Equal(v1, v2 *viper.Viper) bool {
	return genHash(v1) == genHash(v2)
}

// 监控viper
func watchViper(c *Config, cfg *configWatch, watch uint8, isStart bool) {
	if cfg.viper == nil {
		return
	}
	switch watch {
	case WatchFile:
		_ = cfg.viper.ReadInConfig()
		if isStart {
			// 增加日志变更监控
			cfg.viper.OnConfigChange(c.onViperChange)
			cfg.viper.WatchConfig()
		} else {
			cfg.viper.OnConfigChange(nil)
		}
	case WatchRemote:
		if isStart {
			err := cfg.viper.ReadRemoteConfig()
			if err != nil {
				panic(err)
			}
			go func(v *configWatch) {
				ticker := time.Tick(time.Second)
				for c.isWatch && v.enable {
					select {
					case _ = <-ticker:
						e := v.viper.WatchRemoteConfig()
						if e != nil {
							fmt.Printf("%s viper remote listen failure! err=%v\n", time.Now().Format("2006-01-02 15:04:05"), e)
							continue
						}

						c.flush()
					}
				}
			}(cfg)
		}
	}
}

func hasOp(own, op uint8) bool {
	return (own & op) == op
}
