# 资源相关的keys
## SET
* resources.needby.[spotname];//values:resource'ids; [spotname]这个点关联的资源
* resources.needbyend.[spotname];//values: resource'ids; 目的地占用相关的资源，就是一条移动指令的目的地关联了这个资源，需要先占用这个资源才去占点位行走；
* resources.[id];//values: robotid's; [id]是resource'id
* resources.askby.[robotid];//values: resource'ids, 正确被[robotid]请求的资源id；
* allhold.end.f.[id];//values:全路径占用的起点spotname；[id] 标识不同全路径的id;{全路径占用就是这些路径要一次性占用，否则全部退回，但上次已经占到的点除外}
* allhold.end.l.[id];//values:全路径占用的终点spotname
* sopts.[spotname];//values: robotid,带'EX'表示这个点位被谁占用到了
* sopt.near.[spot];//values:spotnames;[spot]有邻近点的点位
* noparking.spots;//values:spotnames; 非停靠点，就是这些点不能作为点位占用的最后一个点；
* req.path.[robotid];//
* robot.path.[robotid];//
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


# LockSpots 说明
如果不是全路径都占用，需要做如下处理(全路径要全部占用，只要一个点不能占用就要回退，所以不需要下面这么负责的处理)：
- 找出路径中需要全路径占用的点的index：[holdalls, holdalle] 
- 跟其它车辆的路径做逆向检查，如果有逆向路径干涉，找出不涉及逆向路径干涉的最后这个的index[max_allow_point_for_ss]（双向路径上已经被逆向车占了，逆向路径才成立），以及被双向路径阻挡得车辆；
- 按当前小占用点数[minHolds]继续向前扩展不能停车点(如果最后一个点是禁停点，则minHolds+1这么继续扩展下去)
- 如果[minHolds]比[max_allow_point_for_ss]大了，则需要用[max_allow_point_for_ss]向后跳过禁停点(如果最后一个点是禁停点，则-1，继续判断)
- 这样找出来的[holds] 这个时候和 [holdalls, holdalle] 相交了，则需要取并集
- 若[holdaalle] 比 [max_allow_point_for_ss] 大了，那么说明全路径不可占用，[holds] 取[holdalls-1]
- 不然 [holdalle] > [holds] 的时候，[holds] 取 [holdalle] 并且继续禁停点检测和扩展，若这样扩展后[holds]又比[max_allow_point_for_ss]大了或者最后一个还是禁停点，又只能全路径不占用，[holds]取[holdalls-1]了；
- 判断目的地资源是否已占到，没占到[holds]就只能是1; **这条可以移到最开头，可以提高效率**
- 找出了最后这次要占的holds，可以开始实际占点位了
- 下面是回滚
- 回滚的时候需要考虑全路径占用，如果最后一个点在全路径的前后中间，则需要整个全路径回退；双向路径逆干涉不用考虑，因为计算占用点数的时候已经考虑了
- 占用成功后需要删除前一次已占用，但这次不需要占用的点


# 可能存在的问题；
1. 回退的时候用Holds=1的问题？如果车辆处在两个点的中间，理论上至少需要占用前后两个点，这个时候如果holds改成了1，则有可能把上次占用到的第二个点给释放了或者过期时间没有延期；
1. 