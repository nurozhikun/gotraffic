package main

import "gitee.com/sienectagv/gozk/zutils"

func main() {
	waitGroup := zutils.NewLoopGroup()
	waitGroup.WaitForEnter(("quit"))
}
