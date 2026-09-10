package repo

import (
	"context"
	"database/sql"
	"time"

	"ykt.dev/admin/internal/db"
)

// MetricsRepo 5 大维度 + 设备/全局日聚合（只读 ykt_admin 库）。
// 表字段名以 DB 实际为准（camelCase 主键）。
type MetricsRepo struct{ DB *sql.DB }

func NewMetricsRepo(db *sql.DB) *MetricsRepo { return &MetricsRepo{DB: db} }

// ===== 连接维度（基于 ykt_event_connection，每行代表一次连接生命周期）=====

func (r *MetricsRepo) ConnectionSummary(ctx context.Context, from, to time.Time) (map[string]any, error) {
	rows, err := db.QueryRows(ctx, r.DB, `
		SELECT
			COUNT(*) AS total,
			SUM(quickDisconnect) AS quick_disconnect,
			SUM(timeoutOffline) AS timeout_offline,
			AVG(durationSec) AS avg_duration_sec
		FROM ykt_event_connection
		WHERE disconnectTime BETWEEN ? AND ?`, from, to)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]any{"total": 0, "quickDisconnect": 0, "timeoutOffline": 0, "avgDurationSec": 0}, nil
	}
	r0 := rows[0]
	return map[string]any{
		"total":           r0["total"],
		"quickDisconnect": r0["quick_disconnect"],
		"timeoutOffline":  r0["timeout_offline"],
		"avgDurationSec":  r0["avg_duration_sec"],
	}, nil
}

func (r *MetricsRepo) ConnectionAbnormalDevices(ctx context.Context, from, to time.Time) ([]map[string]any, error) {
	return db.QueryRows(ctx, r.DB, `
		SELECT deviceId,
		       COUNT(*) AS disconnect_count,
		       SUM(quickDisconnect) AS quick_count,
		       MAX(disconnectTime) AS last_disconnect
		FROM ykt_event_connection
		WHERE disconnectTime BETWEEN ? AND ?
		GROUP BY deviceId
		HAVING disconnect_count >= 3 OR quick_count >= 1
		ORDER BY disconnect_count DESC
		LIMIT 20`, from, to)
}

func (r *MetricsRepo) ConnectionEvents(ctx context.Context, deviceID string, from, to time.Time, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 50
	}
	q := `SELECT eventId, deviceId, IFNULL(instanceId,''), connectTime, disconnectTime,
		       durationSec, quickDisconnect, timeoutOffline, reason, IFNULL(detail,'')
		FROM ykt_event_connection
		WHERE disconnectTime BETWEEN ? AND ?`
	args := []any{from, to}
	if deviceID != "" {
		q += " AND deviceId = ?"
		args = append(args, deviceID)
	}
	q += " ORDER BY disconnectTime DESC LIMIT ?"
	args = append(args, limit)
	return db.QueryRows(ctx, r.DB, q, args...)
}

// ===== 对话维度 =====

func (r *MetricsRepo) DialogSummary(ctx context.Context, from, to time.Time) (map[string]any, error) {
	// ykt_metrics_device_daily 聚合字段
	rows, err := db.QueryRows(ctx, r.DB, `
		SELECT
			IFNULL(SUM(dialogTotalTurns), 0) AS total_turns,
			IFNULL(SUM(dialogFailedTurns), 0) AS failed_turns,
			IFNULL(SUM(sttFailedCount), 0) AS stt_failed,
			IFNULL(SUM(llmFailedCount), 0) AS llm_failed,
			IFNULL(SUM(ttsFailedCount), 0) AS tts_failed,
			IFNULL(AVG(latencyAvgMs), 0) AS avg_latency_ms
		FROM ykt_metrics_device_daily
		WHERE statDate BETWEEN ? AND ?`, from, to)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]any{}, nil
	}
	r0 := rows[0]
	return map[string]any{
		"totalTurns":    r0["total_turns"],
		"failedTurns":   r0["failed_turns"],
		"sttFailed":     r0["stt_failed"],
		"llmFailed":     r0["llm_failed"],
		"ttsFailed":     r0["tts_failed"],
		"avgLatencyMs":  r0["avg_latency_ms"],
	}, nil
}

func (r *MetricsRepo) DialogAbnormalDevices(ctx context.Context, from, to time.Time) ([]map[string]any, error) {
	return db.QueryRows(ctx, r.DB, `
		SELECT deviceId,
		       IFNULL(SUM(dialogFailedTurns), 0) AS failure_count,
		       IFNULL(SUM(slowReplyCount), 0) AS slow_count,
		       MAX(statDate) AS last_seen
		FROM ykt_metrics_device_daily
		WHERE statDate BETWEEN ? AND ? AND hasProblem = 1
		GROUP BY deviceId
		ORDER BY failure_count DESC
		LIMIT 20`, from, to)
}

func (r *MetricsRepo) DialogFailures(ctx context.Context, deviceID string, from, to time.Time, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 50
	}
	q := `SELECT deviceId, statDate, dialogFailedTurns, sttFailedCount, llmFailedCount, ttsFailedCount,
		       IFNULL(lastFailureStep,'') AS last_step, IFNULL(lastFailureReason,'') AS last_reason,
		       IFNULL(lastFailureTime,'') AS last_time
		FROM ykt_metrics_device_daily
		WHERE statDate BETWEEN ? AND ? AND dialogFailedTurns > 0`
	args := []any{from, to}
	if deviceID != "" {
		q += " AND deviceId = ?"
		args = append(args, deviceID)
	}
	q += " ORDER BY lastFailureTime DESC LIMIT ?"
	args = append(args, limit)
	return db.QueryRows(ctx, r.DB, q, args...)
}

// ===== 动作维度 =====

func (r *MetricsRepo) ActionSummary(ctx context.Context, from, to time.Time) (map[string]any, error) {
	rows, err := db.QueryRows(ctx, r.DB, `
		SELECT
			IFNULL(SUM(actionDispatchCount), 0) AS total,
			IFNULL(SUM(actionFailedCount), 0) AS failed,
			AVG(actionFailureRate) AS failure_rate
		FROM ykt_metrics_device_daily
		WHERE statDate BETWEEN ? AND ?`, from, to)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]any{}, nil
	}
	r0 := rows[0]
	return map[string]any{
		"total":       r0["total"],
		"failed":      r0["failed"],
		"failureRate": r0["failure_rate"],
	}, nil
}

func (r *MetricsRepo) ActionFailures(ctx context.Context, deviceID string, from, to time.Time, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 50
	}
	q := `SELECT deviceId, statDate, actionDispatchCount, actionFailedCount, actionFailureRate,
		       IFNULL(lastActionFailureTime,'') AS last_failure_time
		FROM ykt_metrics_device_daily
		WHERE statDate BETWEEN ? AND ? AND actionFailedCount > 0`
	args := []any{from, to}
	if deviceID != "" {
		q += " AND deviceId = ?"
		args = append(args, deviceID)
	}
	q += " ORDER BY lastActionFailureTime DESC LIMIT ?"
	args = append(args, limit)
	return db.QueryRows(ctx, r.DB, q, args...)
}

// ===== 延迟维度 =====

func (r *MetricsRepo) LatencySummary(ctx context.Context, from, to time.Time) (map[string]any, error) {
	rows, err := db.QueryRows(ctx, r.DB, `
		SELECT
			AVG(latencyAvgMs) AS avg_ms,
			MAX(latencyMaxMs) AS max_ms,
			MIN(latencyAvgMs) AS min_ms,
			AVG(latencyP95Ms) AS p95_ms,
			AVG(latencyP99Ms) AS p99_ms
		FROM ykt_metrics_device_daily
		WHERE statDate BETWEEN ? AND ?`, from, to)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]any{}, nil
	}
	r0 := rows[0]
	return map[string]any{
		"avgMs": r0["avg_ms"], "maxMs": r0["max_ms"],
		"minMs": r0["min_ms"], "p95Ms": r0["p95_ms"], "p99Ms": r0["p99_ms"],
	}, nil
}

func (r *MetricsRepo) LatencyTrend(ctx context.Context, from, to time.Time, step string) ([]map[string]any, error) {
	if step == "" {
		step = "day"
	}
	return db.QueryRows(ctx, r.DB, `
		SELECT statDate AS bucket,
		       AVG(latencyAvgMs) AS avg_ms,
		       MAX(latencyP95Ms) AS p95_ms,
		       SUM(latencySampleCount) AS sample_count
		FROM ykt_metrics_device_daily
		WHERE statDate BETWEEN ? AND ?
		GROUP BY statDate
		ORDER BY statDate`, from, to)
}

// ===== 离线维度 =====

func (r *MetricsRepo) OfflineDevices(ctx context.Context, xzDB *sql.DB) ([]map[string]any, error) {
	return db.QueryRows(ctx, xzDB, `
		SELECT deviceId, IFNULL(deviceName,'') AS device_name, IFNULL(version,''), state,
		       IFNULL(ip,''), IFNULL(updateTime,'') AS last_seen
		FROM sys_device
		WHERE state = '0'
		ORDER BY IFNULL(updateTime, '') DESC
		LIMIT 100`)
}

func (r *MetricsRepo) OfflineHistory(ctx context.Context, deviceID string, from, to time.Time, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 50
	}
	return db.QueryRows(ctx, r.DB, `
		SELECT deviceId, statDate, offlineDurationSec, offlineEventCount, hasProblem
		FROM ykt_metrics_device_daily
		WHERE statDate BETWEEN ? AND ? AND offlineDurationSec > 0
		ORDER BY statDate DESC LIMIT ?`, from, to, limit)
}

// ===== 设备日聚合 =====

func (r *MetricsRepo) DeviceDailyList(ctx context.Context, from, to time.Time, noFilter bool, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 100
	}
	if noFilter {
		return db.QueryRows(ctx, r.DB, `
			SELECT deviceId, statDate, state,
			       connectAttempts, connectSuccess, connectFailed, connectFailureRate,
			       disconnectCount, reconnectCount, onlineDurationSec,
			       dialogTotalTurns, dialogFailedTurns, dialogFailureRate,
			       sttFailedCount, llmFailedCount, ttsFailedCount,
			       latencyAvgMs, latencyP95Ms, latencyP99Ms, latencyMaxMs,
			       actionDispatchCount, actionFailedCount, actionFailureRate,
			       hasProblem, IFNULL(problemTypes,'') AS problem_types
			FROM ykt_metrics_device_daily
			ORDER BY statDate DESC, deviceId
			LIMIT ?`, limit)
	}
	return db.QueryRows(ctx, r.DB, `
		SELECT deviceId, statDate, state,
		       connectAttempts, connectSuccess, connectFailed, connectFailureRate,
		       disconnectCount, reconnectCount, onlineDurationSec,
		       dialogTotalTurns, dialogFailedTurns, dialogFailureRate,
		       sttFailedCount, llmFailedCount, ttsFailedCount,
		       latencyAvgMs, latencyP95Ms, latencyP99Ms, latencyMaxMs,
		       actionDispatchCount, actionFailedCount, actionFailureRate,
		       hasProblem, IFNULL(problemTypes,'') AS problem_types
		FROM ykt_metrics_device_daily
		WHERE statDate BETWEEN ? AND ?
		ORDER BY statDate DESC, deviceId
		LIMIT ?`, from, to, limit)
}

func (r *MetricsRepo) GlobalDailyList(ctx context.Context, from, to time.Time, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 100
	}
	return db.QueryRows(ctx, r.DB, `
		SELECT statDate, totalDevices, onlineDevices, offlineDevices, onlineRate,
		       problemDevices, connectAttempts, connectFailed, disconnectTotal,
		       dialogTotalTurns, dialogFailedTurns, sttFailedTotal, llmFailedTotal, ttsFailedTotal,
		       tokenUsageTotal, latencyAvgMs, latencyP95Ms,
		       actionDispatchTotal, actionFailedTotal
		FROM ykt_metrics_global_daily
		WHERE statDate BETWEEN ? AND ?
		ORDER BY statDate DESC
		LIMIT ?`, from, to, limit)
}

func (r *MetricsRepo) ProblemDevices(ctx context.Context, from, to time.Time, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 20
	}
	return db.QueryRows(ctx, r.DB, `
		SELECT deviceId,
		       IFNULL(SUM(disconnectCount), 0) AS disconnects,
		       IFNULL(SUM(dialogFailedTurns), 0) AS dialog_failures,
		       IFNULL(SUM(actionFailedCount), 0) AS action_failures,
		       AVG(latencyAvgMs) AS avg_latency_ms,
		       IFNULL(GROUP_CONCAT(DISTINCT problemTypes), '') AS problem_types
		FROM ykt_metrics_device_daily
		WHERE statDate BETWEEN ? AND ? AND hasProblem = 1
		GROUP BY deviceId
		ORDER BY (disconnects + dialog_failures + action_failures) DESC
		LIMIT ?`, from, to, limit)
}

// ===== AI 监控（只读 ykt_aisaas 库）=====
//
// 数据源：
//   - ykt_aisaas_usage_daily(tenant_id, usage_date, model_type, total_calls, total_tokens, total_chars, total_seconds, total_cost)
//   - ykt_aisaas_usage_detail(bizType, dimension, amount, costCents, modelId, createTime)
//   - ykt_aisaas_rag_log(tenant_id, query, top_score, latency_ms, created_at)
//   - ykt_aisaas_quota(tenantId, periodStart, periodEnd, dimension, limitValue, usedValue)
//   - ykt_aisaas_balance(tenantId, balanceCents, totalConsumed)

// AILlm 按天聚合 LLM 用量与成本。
// 历史路径：ykt_aisaas_usage_daily（CHAT 行）。
// 实时路径：直接从 ykt_aisaas_usage_detail 当天明细聚合补齐 usage_daily 还没刷的部分。
func (r *MetricsRepo) AILlm(ctx context.Context, aiDB *sql.DB, from, to time.Time) ([]map[string]any, error) {
	return db.QueryRows(ctx, aiDB, `
		SELECT usage_date,
		       SUM(calls) AS calls,
		       SUM(tokens) AS tokens,
		       SUM(cost) AS cost
		FROM (
			SELECT usage_date,
			       SUM(total_calls) AS calls,
			       SUM(total_tokens) AS tokens,
			       SUM(total_cost) AS cost
			FROM ykt_aisaas_usage_daily
			WHERE model_type = 'CHAT'
			  AND usage_date BETWEEN ? AND ?
			GROUP BY usage_date
			UNION ALL
			SELECT DATE(createTime) AS usage_date,
			       COUNT(*) AS calls,
			       SUM(amount) AS tokens,
			       0 AS cost
			FROM ykt_aisaas_usage_detail
			WHERE bizType = 'llm'
			  AND DATE(createTime) BETWEEN ? AND ?
			GROUP BY DATE(createTime)
		) t
		GROUP BY usage_date
		ORDER BY usage_date ASC`,
		from.Format("2006-01-02"), to.Format("2006-01-02"),
		from.Format("2006-01-02"), to.Format("2006-01-02"))
}

// AITts 按天聚合 TTS 用量（按字符数）。
// 实时路径：detail.dimension = 'tts_chars' 当天聚合。
func (r *MetricsRepo) AITts(ctx context.Context, aiDB *sql.DB, from, to time.Time) ([]map[string]any, error) {
	return db.QueryRows(ctx, aiDB, `
		SELECT usage_date,
		       SUM(calls) AS calls,
		       SUM(chars) AS chars,
		       SUM(seconds) AS seconds,
		       SUM(cost) AS cost
		FROM (
			SELECT usage_date,
			       SUM(total_calls) AS calls,
			       SUM(total_chars) AS chars,
			       SUM(total_seconds) AS seconds,
			       SUM(total_cost) AS cost
			FROM ykt_aisaas_usage_daily
			WHERE model_type = 'TTS'
			  AND usage_date BETWEEN ? AND ?
			GROUP BY usage_date
			UNION ALL
			SELECT DATE(createTime) AS usage_date,
			       COUNT(*) AS calls,
			       IFNULL(SUM(amount), 0) AS chars,
			       0 AS seconds,
			       0 AS cost
			FROM ykt_aisaas_usage_detail
			WHERE bizType = 'tts'
			  AND DATE(createTime) BETWEEN ? AND ?
			GROUP BY DATE(createTime)
		) t
		GROUP BY usage_date
		ORDER BY usage_date ASC`,
		from.Format("2006-01-02"), to.Format("2006-01-02"),
		from.Format("2006-01-02"), to.Format("2006-01-02"))
}

// AIAsr 按天聚合 ASR 用量（按时长）。
// 实时路径：detail.dimension = 'asr_seconds' 当天聚合。
func (r *MetricsRepo) AIAsr(ctx context.Context, aiDB *sql.DB, from, to time.Time) ([]map[string]any, error) {
	return db.QueryRows(ctx, aiDB, `
		SELECT usage_date,
		       SUM(calls) AS calls,
		       SUM(seconds) AS seconds,
		       SUM(cost) AS cost
		FROM (
			SELECT usage_date,
			       SUM(total_calls) AS calls,
			       SUM(total_seconds) AS seconds,
			       SUM(total_cost) AS cost
			FROM ykt_aisaas_usage_daily
			WHERE model_type = 'ASR'
			  AND usage_date BETWEEN ? AND ?
			GROUP BY usage_date
			UNION ALL
			SELECT DATE(createTime) AS usage_date,
			       COUNT(*) AS calls,
			       IFNULL(SUM(amount), 0) AS seconds,
			       0 AS cost
			FROM ykt_aisaas_usage_detail
			WHERE bizType = 'asr'
			  AND DATE(createTime) BETWEEN ? AND ?
			GROUP BY DATE(createTime)
		) t
		GROUP BY usage_date
		ORDER BY usage_date ASC`,
		from.Format("2006-01-02"), to.Format("2006-01-02"),
		from.Format("2006-01-02"), to.Format("2006-01-02"))
}

// AIRag 最近 RAG 检索日志（含相关性分数、延迟）
func (r *MetricsRepo) AIRag(ctx context.Context, aiDB *sql.DB, from, to time.Time, limit int) ([]map[string]any, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	return db.QueryRows(ctx, aiDB, `
		SELECT id, tenant_id, device_id, knowledge_base_id,
		       LEFT(query, 200) AS query,
		       top_score, latency_ms, created_at
		FROM ykt_aisaas_rag_log
		WHERE created_at BETWEEN ? AND ?
		ORDER BY id DESC
		LIMIT ?`,
		from, to, limit)
}

// AIQuotaTrend 配额使用趋势：每个租户的配额使用率
func (r *MetricsRepo) AIQuotaTrend(ctx context.Context, aiDB *sql.DB) ([]map[string]any, error) {
	return db.QueryRows(ctx, aiDB, `
		SELECT tenantId,
		       dimension,
		       periodStart,
		       periodEnd,
		       limitValue,
		       usedValue,
		       IF(limitValue > 0, ROUND(usedValue / limitValue, 4), 0) AS usage_rate
		FROM ykt_aisaas_quota
		ORDER BY periodEnd DESC
		LIMIT 200`)
}
