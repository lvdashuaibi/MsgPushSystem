package constant

// LuaCheckAndExpireDistributionLock 判断是否拥有分布式锁的归属权，是则设置过期时间
const LUA_ZRANGEBYSCORE_AND_REM = `
-- 定义 ZSet 的键名
local key = KEYS[1]

-- 定义分数的最小值和最大值
local minScore = tonumber(ARGV[1])
local maxScore = tonumber(ARGV[2])

-- 使用 ZRANGEBYSCORE 获取分数范围内的元素
local elements = redis.call('ZRANGEBYSCORE', key, minScore, maxScore)

-- 遍历并删除这些元素
for i, elem in ipairs(elements) do
    -- 使用 ZREM 删除元素
    redis.call('ZREM', key, elem)
end

-- 返回删除的元素数量（如果有需要的话）
return elements
`
