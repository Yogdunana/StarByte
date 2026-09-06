package cache

import "github.com/redis/go-redis/v9"

const parseTokenLua = `
local function parse_token(raw)
  local idx = nil
  for i = #raw, 1, -1 do
    if string.sub(raw, i, i) == ":" then
      idx = i
      break
    end
  end
  if not idx then
    return raw, 1
  end
  local n = tonumber(string.sub(raw, idx + 1))
  if n == nil or n < 1 then
    n = 1
  end
  return string.sub(raw, 1, idx - 1), n
end
`

var acquireScript = redis.NewScript(parseTokenLua + `
local cur = redis.call("GET", KEYS[1])
if cur then
  local owner, n = parse_token(cur)
  if owner == ARGV[1] then
    n = n + 1
    redis.call("SET", KEYS[1], ARGV[1] .. ":" .. tostring(n), "PX", ARGV[2])
    return n
  end
  return 0
end
redis.call("SET", KEYS[1], ARGV[1] .. ":1", "PX", ARGV[2])
return 1
`)

var reenterScript = redis.NewScript(parseTokenLua + `
local cur = redis.call("GET", KEYS[1])
if not cur then
  return 0
end
local owner, n = parse_token(cur)
if owner ~= ARGV[1] then
  return 0
end
n = n + 1
redis.call("SET", KEYS[1], owner .. ":" .. tostring(n), "PX", ARGV[2])
return n
`)

var extendScript = redis.NewScript(parseTokenLua + `
local cur = redis.call("GET", KEYS[1])
if not cur then
  return 0
end
local owner = parse_token(cur)
if owner == ARGV[1] then
  return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0
`)

var unlockOnceScript = redis.NewScript(parseTokenLua + `
local cur = redis.call("GET", KEYS[1])
if not cur then
  return 0
end
local owner, n = parse_token(cur)
if owner ~= ARGV[1] then
  return 0
end
if n <= 1 then
  redis.call("DEL", KEYS[1])
  return -1
end
n = n - 1
redis.call("SET", KEYS[1], owner .. ":" .. tostring(n), "PX", ARGV[2])
return n
`)

var reclaimHeadScript = redis.NewScript(`
if redis.call("LINDEX", KEYS[1], 0) == ARGV[1] then
  redis.call("LPOP", KEYS[1])
  return 1
end
return 0
`)
