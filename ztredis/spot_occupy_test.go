package ztredis

import (
	"gotraffic/ztpbf"
	"strconv"
	"testing"

	"gitee.com/sienectagv/gozk/zredis"
)

/*
*
1. 点位占用之非停靠点测试:
设置p4, p5为非停靠点
robot1：  p1 → p2 → p3 → p4 → p5 → p6 → p7 → p8
*/
func TestSpotOccupy_NoParkingSpots(t *testing.T) {
	r := initialRedisT()
	defer r.Close()

	c := r.Get()
	c.Do("flushdb")
	defer c.Close()

	// 设置非停靠点
	c.Do("SADD", "noparking.spots", "p4")
	c.Do("SADD", "noparking.spots", "p5")

	robotid := 1
	missionid := 101
	path := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8"}
	holdPath, err := holdPathT(robotid, missionid, path, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robotid, err)
	}

	expectHoldPath := []string{"p1", "p2", "p3", "p4", "p5", "p6"}
	success := assertHoldPathT(holdPath, expectHoldPath)
	if !success {
		t.Errorf("hold path: %v , but expect hold path: %v", holdPath, expectHoldPath)
	}
}

/*
*
2. 点位占用之全路径测试
全路径设置：p4为起点，p8为终点
robot1：  p1 → p2 → p3 → p4 → p5 → p6 → p7 → p8
*/
func TestSpotOccupy_HoldAllSpots(t *testing.T) {
	r := initialRedisT()
	defer r.Close()

	c := r.Get()
	c.Do("flushdb")
	defer c.Close()

	// 设置全路径
	first := []string{"p4"}
	last := []string{"p8"}
	r.ResetSet("allhold.end.f.1", first)
	r.ResetSet("allhold.end.l.1", last)

	robotid := 1
	missionid := 101
	path := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8"}
	holdPath, err := holdPathT(robotid, missionid, path, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robotid, err)
	}

	expectHoldPath := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8"}
	success := assertHoldPathT(holdPath, expectHoldPath)
	if !success {
		t.Errorf("hold path: %v , but expect hold path: %v", holdPath, expectHoldPath)
	}
}

/*
*
3. 点位占用之全路径+非停靠点测试
全路径设置：p4为起点，p8为终点
非停靠点设置：p3, p8
robot1：  p1 → p2 → p3 → p4 → p5 → p6 → p7 → p8
expect:   p1 -> p2
*/
func TestSpotOccupy_HoldAllSpotsAndNoParkingSpot(t *testing.T) {
	r := initialRedisT()
	defer r.Close()

	c := r.Get()
	c.Do("flushdb")
	defer c.Close()

	// 设置全路径
	first := []string{"p4"}
	last := []string{"p8"}
	r.ResetSet("allhold.end.f.1", first)
	r.ResetSet("allhold.end.l.1", last)

	// 设置非停靠点
	c.Do("SADD", "noparking.spots", "p3")
	c.Do("SADD", "noparking.spots", "p8")

	robotid := 1
	missionid := 101
	path := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8"}
	holdPath, err := holdPathT(robotid, missionid, path, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robotid, err)
	}

	expectHoldPath := []string{"p1", "p2"}
	success := assertHoldPathT(holdPath, expectHoldPath)
	if !success {
		t.Errorf("hold path: %v , but expect hold path: %v", holdPath, expectHoldPath)
	}
}

/*
*
4. 点位占用之双向路径 + 邻近点测试

地图为：

	    								 p27
	   								  ↑
	   								 p26        p44
	   								  ↑          ↑
	   								 p25        p43
	   								  ↑          ↑
	   								 p24        p42
	   									↑          ↑
	   								 p23        p41
	                     ↑          ↑
	p1 → p2 → p3 → p4 → p5 ←→ p6 ←→ p7 ←→ p8 ←→ p9 ←→ p10 ←→ p11 → p12 → p13
	                     ↑  					                         ↑
	                     ↓                                    ↓
	                    p31          	                      p22
	 				            ↑                                    ↑
	                     ↓                                    ↓
	                    p32                                  p21
	                     ↑                                    ↑
	                     ↓                                    ↓
	                    p33                                  p20

邻近点设置：p22 -> p11
robot1 : p4 -> p5 -> p6 -> p7 -> p8 -> p9 -> p10 -> p11 -> p12 - p13
robot2 : p21 -> p22 -> p11 -> p10 -> p9 -> p8 -> p7 -> p41 -> p42 -> p43 -> p44
*/
func TestSpotOccupy_SsPathAndNearSpot(t *testing.T) {
	r := initialRedisT()
	defer r.Close()

	c := r.Get()
	c.Do("flushdb")
	defer c.Close()

	// 设置邻近点
	nearSpots := []string{"p11"}
	r.ResetSet("spot.near.p22", nearSpots)

	// robot1 占点
	robot1id := 1
	mission1id := 101
	path1 := []string{"p4", "p5", "p6", "p7", "p8", "p9", "p10", "p11", "p12", "p13"}
	holdPath1, err := holdPathT(robot1id, mission1id, path1, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot1id, err)
	}

	expectHoldPath1 := []string{"p4", "p5", "p6", "p7"}
	success1 := assertHoldPathT(holdPath1, expectHoldPath1)
	if !success1 {
		t.Errorf("hold path1: %v , but expect hold path1: %v", holdPath1, expectHoldPath1)
	}

	// robot2 占点
	robot2id := 2
	mission2id := 102
	path2 := []string{"p21", "p22", "p11", "p10", "p9", "p8", "p7", "p41", "p42", "p43", "p44"}
	holdPath2, err := holdPathT(robot2id, mission2id, path2, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot2id, err)
	}

	expectHoldPath2 := []string{"p21"}
	success2 := assertHoldPathT(holdPath2, expectHoldPath2)
	if !success2 {
		t.Errorf("hold path2: %v , but expect hold path2: %v", holdPath2, expectHoldPath2)
	}

}

/*
*
5.1. 点位占用之集合测试：插队功能

全路径设置：  p42为起点，p44为终点
非停靠点设置：
邻近点设置：  p22 -> p11

robot1：  p21 -> p22 -> p11 -> p10 -> p9 -> p8 -> p7 → p41 -> p42 -> p43 -> p44
expect1:  p21 -> p22 -> p11 -> p10
robot2:   p4 -> p5 -> p6 -> p7 -> p8 -> p9 -> p10 -> p11 -> p12 - p13
expect2:  p4 -> p5 -> p6
robot3:   p52 -> p9 -> p8 -> p7 -> p41 -> p42 -> p43 -> p44
expect3:  p52 -> p9 -> p8 -> p7
*/
func TestSpotOccupy_MixtureAll_QueueJumper(t *testing.T) {
	r := initialRedisT()
	defer r.Close()

	c := r.Get()
	c.Do("flushdb")
	defer c.Close()

	// 设置全路径
	first := []string{"p42"}
	last := []string{"p44"}
	r.ResetSet("allhold.end.f.1", first)
	r.ResetSet("allhold.end.l.1", last)

	// 设置邻近点
	nearSpots1 := []string{"p11"}
	r.ResetSet("spot.near.p22", nearSpots1)

	// robot1 占点
	robot1id := 1
	mission1id := 101
	path1 := []string{"p21", "p22", "p11", "p10", "p9", "p8", "p7", "p41", "p42", "p43", "p44"}
	holdPath1, err := holdPathT(robot1id, mission1id, path1, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot1id, err)
	}

	expectHoldPath1 := []string{"p21", "p22", "p11", "p10"}
	success1 := assertHoldPathT(holdPath1, expectHoldPath1)
	if !success1 {
		t.Errorf("hold path1: %v , but expect hold path1: %v", holdPath1, expectHoldPath1)
	}

	// robot2 占点
	robot2id := 2
	mission2id := 102
	path2 := []string{"p4", "p5", "p6", "p7", "p8", "p9", "p10", "p11", "p12", "p13"}
	holdPath2, err := holdPathT(robot2id, mission2id, path2, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot2id, err)
	}

	expectHoldPath2 := []string{"p4", "p5", "p6"}
	success2 := assertHoldPathT(holdPath2, expectHoldPath2)
	if !success2 {
		t.Errorf("hold path2: %v , but expect hold path2: %v", holdPath2, expectHoldPath2)
	}

	// robot3 占点
	robot3id := 3
	mission3id := 103
	path3 := []string{"p52", "p9", "p8", "p7", "p41", "p42", "p43", "p44"}
	holdPath3, err := holdPathT(robot3id, mission3id, path3, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot3id, err)
	}

	expectHoldPath3 := []string{"p52", "p9", "p8", "p7"}
	success3 := assertHoldPathT(holdPath3, expectHoldPath3)
	if !success3 {
		t.Errorf("hold path3: %v , but expect hold path3: %v", holdPath3, expectHoldPath3)
	}
}

/*
*
5.2. 点位占用: 双向路径上有车停靠的情况

robot1：  p8
expect1:  p8
robot2:   p3 -> p4 -> p5 -> p6 -> p7 -> p8 -> p9 -> p10 -> p11 -> p12 - p13
expect2:  p3 -> p4 -> p5 -> p6
robot3:   p52 -> p9 -> p8 -> p7 -> p41 -> p42 -> p43 -> p44
expect3:  p52
robot2:   p4 -> p5 -> p6 -> p7 -> p8 -> p9 -> p10 -> p11 -> p12 - p13
expect2:  p4 -> p5 -> p6
*/
func TestSpotOccupy_MixtureAll_RobotStopOnSS(t *testing.T) {
	r := initialRedisT()
	defer r.Close()

	c := r.Get()
	c.Do("flushdb")
	defer c.Close()

	// robot1 占点
	robot1id := 1
	mission1id := 101
	path1 := []string{"p8"}
	holdPath1, err := holdPathT(robot1id, mission1id, path1, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot1id, err)
	}

	expectHoldPath1 := []string{"p8"}
	success1 := assertHoldPathT(holdPath1, expectHoldPath1)
	if !success1 {
		t.Errorf("hold path1: %v , but expect hold path1: %v", holdPath1, expectHoldPath1)
	}

	// robot2 占点
	robot2id := 2
	mission2id := 102
	path2 := []string{"p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10", "p11", "p12", "p13"}
	holdPath2, err := holdPathT(robot2id, mission2id, path2, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot2id, err)
	}

	expectHoldPath2 := []string{"p3", "p4", "p5", "p6"}
	success2 := assertHoldPathT(holdPath2, expectHoldPath2)
	if !success2 {
		t.Errorf("hold path2: %v , but expect hold path2: %v", holdPath2, expectHoldPath2)
	}

	// robot3 占点
	robot3id := 3
	mission3id := 103
	path3 := []string{"p52", "p9", "p8", "p7", "p41", "p42", "p43", "p44"}
	holdPath3, err := holdPathT(robot3id, mission3id, path3, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot3id, err)
	}

	expectHoldPath3 := []string{"p52"}
	success3 := assertHoldPathT(holdPath3, expectHoldPath3)
	if !success3 {
		t.Errorf("hold path3: %v , but expect hold path3: %v", holdPath3, expectHoldPath3)
	}

	// robot2 占点
	mission4id := 104
	path4 := []string{"p4", "p5", "p6", "p7", "p8", "p9", "p10", "p11", "p12", "p13"}
	holdPath4, err := holdPathT(robot2id, mission4id, path4, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot2id, err)
	}

	expectHoldPath4 := []string{"p4", "p5", "p6"}
	success4 := assertHoldPathT(holdPath4, expectHoldPath4)
	if !success4 {
		t.Errorf("hold path2: %v , but expect hold path2: %v", holdPath4, expectHoldPath4)
	}
}

/*
*
5.3. 点位占用: 双向路径上有车通过的情况

robot1：  p62 -> p7 -> p41 -> p42 -> p43 -> p44
expect1:  p62 -> p7 -> p41 -> p42 -> p43 -> p44
robot2:   p3 -> p4 -> p5 -> p6 -> p7 -> p8 -> p9 -> p10 -> p11 -> p12 - p13
expect2:  p3 -> p4 -> p5 -> p6
robot3:   p52 -> p9 -> p8 -> p7 -> p41 -> p42 -> p43 -> p44
expect3:  p52 -> p9 -> p8
robot2:   p4 -> p5 -> p6 -> p7 -> p8 -> p9 -> p10 -> p11 -> p12 - p13
expect2:  p4 -> p5 -> p6

全路径配置：p41 -> p44
*/
func TestSpotOccupy_MixtureAll_RobotPassAcrossSS(t *testing.T) {
	r := initialRedisT()
	defer r.Close()

	c := r.Get()
	c.Do("flushdb")
	defer c.Close()

	// 设置全路径
	first := []string{"p41"}
	last := []string{"p44"}
	r.ResetSet("allhold.end.f.1", first)
	r.ResetSet("allhold.end.l.1", last)

	// robot1 占点
	robot1id := 1
	mission1id := 101
	path1 := []string{"p62", "p7", "p41", "p42", "p43", "p44"}
	holdPath1, err := holdPathT(robot1id, mission1id, path1, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot1id, err)
	}

	expectHoldPath1 := []string{"p62", "p7", "p41", "p42", "p43", "p44"}
	success1 := assertHoldPathT(holdPath1, expectHoldPath1)
	if !success1 {
		t.Errorf("hold path1: %v , but expect hold path1: %v", holdPath1, expectHoldPath1)
	}

	// robot2 占点
	robot2id := 2
	mission2id := 102
	path2 := []string{"p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10", "p11", "p12", "p13"}
	holdPath2, err := holdPathT(robot2id, mission2id, path2, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot2id, err)
	}

	expectHoldPath2 := []string{"p3", "p4", "p5", "p6"}
	success2 := assertHoldPathT(holdPath2, expectHoldPath2)
	if !success2 {
		t.Errorf("hold path2: %v , but expect hold path2: %v", holdPath2, expectHoldPath2)
	}

	// robot3 占点
	robot3id := 3
	mission3id := 103
	path3 := []string{"p52", "p9", "p8", "p7", "p41", "p42", "p43", "p44"}
	holdPath3, err := holdPathT(robot3id, mission3id, path3, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot3id, err)
	}

	expectHoldPath3 := []string{"p52", "p9", "p8"}
	success3 := assertHoldPathT(holdPath3, expectHoldPath3)
	if !success3 {
		t.Errorf("hold path3: %v , but expect hold path3: %v", holdPath3, expectHoldPath3)
	}

	// robot2 占点
	mission4id := 104
	path4 := []string{"p4", "p5", "p6", "p7", "p8", "p9", "p10", "p11", "p12", "p13"}
	holdPath4, err := holdPathT(robot2id, mission4id, path4, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot2id, err)
	}

	expectHoldPath4 := []string{"p4", "p5", "p6"}
	success4 := assertHoldPathT(holdPath4, expectHoldPath4)
	if !success4 {
		t.Errorf("hold path2: %v , but expect hold path2: %v", holdPath4, expectHoldPath4)
	}
}

/*
*
5.4. 点位占用: 双向路径与全路径重合的情况

robot1：  p21 -> p22 -> p11 -> p10 -> p9 -> p8 -> p7 → p41 -> p42 -> p43 -> p44
expect1 : p21 -> p22 -> p11 -> p10
robot2:   p33 -> p5 -> p6 -> p7 -> p8 -> p9 -> p10 -> p11 -> p12 - p13
expect2 : p33

全路径配置：p5 -> p13
*/
func TestSpotOccupy_MixtureAll_SsPathOverlapAllPath(t *testing.T) {
	r := initialRedisT()
	defer r.Close()

	c := r.Get()
	c.Do("flushdb")
	defer c.Close()

	// 设置全路径
	first := []string{"p5"}
	last := []string{"p13"}
	r.ResetSet("allhold.end.f.1", first)
	r.ResetSet("allhold.end.l.1", last)

	// robot1 占点
	robot1id := 1
	mission1id := 101
	path1 := []string{"p21", "p22", "p11", "p10", "p9", "p8", "p7", "p41", "p42", "p43", "p44"}
	holdPath1, err := holdPathT(robot1id, mission1id, path1, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot1id, err)
	}

	expectHoldPath1 := []string{"p21", "p22", "p11", "p10"}
	success1 := assertHoldPathT(holdPath1, expectHoldPath1)
	if !success1 {
		t.Errorf("hold path1: %v , but expect hold path1: %v", holdPath1, expectHoldPath1)
	}

	// robot2 占点
	robot2id := 2
	mission2id := 102
	path2 := []string{"p33", "p5", "p6", "p7", "p8", "p9", "p10", "p11", "p12", "p13"}
	holdPath2, err := holdPathT(robot2id, mission2id, path2, r)
	if nil != err {
		t.Errorf("the hold path's points of robot(%d) err is %v", robot2id, err)
	}

	expectHoldPath2 := []string{"p33"}
	success2 := assertHoldPathT(holdPath2, expectHoldPath2)
	if !success2 {
		t.Errorf("hold path2: %v , but expect hold path2: %v", holdPath2, expectHoldPath2)
	}
}

func assertHoldPathT(actualHoldPath, expectHoldPath []string) bool {
	if len(actualHoldPath) != len(expectHoldPath) {
		return false
	}

	for index, spot := range actualHoldPath {
		if expectHoldPath[index] != spot {
			return false
		}
	}

	return true
}

func holdPathT(robotid int, missionid int,
	paths []string, p *Pool) ([]string, error) {
	// req := &robpbf.RequestPathSpots{}
	req := &ztpbf.ReqPathSpots{}
	req.RobotId = strconv.Itoa(robotid)
	// req.Mapid = 0
	// req.Command = 1
	req.MissionId = strconv.Itoa(missionid)
	req.ExpireSeconds = 120
	req.PathSpotsLeft = append(req.PathSpotsLeft, paths...)
	if len(paths) < 4 {
		req.RequestSpots = paths
	} else {
		req.RequestSpots = paths[0:4]
	}
	req.AgvType = 1

	res, err := p.LockSpots(req)
	if nil != err {
		return nil, err
	}
	return res.RequestSpots, nil
}

func initialRedisT() *Pool {
	pool := &Pool{
		Pool: zredis.NewPool("127.0.0.1:6379"),
	}
	pool.SetMaxActive(10)
	pool.SetMaxIdle(10)
	pool.SetIdleTime(120)
	return pool
	// rNode := &zkredis.RedisNode{}
	// rNode.Address = "127.0.0.1:6379"
	// rNode.MaxActive = 10
	// rNode.Password = "22222"
	// rNode.IdleTimeout = 120
	// rNode.SlaveAddress = "127.0.0.1:6379"
	// rls, _ := New(rNode)
	// return rls
}

func TestTmpHoldPathsT(t *testing.T) {
	// r := initialRedisT()
	// defer r.Close()
	//
	// c := r.Get()
	// c.Do("flushdb")
	// defer c.Close()

	// 临时测试用例

}
