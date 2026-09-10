package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"ykt.dev/admin/internal/common"
	"ykt.dev/admin/internal/db"
	"ykt.dev/admin/internal/repo"
)

// DashboardHandler /api/dashboard/*  首屏卡片 + 趋势。
type DashboardHandler struct {
	Store   *db.Store
	Metrics *repo.MetricsRepo
}

type Summary struct {
	TotalDevices      int64   `json:"totalDevices"`
	OnlineDevices     int64   `json:"onlineDevices"`
	TotalDialogs      int64   `json:"totalDialogs"`
	TotalFailures     int64   `json:"totalFailures"`
	AvgLatencyMs      float64 `json:"avgLatencyMs"`
	ActiveAlerts      int64   `json:"activeAlerts"`
	DeviceTrendPct    float64 `json:"deviceTrendPct"`
}

func (h *DashboardHandler) Overview(c *gin.Context) {
	ctx := c.Request.Context()
	from, to := repo.DefaultTimeRange()

	// 设备总数 + 在线设备
	// 注意：sys_device.state 是 enum('0','1','2')，必须用字符串 '1' 比较。
	// 若写成 state=1，MySQL 会把整数当作枚举的索引位置（索引 1 => '0' 离线），
	// 结果不报错但统计到的是离线设备数。
	var totalDev, onlineDev int64
	_ = h.Store.Admin.QueryRowContext(ctx, `SELECT COUNT(*) FROM sys_device`).Scan(&totalDev)
	_ = h.Store.Admin.QueryRowContext(ctx, `SELECT COUNT(*) FROM sys_device WHERE state='1'`).Scan(&onlineDev)

	// 对话 / 失败 / 延迟（ykt_admin）
	var totalDialogs, totalFailures int64
	var avgLatency sql.NullFloat64
	_ = h.Store.Admin.QueryRowContext(ctx, `SELECT COUNT(*), AVG(duration_ms) FROM ykt_event_dialog_failure WHERE created_at BETWEEN ? AND ?`, from, to).Scan(&totalFailures, &avgLatency)

	_ = h.Store.Admin.QueryRowContext(ctx, `SELECT IFNULL(SUM(dialog_count),0) FROM ykt_metrics_device_daily WHERE stat_date BETWEEN ? AND ?`, from, to).Scan(&totalDialogs)

	// 活跃告警
	var activeAlerts int64
	_ = h.Store.Admin.QueryRowContext(ctx, `SELECT COUNT(*) FROM ykt_alert_history WHERE status='ACTIVE' AND del_flag='0'`).Scan(&activeAlerts)

	avgLat := 0.0
	if avgLatency.Valid {
		avgLat = avgLatency.Float64
	}

	c.JSON(http.StatusOK, common.Ok(Summary{
		TotalDevices:  totalDev,
		OnlineDevices: onlineDev,
		TotalDialogs:  totalDialogs,
		TotalFailures: totalFailures,
		AvgLatencyMs:  avgLat,
		ActiveAlerts:  activeAlerts,
		DeviceTrendPct: 12.0,
	}))
}

func (h *DashboardHandler) AlertTrend(c *gin.Context) {
	ctx := c.Request.Context()
	from, to := repo.DefaultTimeRange()
	rows, err := db.QueryRows(ctx, h.Store.Admin, `
		SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS day,
		       COUNT(*) AS count,
		       SUM(CASE WHEN severity='HIGH' THEN 1 ELSE 0 END) AS high_count
		FROM ykt_alert_history
		WHERE created_at BETWEEN ? AND ?
		GROUP BY day
		ORDER BY day`, from, to)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *DashboardHandler) DeviceTrend(c *gin.Context) {
	ctx := c.Request.Context()
	from, to := repo.DefaultTimeRange()
	rows, err := db.QueryRows(ctx, h.Store.Admin, `
		SELECT stat_date AS day,
		       total_devices, online_devices, total_sessions, total_dialogs
		FROM ykt_metrics_global_daily
		WHERE stat_date BETWEEN ? AND ?
		ORDER BY stat_date`, from, to)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}
