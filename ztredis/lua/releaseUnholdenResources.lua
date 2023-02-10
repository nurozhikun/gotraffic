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