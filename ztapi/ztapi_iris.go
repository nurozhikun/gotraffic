package ztapi

import (
	"gotraffic/ztpbf"
	"gotraffic/ztredis"

	"gitee.com/sienectagv/gozk/zproto"
	"gitee.com/sienectagv/gozk/zproto/zpbf"
	"gitee.com/sienectagv/gozk/zutils"
)

func NewProtoApiParty(partyUrl string, p *ztredis.Pool) *zproto.ProtoApiParty {
	api := &zproto.ProtoApiParty{
		PartyUrl: partyUrl,
		Handler:  &apiIris{Pool: p},
		Cmds:     make(map[int]*zproto.Command),
	}
	api.Cmds[CmdRequestPathSpots] = &zproto.Command{
		Cmd:        CmdRequestPathSpots,
		Path:       "lockspots",
		MethodName: "ApiRequestPathSpots",
		CreateRequestBody: func() zproto.Message {
			return &ztpbf.ReqPathSpots{}
		},
	}
	return api
}

type apiIris struct {
	*ztredis.Pool
}

// func (ai *apiIris) ReqBodyOfCmd(cmd int) zproto.Message {
// 	switch cmd {
// 	case CmdRequestPathSpots:
// 		return &ztpbf.ReqPathSpots{}
// 	}
// 	return nil
// }

func (ai *apiIris) ApiRequestPathSpots(h *zpbf.Header, req zproto.Message) (zproto.Message, error) {
	reqSpots, ok := req.(*ztpbf.ReqPathSpots)
	if !ok {
		return nil, zutils.ErrWrongReqBody
	}
	return ai.LockSpots(reqSpots)
}
