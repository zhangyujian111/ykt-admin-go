package repo

import (
	"context"
	"database/sql"
)

// AlertRuleDO ykt_alert_rule（实际字段名是 camelCase 主键）。
type AlertRuleDO struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Metric     string     `json:"metric"`
	Expression string     `json:"expression"`
	Threshold  *float64   `json:"threshold,omitempty"`
	Level      string     `json:"level"`
	Enabled    int        `json:"enabled"`
	Cooldown   *int       `json:"cooldown,omitempty"`
}

type AlertHistoryDO struct {
	ID         int64
	RuleID     int64
	RuleName   string
	Metric     string
	Value      float64
	Level      string
	Message    sql.NullString
	CreatedAt  sql.NullTime
	ResolvedAt sql.NullTime
	Status     string
	AckedBy    sql.NullInt64
	AckedAt    sql.NullTime
}

type AlertRepo struct{ DB *sql.DB }

func NewAlertRepo(db *sql.DB) *AlertRepo { return &AlertRepo{DB: db} }

func (r *AlertRepo) ListRules(ctx context.Context) ([]AlertRuleDO, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT ruleId, ruleName, metric, expression, threshold, level, enabled, cooldownSec
		FROM ykt_alert_rule
		WHERE enabled = 1
		ORDER BY ruleId`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AlertRuleDO
	for rows.Next() {
		var a AlertRuleDO
		var threshold sql.NullFloat64
		var cooldown sql.NullInt64
		if err := rows.Scan(&a.ID, &a.Name, &a.Metric, &a.Expression, &threshold, &a.Level, &a.Enabled, &cooldown); err != nil {
			return nil, err
		}
		if threshold.Valid {
			v := threshold.Float64
			a.Threshold = &v
		}
		if cooldown.Valid {
			v := int(cooldown.Int64)
			a.Cooldown = &v
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AlertRepo) ListHistory(ctx context.Context, page, pageSize int) ([]AlertHistoryDO, int64, error) {
	var total int64
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM ykt_alert_history`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT h.alertId, h.ruleId, COALESCE(r.ruleName,'') AS rule_name, h.metric,
		       h.currentValue, h.level, h.message, h.triggeredAt, h.recoveredAt, h.status,
		       h.ackedBy, h.ackedAt
		FROM ykt_alert_history h
		LEFT JOIN ykt_alert_rule r ON r.ruleId = h.ruleId
		ORDER BY h.triggeredAt DESC
		LIMIT ? OFFSET ?`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []AlertHistoryDO
	for rows.Next() {
		var a AlertHistoryDO
		if err := rows.Scan(&a.ID, &a.RuleID, &a.RuleName, &a.Metric, &a.Value, &a.Level,
			&a.Message, &a.CreatedAt, &a.ResolvedAt, &a.Status, &a.AckedBy, &a.AckedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, a)
	}
	return out, total, rows.Err()
}

func (r *AlertRepo) ActiveAlerts(ctx context.Context) ([]AlertHistoryDO, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT h.alertId, h.ruleId, COALESCE(r.ruleName,''), h.metric,
		       h.currentValue, h.level, h.message, h.triggeredAt, h.recoveredAt, h.status,
		       h.ackedBy, h.ackedAt
		FROM ykt_alert_history h
		LEFT JOIN ykt_alert_rule r ON r.ruleId = h.ruleId
		WHERE h.status = 'ACTIVE'
		ORDER BY h.triggeredAt DESC
		LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AlertHistoryDO
	for rows.Next() {
		var a AlertHistoryDO
		if err := rows.Scan(&a.ID, &a.RuleID, &a.RuleName, &a.Metric, &a.Value, &a.Level,
			&a.Message, &a.CreatedAt, &a.ResolvedAt, &a.Status, &a.AckedBy, &a.AckedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AlertRepo) Ack(ctx context.Context, id int64, userID int64) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE ykt_alert_history SET status='ACK', ackedBy=?, ackedAt=NOW() WHERE alertId = ?`, userID, id)
	return err
}
