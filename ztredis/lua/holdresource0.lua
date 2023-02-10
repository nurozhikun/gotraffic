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
			if not bFind then--[[不在新请求的列表中]]
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
