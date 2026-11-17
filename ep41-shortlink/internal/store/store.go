package store

import "errors"

// ErrNotFound 表示短码不存在。
var ErrNotFound = errors.New("short link not found")

// Store 是短链存储的抽象接口。
type Store interface {
	Save(code, url string) error
	Get(code string) (string, error)
	// Exists 用于幂等：同一个长网址是否已生成过短码。
	Exists(url string) (string, bool)
}
