package gotraffic

import (
	stdctx "context"
	"gotraffic/ztapi"
	"gotraffic/ztredis"

	"gitee.com/sienectagv/gozk/zredis"
	"gitee.com/sienectagv/gozk/zutils"
	"github.com/kataras/iris/v12"
)

type Master struct {
	redisPool *ztredis.Pool
	app       *iris.Application
	appAddr   string
	waitGroup *zutils.LoopGroup
}

func (m *Master) InitRedisPool(p *zredis.Pool) error {
	if nil == p {
		return zutils.ErrNullParam
	}
	m.redisPool = &ztredis.Pool{Pool: p}
	return nil
}

func (m *Master) InitIrisApp(addr string) error {
	if nil == m.app {
		m.app = iris.New()
	}
	m.appAddr = addr
	if len(m.appAddr) == 0 {
		m.appAddr = "localhost:8000" //default address
	}
	return nil
}

func (m *Master) Run() {
	if nil == m.waitGroup {
		m.waitGroup = zutils.NewLoopGroup()
	}
	if nil != m.app {
		apiIris := ztapi.NewProtoApiParty("/api/", m.redisPool)
		apiIris.InstallToApp(m.app)
		//
		m.waitGroup.AddAsyncBlock(
			func() {
				m.app.Listen(m.appAddr)
			},
			func() {
				m.app.Shutdown(stdctx.TODO())
			})
	}
	m.waitGroup.WaitForEnter("quit")
}
