package goconfig

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {

	v := viper.New()
	v.SetConfigType("yml")

	var yamlExample = []byte(`
Hacker: true
name: steve
hobbies:
- skateboarding
- snowboarding
- go
clothing:
  jacket: leather
  trousers: denim
age: 35
eyes : brown
beard: true
`)

	_ = v.ReadConfig(bytes.NewBuffer(yamlExample))

	fmt.Println("v.AllSettings:", v.AllSettings())

	AddNoWatchViper(v)

	StartWait()

	OnKeyChange("abc1", func() {
		fmt.Printf("======= update!\n")
	})

	fmt.Println("0app.AllSettings:", GetConfig().AllSettings())

	for {
		time.Sleep(1 * time.Second)
		//	fmt.Printf("%v 1app.AllSettings:%v\n", time.Now(), GetConfig().AllSettings())
	}

}
