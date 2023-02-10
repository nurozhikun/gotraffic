package redistraffic

import (
	"nj.gitlab.com/nj-tools/zkredis"
)

//Traffic is a struct for redis of road traffic
type RdsTraffic struct {
	*zkredis.RedisPool
}

func New(addr *zkredis.RedisNode) (r *RdsTraffic, err error) {
	r = &RdsTraffic{}
	r.RedisPool, err = zkredis.NewRedisPool(addr)
	return r, err
}
