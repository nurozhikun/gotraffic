package redistraffic

import (
	"errors"
	"nj.gitlab.com/nj-tools/l4g"
	"njrobot/robpbf"
)

//NewLockHoldSpots ...
func (r *RdsTraffic) NewLockHoldSpots(reqHolds *robpbf.RequestPathSpots) error {
	reqHolds.Agvtype = reqHolds.Agvtype/10000 + 1 // 1表示iagv，2表示二维码
	res, err := r.LockSpots(reqHolds)
	if nil != err {
		return err
	}
	if len(res.ReqSpotNames) == 0 {
		return errors.New("Failed to hold one spot")
	}
	reqHolds.ReqSpotNames = res.ReqSpotNames
	//二维码车辆
	if reqHolds.Agvtype == 2 {
		//如果是旋转指令，需要占用旋转点
		if (reqHolds.Command == int32(robpbf.MissionCommand_ROTATE_ROBOT_90) ||
			reqHolds.Command == int32(robpbf.MissionCommand_ROTATE_ROBOT_180) ||
			reqHolds.Command == int32(robpbf.MissionCommand_ROTATE_ROBOT_270) ||
			reqHolds.Command == int32(robpbf.MissionCommand_ROTATE_SHELF_NORTH_SIDE_TO_NORTH) ||
			reqHolds.Command == int32(robpbf.MissionCommand_ROTATE_SHELF_EAST_SIDE_TO_NORTH) ||
			reqHolds.Command == int32(robpbf.MissionCommand_ROTATE_SHELF_WEST_SIDE_TO_NORTH) ||
			reqHolds.Command == int32(robpbf.MissionCommand_ROTATE_SHELF_SOUTH_SIDE_TO_NORTH)) &&
			0 == reqHolds.Finishtype {
			ret, err := r.LockAttachSpotsOfRotates(reqHolds.Robotid, 90, reqHolds.ReqSpotNames[0])
			l4g.Debug("@@--@@ rotate borders", reqHolds, ret, err)
			if nil != err || 0 == ret {
				reqHolds.ReqSpotNames = nil
			}
		} else {
			r.UnlockRotateAttachSpots(reqHolds.Robotid)
		}
	}
	return nil
}

func (r *RdsTraffic) ReqholdSpots(pbReq *robpbf.RequestPathSpots) (*robpbf.RequestPathSpots, error) {
	err := r.NewLockHoldSpots(pbReq)
	return pbReq, err
}

func (r *RdsTraffic) ReqholdSpotsTraffic(pbReq *robpbf.RequestPathSpots) (*robpbf.RequestPathSpots, error) {
	// 设置超时时间
	pbReq.Seconds = 90 // 获取占用点
	return r.ReqholdSpots(pbReq)
}
