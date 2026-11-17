package store

import "sync"

// MemoryStore 是 Store 的内存实现，用双向索引加速查询。
type MemoryStore struct {
	mu     sync.RWMutex
	byCode map[string]string // code -> url
	byURL  map[string]string // url  -> code
}

func NewMemory() *MemoryStore {
	return &MemoryStore{
		byCode: make(map[string]string),
		byURL:  make(map[string]string),
	}
}

func (s *MemoryStore) Save(code, url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byCode[code] = url
	s.byURL[url] = code
	return nil
}

func (s *MemoryStore) Get(code string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.byCode[code]
	if !ok {
		return "", ErrNotFound
	}
	return url, nil
}

func (s *MemoryStore) Exists(url string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	code, ok := s.byURL[url]
	return code, ok
}
