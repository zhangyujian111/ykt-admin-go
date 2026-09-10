package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// 单库数据源：所有业务表（ykt_sys_* / ykt_alert_* / ykt_event_* / sys_* /
// ykt_aisaas_*）统一落在同一个库，不做跨库查询。
type Store struct {
	Admin *sql.DB
}

type Config struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLife  time.Duration
}

func openDB(c Config) (*sql.DB, error) {
	d, err := sql.Open("mysql", c.DSN)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if c.MaxOpenConns > 0 {
		d.SetMaxOpenConns(c.MaxOpenConns)
	}
	if c.MaxIdleConns > 0 {
		d.SetMaxIdleConns(c.MaxIdleConns)
	}
	if c.ConnMaxLife > 0 {
		d.SetConnMaxLifetime(c.ConnMaxLife)
	}
	if err := d.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return d, nil
}

// NewStore 打开单个数据库连接池。
func NewStore(cfg Config) (*Store, error) {
	d, err := openDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("mysql: %w", err)
	}
	return &Store{Admin: d}, nil
}

func (s *Store) Close() {
	if s.Admin != nil {
		_ = s.Admin.Close()
	}
}
