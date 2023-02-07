package main

import (
	"fmt"

	"gitee.com/sienectagv/gozk/zcfg"
)

type Config struct {
	Code  string        `ini:"code"`
	Redis zcfg.CfgRedis `ini:"redis"`
}

func main() {
	cfg := &Config{}
	zcfg.IniMapToCfg(cfg)
	fmt.Println(*cfg)
	// waitGroup := zutils.NewLoopGroup()
	// waitGroup.WaitForEnter("quit")
}
