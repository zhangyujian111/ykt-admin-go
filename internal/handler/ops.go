package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"ykt.dev/admin/internal/common"
	"ykt.dev/admin/internal/db"
	"ykt.dev/admin/internal/repo"
)

// OpsHandler 5 大维度运维接口（/api/ops/*）。
type OpsHandler struct {
	Store   *db.Store
	Metrics *repo.MetricsRepo
}

// 时间范围 helper：?from=2026-09-01&to=2026-09-08 或默认 7 天。
func parseRange(c *gin.Context) (from, to time.Time, err error) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		from, to = repo.DefaultTimeRange()
		return
	}
	from, err = time.Parse("2006-01-02", fromStr)
	if err != nil {
		return
	}
	to, err = time.Parse("2006-01-02", toStr)
	if err != nil {
		return
	}
	to = to.Add(24 * time.Hour)
	return
}

// ============== Dashboard 大盘 ==============

func (h *OpsHandler) DashboardSummary(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	var totalDialogs, totalFailures int64
	var avgLat sql.NullFloat64
	var activeAlerts int64

	_ = h.Store.Admin.QueryRowContext(ctx, `
		SELECT IFNULL(SUM(dialogTotalTurns), 0), IFNULL(SUM(dialogFailedTurns), 0)
		FROM ykt_metrics_global_daily WHERE statDate BETWEEN ? AND ?`,
		from, to).Scan(&totalDialogs, &totalFailures)
	_ = h.Store.Admin.QueryRowContext(ctx, `
		SELECT AVG(latencyAvgMs) FROM ykt_metrics_device_daily WHERE statDate BETWEEN ? AND ?`,
		from, to).Scan(&avgLat)
	_ = h.Store.Admin.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM ykt_alert_history WHERE status='ACTIVE'`).Scan(&activeAlerts)

	avg := 0.0
	if avgLat.Valid {
		avg = avgLat.Float64
	}
	c.JSON(http.StatusOK, common.Ok(map[string]any{
		"totalDialogs":  totalDialogs,
		"totalFailures": totalFailures,
		"avgLatencyMs":  avg,
		"activeAlerts":  activeAlerts,
	}))
}

func (h *OpsHandler) ProblemDevices(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	rows, err := h.Metrics.ProblemDevices(ctx, from, to, limit)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *OpsHandler) ErrorTypeStats(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	rows, err := db.QueryRows(ctx, h.Store.Admin, `
		SELECT 'DIALOG' AS type, COUNT(*) AS count FROM ykt_event_dialog_failure WHERE created_at BETWEEN ? AND ?
		UNION ALL
		SELECT 'ACTION' AS type, COUNT(*) AS count FROM ykt_event_action_failure WHERE created_at BETWEEN ? AND ?
		UNION ALL
		SELECT 'CONNECTION' AS type, COUNT(*) AS count FROM ykt_event_connection WHERE event_type='DISCONNECT' AND created_at BETWEEN ? AND ?`,
		from, to, from, to, from, to)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

// ============== Devices Overview（4 tab） ==============
//
// 数据源：ykt_metrics_device_daily（DB 字段是 statDate/deviceId/state/...
// 输出字段映射到前端 DevicesOverview 期望的形状（ConnectionRecord / SessionRecord / LatencyRecord / ActionRecord）。

// parseRangeAny 支持 from/to 或 startTime/endTime（前端两种都用）。
// 返回值：
//   from/to  — 解析后的时间范围
//   noFilter — true 时表示前端两个日期都为空，过滤条件跳过，返所有数据
func parseRangeAny(c *gin.Context) (from, to time.Time, noFilter bool) {
	fromStr := c.Query("from")
	if fromStr == "" {
		fromStr = c.Query("startTime")
	}
	toStr := c.Query("to")
	if toStr == "" {
		toStr = c.Query("endTime")
	}
	// 两个都为空 = 不过滤
	if fromStr == "" && toStr == "" {
		return time.Time{}, time.Time{}, true
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now()
	if toStr == "" {
		to = now
	} else if t, err := time.ParseInLocation("2006-01-02", toStr, loc); err == nil {
		to = t
	} else {
		to = now
	}
	if fromStr == "" {
		from = to.AddDate(0, 0, -7)
	} else if t, err := time.ParseInLocation("2006-01-02", fromStr, loc); err == nil {
		from = t
	} else {
		from = to.AddDate(0, 0, -7)
	}
	return
}

// deviceNameMap 从 xiaozhi.sys_device 取 deviceName
func (h *OpsHandler) deviceNameMap(ctx context.Context) (map[string]string, error) {
	rows, err := h.Store.Admin.QueryContext(ctx, `SELECT deviceId, IFNULL(deviceName, '') FROM sys_device`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]string)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err == nil {
			out[id] = name
		}
	}
	return out, nil
}

// lastEventTimes 取每个 deviceId 的最近一次 connectTime / disconnectTime
type deviceEventTimes struct {
	onlineTime  *string
	offlineTime *string
}

func (h *OpsHandler) lastEventTimes(ctx context.Context, deviceIDs []string) (map[string]deviceEventTimes, error) {
	out := make(map[string]deviceEventTimes)
	if len(deviceIDs) == 0 {
		return out, nil
	}
	// 用参数化 IN
	q := `SELECT deviceId,
		MAX(connectTime) AS lastOnline,
		MAX(disconnectTime) AS lastOffline
	  FROM ykt_event_connection
	  WHERE deviceId IN (?` + strings.Repeat(",?", len(deviceIDs)-1) + `)
	  GROUP BY deviceId`
	args := make([]any, len(deviceIDs))
	for i, v := range deviceIDs {
		args[i] = v
	}
	rows, err := h.Store.Admin.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var online, offline sql.NullString
		if err := rows.Scan(&id, &online, &offline); err == nil {
			et := deviceEventTimes{}
			if online.Valid {
				v := online.String
				et.onlineTime = &v
			}
			if offline.Valid {
				v := offline.String
				et.offlineTime = &v
			}
			out[id] = et
		}
	}
	return out, nil
}

// connectionStateLabel: state('0'/'1'/'2' 或 'O') → online/offline 文本
func connectionStateLabel(v interface{}) string {
	s, _ := v.(string)
	switch s {
	case "1", "O", "online":
		return "online"
	default:
		return "offline"
	}
}

// 在线秒数 → "1h 23m" 文本
func formatOnlineDuration(sec int64) string {
	if sec <= 0 {
		return "—"
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func (h *OpsHandler) DevicesOverviewConnection(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, noFilter := parseRangeAny(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	rows, err := h.Metrics.DeviceDailyList(ctx, from, to, noFilter, limit)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	names, _ := h.deviceNameMap(ctx)
	devIDs := make([]string, 0, len(rows))
	for _, r := range rows {
		if id, ok := r["deviceId"].(string); ok {
			devIDs = append(devIDs, id)
		}
	}
	events, _ := h.lastEventTimes(ctx, devIDs)
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		devID, _ := r["deviceId"].(string)
		stateRaw := r["state"]
		dur := toInt64(r["onlineDurationSec"])
		et := events[devID]
		out = append(out, map[string]any{
			"deviceId":           devID,
			"deviceName":         names[devID],
			"status":             connectionStateLabel(stateRaw),
			"onlineTime":         et.onlineTime,
			"offlineTime":        et.offlineTime,
			"connectionCount":    toInt(r["connectAttempts"]),
			"onlineDuration":     dur,
			"onlineDurationText": formatOnlineDuration(dur),
		})
	}
	c.JSON(http.StatusOK, common.Ok(out))
}

func (h *OpsHandler) DevicesOverviewSession(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, noFilter := parseRangeAny(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	rows, err := h.Metrics.DeviceDailyList(ctx, from, to, noFilter, limit)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	names, _ := h.deviceNameMap(ctx)
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		devID, _ := r["deviceId"].(string)
		total := toInt(r["dialogTotalTurns"])
		failed := toInt(r["dialogFailedTurns"])
		rate := float64(0)
		if total > 0 {
			rate = float64(failed) / float64(total) * 100
		}
		out = append(out, map[string]any{
			"deviceId":           devID,
			"deviceName":         names[devID],
			"sessionTotal":       total,
			"sessionFailures":    failed,
			"sessionFailureRate": rate,
		})
	}
	c.JSON(http.StatusOK, common.Ok(out))
}

// 延迟桶：根据 avgLatency 粗略划分（DB 没有细粒度分桶）
func latencyBuckets(avgMs int64) (b0, b1, b2, b3, b4, b5 int64) {
	if avgMs <= 1000 {
		b0 = 1
	} else if avgMs <= 2000 {
		b1 = 1
	} else if avgMs <= 3000 {
		b2 = 1
	} else if avgMs <= 4000 {
		b3 = 1
	} else if avgMs <= 5000 {
		b4 = 1
	} else {
		b5 = 1
	}
	return
}

func (h *OpsHandler) DevicesOverviewLatency(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, noFilter := parseRangeAny(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	rows, err := h.Metrics.DeviceDailyList(ctx, from, to, noFilter, limit)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	names, _ := h.deviceNameMap(ctx)
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		devID, _ := r["deviceId"].(string)
		avg := toInt64(r["latencyAvgMs"])
		total := toInt(r["dialogTotalTurns"])
		b0, b1, b2, b3, b4, b5 := latencyBuckets(avg)
		mk := func(n int64) (count int64, rate float64) {
			if total <= 0 {
				return n, 0
			}
			return n, float64(n) / float64(total) * 100
		}
		c0, r0 := mk(b0)
		c1, r1 := mk(b1)
		c2, r2 := mk(b2)
		c3, r3 := mk(b3)
		c4, r4 := mk(b4)
		c5, r5 := mk(b5)
		out = append(out, map[string]any{
			"deviceId":     devID,
			"deviceName":   names[devID],
			"sessionTotal": total,
			"delay0to1":    c0, "delay0to1Rate": r0,
			"delay1to2":    c1, "delay1to2Rate": r1,
			"delay2to3":    c2, "delay2to3Rate": r2,
			"delay3to4":    c3, "delay3to4Rate": r3,
			"delay4to5":    c4, "delay4to5Rate": r4,
			"delay5plus":   c5, "delay5plusRate": r5,
		})
	}
	c.JSON(http.StatusOK, common.Ok(out))
}

func (h *OpsHandler) DevicesOverviewAction(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, noFilter := parseRangeAny(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	rows, err := h.Metrics.DeviceDailyList(ctx, from, to, noFilter, limit)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	names, _ := h.deviceNameMap(ctx)
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		devID, _ := r["deviceId"].(string)
		dispatch := toInt(r["actionDispatchCount"])
		failed := toInt(r["actionFailedCount"])
		success := dispatch - failed
		if success < 0 {
			success = 0
		}
		rate := float64(100)
		if dispatch > 0 {
			rate = float64(success) / float64(dispatch) * 100
		}
		out = append(out, map[string]any{
			"deviceId":             devID,
			"deviceName":           names[devID],
			"actionDispatchCount":  dispatch,
			"actionSuccessCount":   success,
			"actionSuccessRate":    rate,
		})
	}
	c.JSON(http.StatusOK, common.Ok(out))
}

// helpers
func toInt(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case int32:
		return int(x)
	case int64:
		return int(x)
	case float64:
		return int(x)
	case string:
		n, _ := strconv.Atoi(x)
		return n
	}
	return 0
}

func toInt64(v interface{}) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case float64:
		return int64(x)
	case string:
		n, _ := strconv.ParseInt(x, 10, 64)
		return n
	}
	return 0
}

func (h *OpsHandler) DeviceDetail(c *gin.Context) {
	ctx := c.Request.Context()
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeBadRequest, "deviceId required"))
		return
	}
	from, to, _ := parseRange(c)
	summary, err := db.QueryRows(ctx, h.Store.Admin, `
		SELECT device_id, stat_date, online_seconds, session_count, dialog_count, action_count,
		       disconnect_count, dialog_failure_count, action_failure_count,
		       avg_dialog_latency_ms, IFNULL(last_connect_at, '') AS last_connect_at
		FROM ykt_metrics_device_daily
		WHERE device_id = ? AND stat_date BETWEEN ? AND ?
		ORDER BY stat_date DESC LIMIT 30`, deviceID, from, to)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	events, _ := h.Metrics.ConnectionEvents(ctx, deviceID, from, to, 20)
	c.JSON(http.StatusOK, common.Ok(map[string]any{
		"deviceId": deviceID,
		"summary":  summary,
		"events":   events,
	}))
}

// ============== 5 大维度（独立接口） ==============

func (h *OpsHandler) ConnectionSummary(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	out, err := h.Metrics.ConnectionSummary(ctx, from, to)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(out))
}

func (h *OpsHandler) ConnectionAbnormalDevices(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	rows, err := h.Metrics.ConnectionAbnormalDevices(ctx, from, to)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *OpsHandler) ConnectionEvents(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	deviceID := c.Query("deviceId")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	rows, err := h.Metrics.ConnectionEvents(ctx, deviceID, from, to, limit)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *OpsHandler) DialogSummary(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	out, err := h.Metrics.DialogSummary(ctx, from, to)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(out))
}

func (h *OpsHandler) DialogAbnormalDevices(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	rows, err := h.Metrics.DialogAbnormalDevices(ctx, from, to)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *OpsHandler) DialogFailures(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	deviceID := c.Query("deviceId")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	rows, err := h.Metrics.DialogFailures(ctx, deviceID, from, to, limit)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *OpsHandler) ActionSummary(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	out, err := h.Metrics.ActionSummary(ctx, from, to)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(out))
}

func (h *OpsHandler) ActionFailures(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	deviceID := c.Query("deviceId")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	rows, err := h.Metrics.ActionFailures(ctx, deviceID, from, to, limit)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *OpsHandler) LatencySummary(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	out, err := h.Metrics.LatencySummary(ctx, from, to)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(out))
}

func (h *OpsHandler) LatencyTrend(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	step := c.DefaultQuery("step", "hour")
	rows, err := h.Metrics.LatencyTrend(ctx, from, to, step)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *OpsHandler) OfflineDevices(c *gin.Context) {
	ctx := c.Request.Context()
	rows, err := h.Metrics.OfflineDevices(ctx, h.Store.Admin)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *OpsHandler) OfflineHistory(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	deviceID := c.Query("deviceId")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	rows, err := h.Metrics.OfflineHistory(ctx, deviceID, from, to, limit)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

// ============== AI 监控（只读 aisaas）==============

func (h *OpsHandler) AILlm(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	rows, err := h.Metrics.AILlm(ctx, h.Store.Admin, from, to)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *OpsHandler) AITts(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	rows, err := h.Metrics.AITts(ctx, h.Store.Admin, from, to)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *OpsHandler) AIAsr(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	rows, err := h.Metrics.AIAsr(ctx, h.Store.Admin, from, to)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *OpsHandler) AIRag(c *gin.Context) {
	ctx := c.Request.Context()
	from, to, _ := parseRange(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	rows, err := h.Metrics.AIRag(ctx, h.Store.Admin, from, to, limit)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *OpsHandler) AIQuotaTrend(c *gin.Context) {
	ctx := c.Request.Context()
	rows, err := h.Metrics.AIQuotaTrend(ctx, h.Store.Admin)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}
