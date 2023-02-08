package main

import (
	"time"

	"gitee.com/sienectagv/gozk/zcfg"
	"gitee.com/sienectagv/gozk/zlogger"
	"gitee.com/sienectagv/gozk/zredis"
	"gitee.com/sienectagv/gozk/zutils"
)

type Config struct {
	Code  string        `ini:"code"`
	Redis zcfg.CfgRedis `ini:"redis"`
}

func main() {
	cfg := &Config{}
	zcfg.IniMapToCfg(cfg)
	redisPool := zredis.NewPool(cfg.Redis.AddrTcp)
	defer redisPool.Close()
	// fmt.Println(*cfg)
	waitGroup := zutils.NewLoopGroup()
	waitGroup.GoLoop("redis",
		func() int {
			zlogger.Info("test loop")
			return 10
		},
		time.Millisecond*100,
		func() {})
	waitGroup.WaitForEnter("quit")
}
