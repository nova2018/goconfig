package goconfig

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"regexp"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	toml := []byte(`
[[a.x]]
abc=1
[[a.x]]
abc=2
[[a.x]]
abc=3

[b]
1=5
2=3
c=6
`)
	v := viper.New()
	v.SetConfigType("toml")
	_ = v.ReadConfig(bytes.NewBuffer(toml))

	cfg := New()
	cfg.SetConfig(v)

	cfg.OnKeyChange("a.x", func() {
		fmt.Println("a.x update")
	})

	cfg.OnMapKeyChange("a.x", func(e ConfigUpdateEvent) {
		fmt.Println("a.x map key update", e)
	})

	cfg.OnMapKeyChange("b", func(e ConfigUpdateEvent) {
		fmt.Println("b map key update", e)
	})

	pattern, _ := regexp.Compile(`^c\.1$`)
	cfg.OnMatchKeyChange(pattern, func(e ConfigUpdateEvent) {
		fmt.Println("c.1 match key update", e)
	})
	pattern2, _ := regexp.Compile(`^c`)
	cfg.OnMatchKeyChange(pattern2, func(e ConfigUpdateEvent) {
		fmt.Println("c match key update", e)
	})

	toml2 := []byte(`
[[a.x]]
abc=1
[[a.x]]
abc=2

[[c]]
x=1
[[c]]
x=2
[[c]]
x=3
[cc]
x=2
`)
	v2 := viper.New()
	v2.SetConfigType("toml")
	_ = v2.ReadConfig(bytes.NewBuffer(toml2))
	cfg.SetConfig(v2)

	time.Sleep(time.Second)
	fmt.Println("==============")

	toml3 := []byte(`
[[c]]
x=1
[[c]]
x=2
[[c]]
x=3
[cc]
x=1
`)
	v3 := viper.New()
	v3.SetConfigType("toml")
	_ = v3.ReadConfig(bytes.NewBuffer(toml3))
	cfg.SetConfig(v3)

	time.Sleep(time.Second)
	fmt.Println("==============")

	cfg.SetConfig(v2)

	time.Sleep(time.Second)
	fmt.Println("==============")

	cfg.SetConfig(v)

	time.Sleep(time.Second)
}
