package goconfig

import (
	"crypto/md5"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"log"
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
