package ztredis

import "gitee.com/sienectagv/gozk/zredis"

type Pool struct {
	*zredis.Pool
}

func Create(p *zredis.Pool) *Pool {
	return &Pool{Pool: p}
}
