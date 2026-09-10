package repo

import (
	"context"
	"database/sql"
	"time"
)

type Tenant struct {
	ID       int64
	Code     string
	Name     string
	Status   int8
	IsAdmin  int8
}

type TenantRepo struct{ DB *sql.DB }

func NewTenantRepo(db *sql.DB) *TenantRepo { return &TenantRepo{DB: db} }

func (r *TenantRepo) List(ctx context.Context) ([]Tenant, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, code, name, status, is_admin
		FROM ykt_sys_tenant
		WHERE del_flag = '0' AND status = '0'
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Code, &t.Name, &t.Status, &t.IsAdmin); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TenantRepo) GetByID(ctx context.Context, id int64) (*Tenant, error) {
	t := &Tenant{}
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, code, name, status, is_admin
		FROM ykt_sys_tenant
		WHERE id = ? AND del_flag = '0'`, id)
	if err := row.Scan(&t.ID, &t.Code, &t.Name, &t.Status, &t.IsAdmin); err != nil {
		return nil, err
	}
	return t, nil
}

// TimeRange 默认 7 天（muliti-toy 前端默认查询范围）。
func DefaultTimeRange() (time.Time, time.Time) {
	now := time.Now()
	return now.AddDate(0, 0, -7), now
}
