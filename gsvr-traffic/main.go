package main

import (
	"gotraffic"

	"gitee.com/sienectagv/gozk/zcfg"
	"gitee.com/sienectagv/gozk/zredis"
)

type Config struct {
	Code  string        `ini:"code"`
	Redis zcfg.CfgRedis `ini:"redis"`
}

func main() {
	cfg := &Config{}
	zcfg.IniMapToCfg(cfg)
	master := &gotraffic.Master{}
	master.InitRedisPool(zredis.NewPool(cfg.Redis.AddrTcp))
	master.InitIrisApp()
	master.Run()
	// fmt.Println(*cfg)
	// waitGroup := zutils.NewLoopGroup()
	// waitGroup.GoLoop("redis",
	// 	func() int {
	// 		zlogger.Info("test loop")
	// 		return 10
	// 	},
	// 	time.Millisecond*100,
	// 	func() {})
	// waitGroup.WaitForEnter("quit")
}
