package service

import (
	"sync/atomic"

	"shortlink/internal/base62"
	"shortlink/internal/cache"
	"shortlink/internal/store"
)

type Shortener struct {
	store store.Store
	cache *cache.Cache
	seq   atomic.Int64
}

func New(st store.Store, c *cache.Cache) *Shortener {
	return &Shortener{store: st, cache: c}
}

// Shorten 创建短链；同一个长网址返回已有短码（幂等）。
func (s *Shortener) Shorten(url string) (string, error) {
	if code, ok := s.store.Exists(url); ok {
		return code, nil
	}
	id := s.seq.Add(1)
	code := base62.Encode(id)
	if err := s.store.Save(code, url); err != nil {
		return "", err
	}
	s.cache.Set(code, url)
	return code, nil
}

// Resolve 查询短码对应长网址，先查缓存再查存储。
func (s *Shortener) Resolve(code string) (string, error) {
	if url, ok := s.cache.Get(code); ok {
		return url, nil
	}
	url, err := s.store.Get(code)
	if err != nil {
		return "", err
	}
	s.cache.Set(code, url)
	return url, nil
}
