package ztredis

import (
	"errors"
	"gotraffic/ztpbf"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

func (p *Pool) AddAllholds(id int, t int, fspots string, lspots string) {
	if t == 1 {
		values := strings.Split(fspots, ",")
		key := "allhold.mid.f." + strconv.Itoa(id)
		p.ResetSet(key, values)
		values = strings.Split(lspots, ",")
		key = "allhold.mid.l." + strconv.Itoa(id)
		p.ResetSet(key, values)
	} else {
		values := strings.Split(fspots, ",")
		key := "allhold.end.f." + strconv.Itoa(id)
		p.ResetSet(key, values)
		values = strings.Split(lspots, ",")
		key = "allhold.end.l." + strconv.Itoa(id)
		p.ResetSet(key, values)
	}
}

func (p *Pool) AddSpotsNear(spot, nears string) {
	key := "spot.near." + spot
	values := strings.Split(nears, ",")
	p.ResetSet(key, values)
}

const luaArgvLen = 6

func (p *Pool) LockSpots(req *ztpbf.ReqPathSpots) (res *ztpbf.ReqPathSpots, err error) {
	script := redis.NewScript(len(req.PathSpotsLeft), LuaLockSpots)
	conn := p.Get()
	defer conn.Close()
	para := make([]interface{}, 0, luaArgvLen+len(req.PathSpotsLeft))
	for _, v := range req.PathSpotsLeft {
		para = append(para, v)
	}
	argv3 := req.ReqMinCount
	if int(req.ReqMinCount) == len(req.PathSpotsLeft) { //argv3 == 0 表示一次性请求到所有剩余的点位，如果没有请求到则回退；
		argv3 = 0 //len(req.PathNamesLeft)
	}
	argv5 := 0
	//没有请求到所有点，是移动指令，路径上有双向点
	if req.HasTwoWay && len(req.PathSpotsLeft) > 1 {
		argv5 = 1
	}
	//fmt.Println("DEBUG in LockSpots argv3, argv5:", argv3, argv5)
	para = append(para, req.RobotId, req.ExpireSeconds, argv3, req.MissionId, argv5, req.AgvType)
	ss, err := redis.Strings(script.Do(conn, para...))
	if nil != err {
		return nil, err
	}
	//...遍历点,进行资源状态检查，如果状态不对，不下发执行；
	req.RequestSpots = ss
	return req, nil
}

func (p *Pool) UnLockAllSpots(robotid string) (res *ztpbf.ReqPathSpots, err error) {
	req := &ztpbf.ReqPathSpots{RobotId: robotid}
	req.ExpireSeconds = 90
	return p.LockSpots(req)
}

// exceptKeys 包含spots.的前缀
func (p *Pool) UnlockSpotsExceptAndResetTraffic(robotid string, exceptKeys ...string) (int, error) {
	script := redis.NewScript(len(exceptKeys), LuaUnlockKeysExceptAndResetTraffic)
	conn := p.Get()
	defer conn.Close()
	para := make([]interface{}, 0, 1+len(exceptKeys))
	for _, v := range exceptKeys {
		para = append(para, v)
	}
	para = append(para, robotid)
	ret, err := redis.Int(script.Do(conn, para...))
	if nil != err {
		return 0, err
	}

	return ret, nil
}

// LockAttachSpotsOfRotates() keys are rotates, only one mostly
func (p *Pool) LockAttachSpotsOfRotates(value interface{}, secs int, keys ...string) (int, error) {
	script := redis.NewScript(len(keys), LuaLockAttachSpotsOfRotates)
	conn := p.Get()
	defer conn.Close()
	para := make([]interface{}, 0, 2+len(keys))
	for _, v := range keys {
		para = append(para, v)
	}
	para = append(para, value, secs)
	ret, err := redis.Int(script.Do(conn, para...))
	if nil != err {
		return 0, err
	}

	return ret, nil
}

func (p *Pool) UnlockRotateAttachSpots(robotid int32) (int, error) {
	script := redis.NewScript(0, LuaUnlockRotateAttachSpots)
	conn := p.Get()
	defer conn.Close()
	ret, err := redis.Int(script.Do(conn, robotid))
	if nil != err {
		return 0, err
	}
	return ret, nil
}

func (p *Pool) ReadHoldenResources() (map[string]string, error) {
	script := redis.NewScript(0, LuaReadHoldenResources)
	conn := p.Get()
	defer conn.Close()
	ret, err := redis.StringMap(script.Do(conn))
	if nil != err {
		return nil, err
	}
	return ret, nil
}

func (p *Pool) ReadAskedResources() (map[string]string, error) {
	script := redis.NewScript(0, LuaReadAskedResources)
	conn := p.Get()
	defer conn.Close()
	ret, err := redis.StringMap(script.Do(conn))
	if nil != err {
		return nil, err
	}
	return ret, nil
}

func (p *Pool) LockedKeys(value interface{}, secs int, keys ...string) (int, error) {
	script := redis.NewScript(len(keys), LuaLockKeys)
	conn := p.Get()
	defer conn.Close()
	para := make([]interface{}, 0, 2+len(keys))
	for _, v := range keys {
		para = append(para, v)
	}
	para = append(para, value, secs)
	ret, err := redis.Int(script.Do(conn, para...))
	if nil != err {
		return 0, err
	}
	return ret, nil
}

// HoldOneResource try to hold the resource for robotids
func (p *Pool) HoldOneResource(code string, maxnum int, secs int, robotids ...string) (ids []string, err error) {
	script := redis.NewScript(len(robotids), LuaHoldResource)
	conn := p.Get()
	defer conn.Close()
	para := make([]interface{}, 0, 3+len(robotids))
	for _, v := range robotids {
		para = append(para, v)
	}
	para = append(para, code, maxnum, secs)
	ret, err := redis.Strings(script.Do(conn, para...))
	if nil != err {
		return nil, err
	}
	return ret, nil
}

func (p *Pool) GetCandidateOfResourceHolder(code string, robotids ...string) (ids []string, err error) {
	script := redis.NewScript(len(robotids), LuaGetCandidateOfResourceHolder)
	conn := p.Get()
	defer conn.Close()
	para := make([]interface{}, 0, 1+len(robotids))
	for _, v := range robotids {
		para = append(para, v)
	}
	para = append(para, code)
	ret, err := redis.Strings(script.Do(conn, para...))
	if nil != err {
		return nil, err
	}
	return ret, nil
}

func (p *Pool) ReleaseResourcesUnholdenRobots(resRobs map[string]string) error {
	conn := p.Get()
	defer conn.Close()
	// 获取redis内资源点
	keys, err := redis.Strings(conn.Do("KEYS", "resource.*"))
	if nil != err {
		return err
	}
	var keysToDel []interface{}
	for _, k := range keys {
		code := string(k[9:])
		ids, ok := resRobs[code]
		if !ok {
			keysToDel = append(keysToDel, k)
			continue
		}
		ssids := strings.Split(ids, ",")
		nowids, err := redis.Strings(conn.Do("SMEMBERS", k))
		if nil != err {
			continue
		}
		//从resource.[code]删除不asked的robotid
		unexists := make([]interface{}, 1)
		unexists[0] = k
		for _, s := range nowids {
			bExit := false
			for _, s2 := range ssids {
				if s == s2 {
					bExit = true
					break
				}
			}
			if !bExit {
				unexists = append(unexists, s)
			}
		}
		if len(unexists) > 1 {
			conn.Do("SREM", unexists...)
		}
	}
	if len(keysToDel) > 0 {
		conn.Do("DEL", keysToDel...)
	}
	return nil
}

// ReleaseResourcesUnholden() release un asked resources in redis
func (p *Pool) ReleaseResourcesUnholden(robRes map[int][]string) error {
	conn := p.Get()
	defer conn.Close()
	//...
	keys, err := redis.Strings(conn.Do("KEYS", "resources.holdby.*"))
	if nil != err {
		return err
	}
	var keysToDel []interface{}
	for _, k := range keys {
		id, err := strconv.Atoi(k[17:])
		if nil != err {
			continue
		}
		vs, ok := robRes[id]
		if !ok {
			keysToDel = append(keysToDel, k)
			continue
		}
		//看看这个机器人不需要哪些资源了，并从set中移除
		nows, err := redis.Strings(conn.Do("SMEMBERS", k))
		if nil != err {
			continue
		}
		unexists := make([]interface{}, 1)
		unexists[0] = k
		for _, code := range nows {
			bExit := false
			for _, x := range vs {
				if x == code {
					bExit = true
					break
				}
			}
			if !bExit {
				unexists = append(unexists, code)
			}
		}
		if len(unexists) > 1 {
			conn.Do("SREM", unexists...)
		}
	}
	if len(keysToDel) > 0 {
		conn.Do("DEL", keysToDel...)
	}
	return nil
}

// FreeUnholdRobotOfResources ...
func (p *Pool) FreeUnholdRobotOfResources() (resHoldenRobots []string, err error) {
	script := redis.NewScript(0, LuaFreeRobotInHoldenResourece)
	conn := p.Get()
	defer conn.Close()
	resHoldenRobots, err = redis.Strings(script.Do(conn))
	if nil != err {
		return nil, err
	}
	return
}

func (p *Pool) ReadNoparkingSpots() ([]string, error) {
	conn := p.Get()
	defer conn.Close()
	bytesList, err := redis.ByteSlices(conn.Do("HGETALL", "noparking.spots"))
	if nil != err {
		return nil, err
	}

	var strList []string
	for i := range bytesList {
		strList = append(strList, string([]byte(bytesList[i])))
	}

	return strList, nil
}

func (p *Pool) ReadRotateSpotBorders(rotateSpot string) ([]string, error) {
	conn := p.Get()
	defer conn.Close()
	key := "rotate.borders." + rotateSpot
	borders, err := redis.Strings(conn.Do("SMEMBERS", key))
	if nil != err {
		return nil, err
	}

	return borders, nil
}

func (p *Pool) GetLiftFloor(code string) (floor int, err error) {
	conn := p.Get()
	defer conn.Close()
	key := "resourceInfo." + code
	field := "floor"
	var floors []int
	floors, err = conn.ReadHashIntValues(key, field)
	if nil != err {
		return
	}
	if len(floors) == 0 {
		err = errors.New("No floor value get from redis")
		return
	}
	floor = floors[0]
	return
}
