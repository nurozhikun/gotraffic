// 对资源请求的占用
package ztres

import "gotraffic/ztredis"

type Res struct {
	*ztredis.Pool
}
