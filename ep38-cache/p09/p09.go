package p09

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var casSet = redis.NewScript(`
local old = redis.call('HMGET', KEYS[1], 'v', 'ver')
if (not old[2]) or tonumber(ARGV[2]) >= tonumber(old[2]) then
  redis.call('HSET', KEYS[1], 'v', ARGV[1], 'ver', ARGV[2])
  redis.call('EXPIRE', KEYS[1], ARGV[3])
  return 1
end
return 0
`)

func Backfill(ctx context.Context, rdb *redis.Client, key, val string, ver int64, ttl time.Duration) (bool, error) {
	n, err := casSet.Run(ctx, rdb, []string{key}, val, ver, int(ttl.Seconds())).Int()
	return n == 1, err
}
