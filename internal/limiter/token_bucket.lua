-- KEYS[1] = bucket key
-- ARGV[1] = capacity, ARGV[2] = refill tokens/sec, ARGV[3] = key TTL seconds
local t = redis.call("TIME")                      -- use Redis's clock, not the app's
local now = t[1] * 1000 + math.floor(t[2] / 1000) -- ms

local capacity = tonumber(ARGV[1])
local rate     = tonumber(ARGV[2])

local data   = redis.call("HMGET", KEYS[1], "tokens", "ts")
local tokens = tonumber(data[1])
local ts     = tonumber(data[2])

if tokens == nil then          -- first request: start with a full bucket
  tokens = capacity
  ts = now
end

local elapsed = math.max(0, now - ts) / 1000
tokens = math.min(capacity, tokens + elapsed * rate)

local allowed = 0
local retry_ms = 0
if tokens >= 1 then
  tokens = tokens - 1
  allowed = 1
else
  retry_ms = math.ceil((1 - tokens) / rate * 1000)
end

redis.call("HSET", KEYS[1], "tokens", tokens, "ts", now)
redis.call("EXPIRE", KEYS[1], tonumber(ARGV[3]))

return {allowed, math.floor(tokens), retry_ms}