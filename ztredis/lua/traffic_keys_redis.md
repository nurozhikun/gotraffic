# 资源相关的keys
## SET
* resources.needby.[spotname];//values:resource'ids; [spotname]这个点关联的资源
* resources.needbyend;//values: resource'ids; 目的地占用相关的资源，就是一条移动指令的目的地关联了这个资源，需要先占用这个资源才去占点位行走；
* resources.[id];//values: robotid's; [id]是resource'id
* resources.askby.[robotid];//values: resource'ids, 正确被[robotid]请求的资源id；
* allhold.end.f.[id];//values:全路径占用的起点spotname；[id] 标识不同全路径的id;{全路径占用就是这些路径要一次性占用，否则全部退回，但上次已经占到的点除外}
* allhold.end.l.[id];//values:全路径占用的终点spotname
* sopts.[spotname];//values: robotid,带'EX'表示这个点位被谁占用到了
* sopt.near.[spot];//values:spotnames;[spot]有邻近点的点位
* noparking.spots;//values:spotnames; 非停靠点，就是这些点不能作为点位占用的最后一个点；

## HASH
### req.res.[robotid];

* stopbyrobot;//value:robotid; [robotid]这台agv被哪辆车阻挡了
* stopbysopt;//value: spotname; [robotid]这台agv被哪个点阻挡了
* missionid;//value: missionid; 请求到点位的任务路径；

### ssobstacle
* robotid;//value:robotid; the robot of key is stopped by the robot of value as sspath(双向路径干涉)

## LIST
* robot.path.[robotid];//value: spotnames;机器人[robotid的路径
* hold.by.[robotid];//value: spotnames;被机器人[robotid]占用的点位
