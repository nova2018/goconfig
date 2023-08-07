package goconfig

import (
	"crypto/md5"
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"log"
	"reflect"
	"strings"
)

func genHash(v *viper.Viper) string {
	c := v.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("unable to marshal config to YAML: %v", err)
	}
	hash := md5.Sum(bs)
	return string(hash[:])
}

func keySlice(viper *viper.Viper, key string, value ...reflect.Value) []string {
	if len(value) == 0 || value[0].IsZero() {
		v := viper.Get(key)
		if v == nil {
			return []string{}
		}
		value = []reflect.Value{reflect.ValueOf(v)}
	}
	listKey := make([]string, 0)
	switch value[0].Kind() {
	case reflect.Slice, reflect.Array:
		ml := value[0].Len()
		for i := 0; i < ml; i++ {
			elem := value[0].Index(i)
			newKey := fmt.Sprintf("%s.%d", key, i)
			r := keySlice(viper, newKey, elem)
			listKey = append(listKey, r...)
		}
	case reflect.Map:
		keys := viper.Sub(key).AllKeys()
		for _, k := range keys {
			newKeys := fmt.Sprintf("%s.%s", key, k)
			r := keySlice(viper, newKeys)
			listKey = append(listKey, r...)
		}
	default:
		return []string{key}
	}
	return listKey
}

func allKeys(cfg *viper.Viper) []string {
	listKey := make([]string, 0)
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
				newListKey = append(newListKey, newKey)
			}
		}
	}
	return newListKey
}
