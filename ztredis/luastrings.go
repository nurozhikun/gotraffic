package ztredis

const (
	//iAgv和二维码都适用的路径占用方法
	//KEYS 是传入的请求路径（完整的剩余路径）
	//ARGV[1] = robotid,
	//ARGV[2] = expire_time_second,
	//ARGV[3] = min_holds_spots_len(0获取所有剩余路径点),
	//ARGV[4] = missionid,
	//ARGV[5] = 1表示需要做逆向路径检查, 0表示不用做检查，可以直接删除路径了；
	//ARGV[6] = 0 默认，1 是iAGV占用，2 二维码占用
	LuaLockSpots = `
	--[[全局变量，函数中可直接调用]]
	local stopbyrobot
	--[[函数:找出占用的路径目的地关联哪些资源]]
	local func_check_pathendresource_asked = function(endspotname)
		local asked = {}
		local idx = 1
		local reses = redis.call('SMEMBERS', 'resources.needby.'..endspotname)
		for i = 1, #reses do
			if redis.call('SISMEMBER', 'resources.needbyend', reses[i]) == 1 then
				asked[idx] = reses[i]
				idx = idx + 1
			end
		end
		return asked
	end
	--[[函数:判断这些资源是否都占用到了]]
	local func_check_resources_holden = function(asked, robotid)
		for i = 1, #asked do
			if redis.call('SISMEMBER', 'resource.'..asked[i], robotid) ~= 1 then
				return false
			end
		end
		return true
	end
	--[[函数：判断路径是否有一段需要一次性全部占用]]
	local func_allholds_check_end = function(ks)
		local ekeys = redis.call('KEYS', 'allhold.end.l.*')
		for i = 1,#ekeys do
			if redis.call('SISMEMBER', ekeys[i], ks[#ks]) == 1 then
				--[[若起始点在全路径终点集合中的话，则不全路径占用]]
				if redis.call('SISMEMBER', ekeys[i], ks[1]) == 1 then
					return 0, 0
				end
				local endkey = 'allhold.end.f.'..string.sub(ekeys[i], 15)
				for j = 1, #ks-1 do
					if redis.call('SISMEMBER', endkey, ks[j]) == 1 then
						return j, #ks --[[找到了，返回当前点到终点]]
					end
				end
				return 1, #ks --[[没找到，返回起点到终点]]
			end
		end
		--[[如果没找到，才使用中间点的规则]]
		return 0, 0
	end
	--[[函数：判断出需要全路径占用的起始、终止index]]
	local func_allholds_check = function(ks)
		local i1, i2 = func_allholds_check_end(ks)
		if i2 > i1 then
			return i1, i2
		end
		return i2, i1
	end
	--[[函数：判断两个字符串是否逆向相交,返回相交部分在第一个字符串处的下标]]
	local func_ss_check = function(s1, s2)
		if #s1 < 2 or #s2 < 2 then
			return 0, 0
		end
		local iStart = 0
		local iEnd = 0
		for i = 1, #s1 do
			for j = #s2, 1, -1 do
				if s1[i] == s2[j] then
					iStart = i
					iEnd = i
					local count = math.min(#s1-i, j-1)
					for n = 1, count do
						if s1[i+n] == s2[j-n] then
							iEnd = i+n
						else
							break
						end
					end
					break
				end
			end
			if iStart ~= iEnd then --[[这两个不等，才有两个相交，如果相等只有一个点相交，一个点相交不是双向路径逆重叠]]
				break
			end
		end
		return iStart, iEnd
	end
	--[[函数：spots.* key 的点位占用, return: 0没占成功；1新占到的点；2旧占到的点]]
	local func_lock_spot = function(spot, robotid, expire_time)
		local key = 'spots.'..spot
		if not redis.call('SET', key, robotid, 'NX','EX', expire_time) then
			local v = redis.call('GET',key)
			if v == robotid then
				redis.call('SET', key, robotid,'EX', expire_time)
				return 2
			else
				stopbyrobot = v
				redis.call('HMSET', 'req.res.'..robotid, 'stopbyrobot', v, 'stopbyspot', spot) --[[存点位占用失败的时候，被阻挡的robotid到redis hash中]]
				redis.call('HSET', 'stopspots', robotid, spot)
				return 0
			end
		else
			return 1
		end
	end
	--[[函数：]]
	local func_unlock_spot = function(spotkey)
		redis.call('DEL', spotkey)
	end
	--[[函数：]]
	local func_unlock_newsetspot = function(tbl, bAll)
		for k, v in pairs(tbl) do
			if bAll or v == 1 then
				func_unlock_spot('spots.'..k)
			end
		end
	end
	--[[占用一个spot的函数，返回成功占用的键值和邻近点的键值]]
	--[[函数：iAgv车占用]]
	local func_hold_iagv_spot = function(spot, robotid, expire_time) --[[时间单位是秒s]]
		local ret = {}
		--[[占用自己]]
		local n = func_lock_spot(spot, robotid, expire_time)
		if n == 0 then --[[自己没有占用成功]]
			return nil
		end
		ret[spot] = n
		--[[占用邻近点]]
		local nears = redis.call('SMEMBERS', 'spot.near.'..spot)
		for i = 1, #nears do
			n = func_lock_spot(nears[i], robotid, expire_time)
			if n == 0 then --[[邻近点没占成功]]
				func_unlock_newsetspot(ret, true)
				return nil
			end
			ret[nears[i]] = n
		end
		return ret
	end
	--[[函数：二维码车占用]]
	local func_hold_sagv_spot = function(spot, robotid, expire_time)
		local ret = {}
		if redis.call('EXISTS', 'rotate.spots.'..spot) == 1 then
			local robs = redis.call('SMEMBERS','rotate.spots.'..spot)
			if #robs > 1 or robs[1] ~= robotid then
				return nil
			end
		end
		local n = func_lock_spot(spot, robotid, expire_time)
		if n == 0 then
			return nil
		end
		ret[spot] = n
		return ret
	end
	--[[函数：获取robot path的长度]]
	local func_get_robot_path_len = function(robotid)
		local ret = {}
		local paths = redis.call('KEYS', 'robot.path.*')
		for i = 1, #paths do
			local srobid = string.sub(paths[i], 12)
			if srobid == robotid then
				local otherPath = redis.call('LRANGE', paths[i], 0, -1)
				ret =  #otherPath
				break
			end
		end
		return ret
	end
	--[[获得点在路径中的位置，返回0表示点不在路径中]]
	local func_spot_index_of_path = function(path, spot)
		local spot_index = 0

		for i = 1, #path do
			if path[i] == spot then
				spot_index = i
			end
		end

		return spot_index
	end
	--[[根据邻近点扩展双向路径]]
	local func_extend_ss_seg = function(path, ss_start, ss_end)
		local first_index, last_index
		first_index = ss_start
		last_index = ss_end

		local path_seg_mark = {}
		for i = 1, #path do
			path_seg_mark[i] = 0
		end

		for i = 1, #path do
			local spot = path[i]
			local nears = redis.call('SMEMBERS', 'spot.near.'..spot)
			for j = 1, #nears do
				local near_spot_index
				near_spot_index = func_spot_index_of_path(path, nears[j])
				-- [[若邻近点不在路径中，则跳过]]
				if near_spot_index ~= 0 then
					-- [[设置path_seg_mark]]
					local path_spot_index = i
					local near_seg_start = path_spot_index
					local near_seg_end = near_spot_index
					if path_spot_index > near_spot_index then
						near_seg_start = near_spot_index
						near_seg_end = path_spot_index
					end

					for k = near_seg_start, near_seg_end do
						path_seg_mark[k] = 1
					end
				end
			end
		end

		--[[双向路径向前扩充]]
		if path_seg_mark[first_index] == 1 then
			for i = first_index, 1, -1 do
				if path_seg_mark[i] ~= 1 then
					break
				end
				first_index = i
			end
		end

		--[[双向路径向后扩充]]
		if path_seg_mark[last_index] == 1 then
			for i = last_index, #path do
				if path_seg_mark[i] ~= 1 then
					break
				end
				last_index = i
			end
		end

		return first_index, last_index
	end
	--[[判断扩展后的双向路径上有没有目标车辆，若有则此双向路径有效]]
	local func_is_ss_valid  = function(path, first_index, last_index, orobid, srobid)
		local ss_is_valid = 0
		local obstacle_robot_id

		for index = first_index, last_index do
			local robotid = redis.call('GET','spots.'..path[index])
			--[[若占用路径点小车的路径只有一个点，那么需要设置障碍]]
			local pathLen = nil
			pathLen = func_get_robot_path_len(robotid)
			if pathLen ~= nil and pathLen == 1 then --[[若有车停在双向路径上]]
				redis.call('HMSET', 'req.res.'..orobid, 'stopbyrobot', robotid, 'stopbyspot', path[index]) --[[存点位占用失败的时候，被阻挡的robotid到redis hash中]]
				redis.call('HSET', 'stopspots', orobid, path[index])
				ss_is_valid = 1
				obstacle_robot_id = robotid
				break
			end
			--[[判断点位是否被目标robot占用]]
			if robotid == srobid then
				ss_is_valid = 1
				obstacle_robot_id = srobid
				break
			end
		end
		return ss_is_valid, obstacle_robot_id
	end
	--[[找到可停靠点]]
	local func_find_parking_spot = function(path, start_id, end_id)
		local ret = start_id

		if start_id > end_id then
			for i = start_id,end_id,-1 do
				ret = i
				if redis.call('SISMEMBER', 'noparking.spots', path[i]) ~= 1 then
					break
				end
			end
		else
			for i = start_id,end_id do
				ret = i
				if redis.call('SISMEMBER', 'noparking.spots', path[i]) ~= 1 then
					break
				end
			end
		end

		return ret
	end
	--[[****************主函数的开始位置************************************************************************]]
	--[[检查判断出应该一次获取几个点比较合适]]
	local inRobotId = ARGV[1]
	local minHolds = tonumber(ARGV[3])--[[0表示整个路径都占用到]]
	local expireTime = ARGV[2]
	local holdalls = 0
	local holdalle = 0
	local holds
	local delToIdx = 1 --[[这个Idx本身不回滚删除,至少要占着一个点]]
	local hasAllHoldPath = false
	--[[保存车辆路径]]
	redis.call('DEL', 'robot.path.'..inRobotId) --[[先删除小车自己的path]]
	if #KEYS > 0 then
		redis.call('RPUSH', 'robot.path.'..inRobotId, unpack(KEYS))
	end
	--[[redis.call('EXPIRE', 'robot.path.'..inRobotId, expireTime)]]
	if minHolds ~= 0 then --[[不是全路径占用，判断是否有一段是需要全路径占用的]]
		holdalls, holdalle = func_allholds_check(KEYS)
		if holdalle > holdalls then
			hasAllHoldPath = true --[[判断出需要有全路径占用的点]]
		end

		--[[最大可达无双向路径逆向干涉的点]]
		local max_allow_point_for_ss = #KEYS
		--[[检测双向路径的逆向干涉]]
		redis.call('DEL', 'req.path.'..inRobotId) --[[先删除小车自己的path]]
		redis.call('HDEL', 'ssobstacle', inRobotId)
		if tonumber(ARGV[5]) == 1 then --[[如果指令要做逆向检查，并不是占有整个路径]]
			local paths = redis.call('KEYS', 'req.path.*')
			local ss_obstacle_robid
			for i = 1, #paths do
				local srobid = string.sub(paths[i], 10)
				local otherPath = redis.call('LRANGE', paths[i], 0, -1)
				local is, ie = func_ss_check(KEYS, otherPath)
				if ie > is then --[[ie是最后一个相等点，所以ie>is就可以了]]
					--[[根据邻近点信息扩展双向路径]]
					local first_index, last_index
					first_index, last_index = func_extend_ss_seg(KEYS, is, ie)
					--[[判断扩展后的双向路径上有没有目标车辆，若有则有效]]
					local ss_is_valid, obstacle_robot_id = func_is_ss_valid(KEYS, first_index, last_index, inRobotId, srobid)
					if ss_is_valid == 1 and first_index <= max_allow_point_for_ss then
						max_allow_point_for_ss = math.max(1,first_index-1)
						ss_obstacle_robid = obstacle_robot_id
					end
				end
			end
			--[[记录双向路径的阻碍]]
			if ss_obstacle_robid ~= nil then
				redis.call('HSET', 'ssobstacle', inRobotId, ss_obstacle_robid)
			end

			if #KEYS > 0 then
				redis.call('RPUSH', 'req.path.'..inRobotId, unpack(KEYS))
				redis.call('EXPIRE', 'req.path.'..inRobotId, expireTime)
			end
		end
		--[[根据传入的最小占用路径点点数，依情况进行扩展]]
		holds = minHolds
		--[[如果当前点是禁止停车点，最后一个点位要向路径的后面推]]
		holds = func_find_parking_spot(KEYS, holds, #KEYS)
		--[[若已经向后推了,且推多了，则向前推]]
		if holds > max_allow_point_for_ss then
			holds = func_find_parking_spot(KEYS, max_allow_point_for_ss, 1)
		end
		--[[若与全路径段有交集，则取并集]]
		if holdalls < holdalle and holdalls <= holds then
			--[[若不涉及双向路径]]
			if holdalle <= max_allow_point_for_ss then
				local former_holds = holds
				holds = math.max(holds, holdalle)
				--[[若全路径的最后一个点为禁止停靠点的处理]]
				if holds > former_holds then
					--[[如果当前点是禁止停车点，最后一个点位要向路径的后面推]]
					holds = func_find_parking_spot(KEYS, holds, #KEYS)
					--[[若后推过的点过多超过允许的范围，或者后推的最后一个点依然为非停靠点，则需回退到iStart之前]]
					if holds > max_allow_point_for_ss or redis.call('SISMEMBER', 'noparking.spots', KEYS[holds]) == 1 then
						holds = math.max(1,holdalls-1)
						--[[继续排除非停靠点]]
						holds = func_find_parking_spot(KEYS, holds, 1)
					end
				end
			else
				--[[若涉及双向路径，则全路径不可占用]]
				holds = math.max(1, holdalls-1)
				--[[向前寻找停靠点]]
				holds = func_find_parking_spot(KEYS, holds, 1)
			end
		end
	else
		holds = #KEYS --[[需要占用的最后一个点的下标(和个数是一样的)]]
	end
	--[[redis.call('HMSET', 'th.'..inRobotId, 'holds', holds, 'holdalls', holdalls, 'holdalle', holdalle)]]
	--[[若Path终点有关联的资源，判断是否有资源没有占用到，如果没占用到holds直接改成1表示]]
	local askedByEnd = {}
	if #KEYS > 0 then
		askedByEnd = func_check_pathendresource_asked(KEYS[#KEYS])
		if #askedByEnd > 0 then
			if not func_check_resources_holden(askedByEnd, inRobotId) then --[[目的地资源没占用成功，则只占用车辆当前点]]
				holds = 1
			end
		end
	end
	--[[真正开始占用KEYS中的点位了]]
	local spotsLocked = {} --[[字典数组，每个值是又是一个字典]]
	local iFailed = 0
	for i = 1, holds do
		local ret = nil
		if tonumber(ARGV[6]) == 2 then
			ret = func_hold_sagv_spot(KEYS[i], inRobotId, expireTime)
		else
			ret = func_hold_iagv_spot(KEYS[i], inRobotId, expireTime)
		end
		if not ret then
			iFailed = i
			break
		end
		spotsLocked[i] = ret
	end
	local idxLast = #spotsLocked

	--[[--debug]]

	--[[该占的都占到了，删除stop数据]]
	delToIdx = idxLast
	if iFailed == 0 then --[[全部占用成功了，不用回退]]
		redis.call('HDEL', 'req.res.'..inRobotId, 'stopbyrobot', 'stopbyspot', 'stopbyresource')
		redis.call('HDEL', 'stopspots', inRobotId)
	else --[[有点没占成功并且需要回滚的，回滚没占成功的点位"//and delToIdx > 0 and delToIdx < #spotsLocked"]]
		if minHolds == 0 then --[[需要全路径占用的，直接回退到1]]
			delToIdx = 1
		else --[[不是全路径占用的，计算具体的回退点]]
			if delToIdx >= holdalls and delToIdx <= holdalle then
				delToIdx = math.max(1, holdalls-1)
			end
		end
	end
	--[[redis.call('HMSET', 'debug.'..inRobotId, 'delToIdx', delToIdx, 'holds', holds)]]
	--[[检查路径上的资源是否都占到了]]
	local iFailedToHoldResource = 0
	redis.call('DEL', 'resources.askby.'..inRobotId)
	--[[添加目的地相关的资源]]
	if #askedByEnd > 0 then
		redis.call('SADD', 'resources.askby.'..inRobotId, unpack(askedByEnd))
	end
	local resKeys = {}
	for i = 1, delToIdx do
		resKeys[i] = 'resources.needby.'..KEYS[i]
		if iFailedToHoldResource == 0 then
			local resCodes = redis.call('SMEMBERS', resKeys[i])
			for j = 1, #resCodes do
				if redis.call('SISMEMBER', 'resource.'..resCodes[j], inRobotId) ~= 1 then
					iFailedToHoldResource = i
					redis.call('HMSET', 'req.res.'..inRobotId, 'stopbyrobot', 0, 'stopbyspot', KEYS[i], 'stopbyresource', resCodes[j])
					redis.call('HSET', 'stopspots', inRobotId, resKeys[i])
					break
				end
			end
		end
	end
	if #resKeys > 0 then
		redis.call('SUNIONSTORE', 'resources.askby.'..inRobotId, 'resources.askby.'..inRobotId, unpack(resKeys))
		redis.call('EXPIRE', 'resources.askby.'..inRobotId, expireTime)
	end
	if iFailedToHoldResource > 0 then
		delToIdx = math.max(1, iFailedToHoldResource - 1)
	end
	--[[真的要回滚占用到的点了]]
	if delToIdx < #spotsLocked then
		for i = delToIdx+1, #spotsLocked do
			func_unlock_newsetspot(spotsLocked[i], true)
		end
		--[[重新占自身所在点，以免之前回滚的时候被清除]]
		local ret = nil
		if tonumber(ARGV[6]) == 2 then
			ret = func_hold_sagv_spot(KEYS[1], inRobotId, expireTime)
		else
			ret = func_hold_iagv_spot(KEYS[1], inRobotId, expireTime)
		end
		--[[回滚资源]]
		idxLast = delToIdx
	end
	--[[检查最后一个点位是不是禁止停车点，如果是禁止停车点，需要回退到可以停车的地方为止]]
	for i = idxLast, 1, -1 do
		local isnoparking = false
		for k, v in pairs(spotsLocked[i]) do
			if redis.call('SISMEMBER', 'noparking.spots', k) == 1 then
				isnoparking = true
				idxLast = i - 1
				break
			end
		end
		if not isnoparking then
			break
		end
	end
	--[[获取占用到的点位，asked的resource]]
	local allkeys = {}
	local idx = 1
	local allkeystag = {}
	for i = 1, idxLast do
		for k, v in pairs(spotsLocked[i]) do
			if not allkeystag[k] then
				allkeystag[k] = k
				allkeys[idx] = k
				idx = idx + 1
			end
		end
	end
	--[[保存占用到的点位到redis列表中]]
	redis.call('DEL', 'hold.by.'..inRobotId)
	if #allkeys > 0 then
		redis.call('RPUSH', 'hold.by.'..inRobotId, unpack(allkeys))
		redis.call('EXPIRE', 'hold.by.'..inRobotId, expireTime)
	end
	--[[清除其它无用的占用点位，如果有交管点释放，设置为0]]
	local ekeys = redis.call('KEYS','spots.*')
	local flag = nil
	for i = 1, #ekeys do
		local name = string.sub(ekeys[i], 7)
		if not allkeystag[name] then
			if redis.call('GET', ekeys[i])==inRobotId then
				func_unlock_spot(ekeys[i])
			end
		end
	end

	--[[取得真正占用到的路径点作为返回值]]
	allkeys = {}
	for i = 1, idxLast do
		allkeys[i] = KEYS[i]
	end

	--[[记录req.res]]
	redis.call('HMSET', 'req.res.'..inRobotId, 'missionid', ARGV[4])
	redis.call('EXPIRE', 'req.res.'..inRobotId, expireTime)
	return allkeys
	`
	LuaHoldResource = `
	local ret = {}
	local idx = 1
	local holdens
	local key = 'resource.'..ARGV[1]
	local maxnum = tonumber(ARGV[2])
	local adds = {}
	local ai = 1
	if #KEYS < 1 then
		redis.call('DEL', key)
	else
		--[[查找并删除对该资源不占用的车辆]]
		holdens = redis.call('SMEMBERS', key)
		for i = 1, #holdens do
			local bFind = false
			for j = 1, #KEYS do
				if holdens[i] == KEYS[j] then
					bFind = true
					break
				end
			end
			if not bFind then
				ret[idx] = holdens[i]
				idx = idx + 1
			end
		end
		if idx > 1 then
			redis.call('SREM', key, unpack(ret))
		end
		ret = {}
		idx = 1
		for i = 1, #KEYS do
			if redis.call('SISMEMBER', key, KEYS[i]) == 1 then
				ret[idx] = KEYS[i] --[[已经占用到该资源的车辆ID]]
				idx = idx + 1
			else
				adds[ai] = KEYS[i] --[[还没有占用到该资源的车辆ID]]
				ai = ai + 1
			end
		end
		--[[test...]]
		--[[redis.call('HMSET', 'ddbg', 'idx', idx, 'ai', ai, 'addslen', #adds)]]
		if idx <= maxnum and ai > 1 then
			local len = math.min(maxnum-idx+1, #adds)
			local appends = {}
			for i = 1, len do
				appends[i] = adds[i]
				ret[idx] = adds[i]
				idx = idx + 1
			end
			redis.call('SADD', key, unpack(appends))
		end
	end
	return ret
	`
	//释放占用的点位，当释放的是个交通管制点的时候，需要把它占用住
	LuaHoldResourceuaUnlockKeysExceptAndResetTraffic = `
        local ret = 0
        local spotPrefix = "spots."
        local ekeys = redis.call('KEYS',"spots.*")
		local flag = nil
		for i = 1, #ekeys do
			if redis.call('get',ekeys[i])==ARGV[1] then
				for j = 1, #KEYS do
					if ekeys[i] == KEYS[j] then
						flag = 1
					end
				end
				if flag == nil then
					redis.call('del',ekeys[i])
				end
				flag = nil
			end
		end

		local spotBorders = redis.call('SMEMBERS', 'traffic.spots')
		for k, v in pairs(spotBorders) do
			if redis.call('EXISTS', spotPrefix..v) == 0 then
				redis.call('SET', spotPrefix..v, 0)
				ret = ret + 1
			end
		end
		return ret
		`
	//释放占用的点位，当释放的是个交通管制点的时候，需要把它占用住
	LuaUnlockKeysExceptAndResetTraffic = `
        local ret = 0
        local spotPrefix = "spots."
        local ekeys = redis.call('KEYS',"spots.*")
		local flag = nil
		for i = 1, #ekeys do
			if redis.call('get',ekeys[i])==ARGV[1] then
				for j = 1, #KEYS do
					if ekeys[i] == KEYS[j] then
						flag = 1
					end
				end
				if flag == nil then
					redis.call('del',ekeys[i])
				end
				flag = nil
			end
		end

		local spotBorders = redis.call('SMEMBERS', 'traffic.spots')
		for k, v in pairs(spotBorders) do
			if redis.call('EXISTS', spotPrefix..v) == 0 then
				redis.call('SET', spotPrefix..v, 0)
				ret = ret + 1
			end
		end
		return ret
		`
	//这个方法是给二维码地图使用的，所以没有邻近点的概念
	//设置一个旋转点边上的4个旋转时需要占用的点，这些点存在rotate.borders.[spotname]
	//KEYS 是路径点旋转点，一般是一个，ARGV[1]是占用的value, ARGV[2]是过期时间
	LuaLockAttachSpotsOfRotates = `
		local ret = 1
		local spotPrefix = 'spots.'
		local rotatePrefix = 'rotate.spots.'
		local borderPrefix = 'rotate.borders.'
		local addedKeys = {}
		local bFailed = false
		for i = 1, #KEYS do
			local spotBorders = redis.call('SMEMBERS', borderPrefix..KEYS[i])
			for k, v in pairs(spotBorders) do
				if redis.call('EXISTS', spotPrefix..v) == 1 then
				  local spotCarID = redis.call('GET', spotPrefix..v)
					if spotCarID ~= ARGV[1] then
					 	bFailed = true
						break
					end
				else
					if not redis.call('SADD', rotatePrefix..v, ARGV[1]) or not redis.call('EXPIRE', rotatePrefix..v, ARGV[2]) then
						bFailed = true
						break
					end
				end
				addedKeys[ret] = v
				ret = ret + 1
			end
			if bFailed then
				break
			end
		end
		if bFailed then
			for k, v in pairs(addedKeys) do
				redis.call('SREM', rotatePrefix..v, ARGV[1])
			end
			ret = 0
		end
		return ret
	`
	//这个方法是给二维码地图使用的，所以没有邻近点的概念
	//释放一个车辆占用的旋转点，这些点存在rotate.borders.[spotname]
	//ARGV[1]是占用的value
	LuaUnlockRotateAttachSpots = `
		local ret = 0
		local rotatePrefix = 'rotate.spots.'
        local rotateKeys = redis.call('KEYS', 'rotate.spots.*')
        for k, v in pairs(rotateKeys) do
            redis.call('SREM', v, ARGV[1])
            ret = ret + 1
        end
		return ret
	`
	LuaReadHoldenResources = `
	local resources = {}
	local ret = {}
	local idx = 1
	local holdenkeys = redis.call('KEYS', 'resource.*')
	for i = 1, #holdenkeys do
		local resid = string.sub(holdenkeys[i], 10, -1)
		local robids = redis.call('smembers', holdenkeys[i])
		if #robids > 0 then
			for j = 1,#robids do
				if not resources[resid] then
					resources[resid] = robids[j]
				else
					resources[resid] = resources[resid]..','..robids[j]
				end
			end
		end
	end

	for k,v in pairs(resources) do
		ret[idx] = k
		idx = idx + 1
		ret[idx] = v
		idx = idx + 1
	end

	return ret
	`
	LuaReadAskedResources = `
	local resources = {}
	local ret = {}
	local idx = 1
	--[[读取已经占用的点]]
	local askedkeys = redis.call('KEYS', 'resources.askby.*')
	for i = 1, #askedkeys do
		--[[获取占用AGVID]]
		local robotid = string.sub(askedkeys[i], 17, -1)
		--[[local rs = redis.call('sdiff', askedkeys[i], 'resources.holdby.'..robotid)]]
		local rs = redis.call('smembers', askedkeys[i])
		if #rs > 0 then
			for j = 1, #rs do
				if not resources[rs[j]] then
					resources[rs[j]] = robotid
				else
					resources[rs[j]] = resources[rs[j]]..','..robotid
				end
			end
		end
	end
	for k, v in pairs(resources) do
		ret[idx] = k
		idx = idx + 1
		ret[idx] = v
		idx = idx + 1
	end
	return ret
	`
	LuaLockKeys = `
	local ret = {}
	local idx = 1
    for i = 1, #KEYS do
        if not redis.call('SET', KEYS[i], ARGV[1], 'NX','EX', ARGV[2]) then
			if redis.call('GET',KEYS[i])~=ARGV[1] then
				break
			else
				redis.call('SET', KEYS[i], ARGV[1],'EX', ARGV[2])
			end
		end
		ret[idx] = KEYS[i]
		idx = idx + 1
    end
    return ret
	`
	// 找最短的申请占用这个点的机器人编号, 如果这个点刚被释放，需要把这个点设置成被0占用
	// -2 已经被占用
	// -1 释放
	// =0 没有请求
	// >0 请求
	LuaGetApplySpotRequest = `
        local current = redis.call('GET', 'spots.'..ARGV[1])
        if not current then
            redis.call('SET', 'spots.'..ARGV[1], 0)
            return -1
        elseif current ~= "0" then
            return -2
        end
        local tmp = redis.call('KEYS', 'req.spots.*')
		local ret = 0
		local idx = 999
        for i=1, #tmp do
            local rid = tonumber(string.sub(tmp[i], 11, -1))
            local spots = redis.call('LRANGE', 'req.spots.'..rid, 0, -1)
			for j=1, #spots do
				if spots[j] == ARGV[1] then
					if j < idx then
						ret = rid
						idx = j
					end
				end
			end
        end
		return ret
     `
	// 设置占用点
	// 需要这个点的当前占用是0
	LuaSetApplySpotResult = `
        local key = 'spots.'..ARGV[1]
        local current = redis.call('GET', key)
        if current ~= "0"  then
            return -1
        end
        redis.call('SET', key, ARGV[2])
        return 0
	`
	LuaFreeRobotInHoldenResourece = `
    local ret = {}
    local holdenResources = redis.call('KEYS', 'resource.*')
    for i = 1, #holdenResources do
        local code = string.sub(holdenResources[i], 10, -1)
        local removed = {}
        local idx = 1
        local robotids = redis.call('SMEMBERS', holdenResources[i])
        for j = 1, #robotids do
            if redis.call('SISMEMBER', 'resources.askby.'..robotids[j], code) == 1 then
                if not ret[code] then
                    ret[code] = robotids[j]
                else
                    ret[code] = ret[code]..','..robotids[j]
                end
            else
                removed[idx] = robotids[j]
            end
        end
        if #removed > 0 then
            redis.call('SREM', holdenResources[i], unpack(removed))
        end
    end
    local ret2 = {}
    local idx = 1
    for k, v in pairs(ret) do
		ret2[idx] = k
		idx = idx + 1
		ret2[idx] = v
		idx = idx + 1
	end
	return ret2
	`
	LuaGetCandidateOfResourceHolder = `
	local ret = {}
	local key = 'resource.'..ARGV[1]
	local ai = 1
	--[[redis.call('HMSET', 'ddbg', 'idx', idx, 'ai', ai, 'addslen', #adds)]]
	if #KEYS < 1 then
		redis.call('DEL', key)
	else
		for i = 1, #KEYS do
			if redis.call('SISMEMBER', key, KEYS[i]) ~= 1 then
				ret[ai] = KEYS[i] --[[还没有占用到该资源的车辆ID]]
				ai = ai + 1
			end
		end
	end
	return ret
	`
)
