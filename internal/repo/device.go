package repo

import (
	"context"
	"database/sql"
)

// DeviceRepo 设备元数据（只读 xiaozhi.sys_device）。
type DeviceRepo struct{ DB *sql.DB }

func NewDeviceRepo(db *sql.DB) *DeviceRepo { return &DeviceRepo{DB: db} }

func (r *DeviceRepo) List(ctx context.Context, keyword, state string, page, pageSize int) ([]map[string]any, int64, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}

	where := " WHERE 1=1 "
	args := []any{}
	if keyword != "" {
		where += " AND (deviceId LIKE ? OR deviceName LIKE ?) "
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	if state == "online" {
		where += " AND state = '1' "
	} else if state == "offline" {
		where += " AND state = '0' "
	}

	var total int64
	if err := r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys_device"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `SELECT deviceId, deviceName, IFNULL(type,''), IFNULL(version,''), IFNULL(chipModelName,''),
	       state, IFNULL(ip,''), IFNULL(createTime,''), IFNULL(updateTime,'')
	FROM sys_device` + where + " ORDER BY IFNULL(updateTime, '') DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, name, t, ver, chip, state, ip, ct, ut sql.NullString
		if err := rows.Scan(&id, &name, &t, &ver, &chip, &state, &ip, &ct, &ut); err != nil {
			return nil, 0, err
		}
		out = append(out, map[string]any{
			"deviceId":      id.String,
			"deviceName":    name.String,
			"type":          t.String,
			"version":       ver.String,
			"chipModelName": chip.String,
			"state":         state.String,
			"ip":            ip.String,
			"createTime":    ct.String,
			"updateTime":    ut.String,
		})
	}
	return out, total, rows.Err()
}

func (r *DeviceRepo) VersionDistribution(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT IFNULL(version, 'unknown') AS version, COUNT(*) AS count
		FROM sys_device
		GROUP BY version
		ORDER BY count DESC
		LIMIT 20`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var v string
		var c int64
		if err := rows.Scan(&v, &c); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"version": v, "count": c})
	}
	return out, rows.Err()
}

func (r *DeviceRepo) TypeDistribution(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT IFNULL(type, 'unknown') AS type, COUNT(*) AS count
		FROM sys_device
		GROUP BY type
		ORDER BY count DESC
		LIMIT 20`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var t string
		var c int64
		if err := rows.Scan(&t, &c); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"type": t, "count": c})
	}
	return out, rows.Err()
}

func (r *DeviceRepo) Detail(ctx context.Context, deviceID string) (map[string]any, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT deviceId, deviceName, IFNULL(type,''), IFNULL(version,''), IFNULL(chipModelName,''),
		       state, IFNULL(ip,''), IFNULL(location,''), IFNULL(wifiName,''),
		       IFNULL(createTime,''), IFNULL(updateTime,'')
		FROM sys_device WHERE deviceId = ?`, deviceID)
	out := map[string]any{}
	var id, name, t, ver, chip, state, ip, loc, wifi, ct, ut sql.NullString
	if err := row.Scan(&id, &name, &t, &ver, &chip, &state, &ip, &loc, &wifi, &ct, &ut); err != nil {
		return nil, err
	}
	out["deviceId"] = id.String
	out["deviceName"] = name.String
	out["type"] = t.String
	out["version"] = ver.String
	out["chipModelName"] = chip.String
	out["state"] = state.String
	out["ip"] = ip.String
	out["location"] = loc.String
	out["wifiName"] = wifi.String
	out["createTime"] = ct.String
	out["updateTime"] = ut.String
	return out, nil
}
