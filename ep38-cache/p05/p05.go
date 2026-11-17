package p05

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type cachedVal struct {
	Data     string `json:"data"`
	ExpireAt int64  `json:"expire_at"`
}

type ProductCache struct {
	rdb *redis.Client
	sf  singleflight.Group
}

func (c *ProductCache) load(ctx context.Context, id string) (string, error) { return "", nil }
func (c *ProductCache) loadSync(ctx context.Context, key, id string) (string, error) {
	return c.load(ctx, id)
}

func (c *ProductCache) Get(ctx context.Context, id string) (string, error) {
	key := "product:" + id
	raw, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return c.loadSync(ctx, key, id)
	}
	if err != nil {
		return "", err
	}
	var cv cachedVal
	if err := json.Unmarshal(raw, &cv); err != nil {
		return "", err
	}
	if time.Now().Unix() > cv.ExpireAt {
		go c.refresh(key, id) // 旧值先返回，后台刷新
	}
	return cv.Data, nil
}

func (c *ProductCache) refresh(key, id string) {
	c.sf.Do(key, func() (any, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		val, err := c.load(ctx, id)
		if err != nil {
			return nil, err
		}
		cv := cachedVal{Data: val, ExpireAt: time.Now().Add(10 * time.Minute).Unix()}
		b, _ := json.Marshal(cv)
		c.rdb.Set(ctx, key, b, 24*time.Hour) // 兜底 TTL
		return nil, nil
	})
}
