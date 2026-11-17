package store

import "database/sql"

// MySQLStore 是 Store 的 MySQL 实现，生产环境用。
type MySQLStore struct {
	db *sql.DB
}

func NewMySQL(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

func (s *MySQLStore) Save(code, url string) error {
	_, err := s.db.Exec(
		"INSERT INTO short_links(code, url) VALUES(?, ?) ON DUPLICATE KEY UPDATE url=VALUES(url)",
		code, url,
	)
	return err
}

func (s *MySQLStore) Get(code string) (string, error) {
	var url string
	err := s.db.QueryRow("SELECT url FROM short_links WHERE code = ?", code).Scan(&url)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	return url, err
}

func (s *MySQLStore) Exists(url string) (string, bool) {
	var code string
	err := s.db.QueryRow("SELECT code FROM short_links WHERE url = ?", url).Scan(&code)
	return code, err == nil
}
