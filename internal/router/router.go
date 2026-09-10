package router

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"ykt.dev/admin/internal/auth"
	"ykt.dev/admin/internal/db"
	"ykt.dev/admin/internal/handler"
	"ykt.dev/admin/internal/middleware"
	"ykt.dev/admin/internal/repo"
)

// Deps 路由依赖。
type Deps struct {
	Store   *db.Store
	JWT     *auth.Manager
	Users   *repo.UserRepo
	Tenants *repo.TenantRepo
	Alerts  *repo.AlertRepo
	Devices *repo.DeviceRepo
	Metrics *repo.MetricsRepo
	System  *handler.SystemHandler
}

// New 构造 gin 引擎并注册全部路由。
func New(d Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "satoken", "Authorization", "tokenName"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// 公开
	authH := &handler.AuthHandler{Users: d.Users, Tenants: d.Tenants, JWT: d.JWT}
	dashH := &handler.DashboardHandler{Store: d.Store, Metrics: d.Metrics}
	opsH := &handler.OpsHandler{Store: d.Store, Metrics: d.Metrics}
	deviceH := &handler.DeviceHandler{Devices: d.Devices}
	alertH := &handler.AlertHandler{Alerts: d.Alerts}
	tenantH := &handler.TenantHandler{Tenants: d.Tenants}
	eventH := &handler.EventHandler{Store: d.Store}
	mfaH := &handler.MfaHandler{}
	ssoH := &handler.SsoHandler{}
	sysH := d.System

	// Health
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().Format(time.RFC3339)})
	})
	r.GET("/actuator/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
			"checks": gin.H{"admin": "ok"},
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 公开接口
	r.POST("/api/login", authH.Login)
	r.GET("/api/sso/providers", ssoH.Providers)

	// 事件接入（内部调用，xiaozhi-server 用，无 token 校验；生产应加内部 token）
	r.POST("/api/events", eventH.Ingest)
	r.POST("/api/events/batch", eventH.Batch)

	// 需要登录
	authed := r.Group("/")
	authed.Use(middleware.Auth(d.JWT))
	{
		authed.GET("/api/auth/oauth/me", authH.Me)
		authed.GET("/api/mfa/status", mfaH.Status)

		// Dashboard
		authed.GET("/api/dashboard/overview", dashH.Overview)
		authed.GET("/api/dashboard/alert-trend", dashH.AlertTrend)
		authed.GET("/api/dashboard/device-trend", dashH.DeviceTrend)

		// 运维大盘
		authed.GET("/api/ops/dashboard/summary", opsH.DashboardSummary)
		authed.GET("/api/ops/dashboard/problem-devices", opsH.ProblemDevices)
		authed.GET("/api/ops/dashboard/error-type-stats", opsH.ErrorTypeStats)

		// 5 大维度 tab
		authed.GET("/api/ops/devices-overview/connection", opsH.DevicesOverviewConnection)
		authed.GET("/api/ops/devices-overview/session", opsH.DevicesOverviewSession)
		authed.GET("/api/ops/devices-overview/latency", opsH.DevicesOverviewLatency)
		authed.GET("/api/ops/devices-overview/action", opsH.DevicesOverviewAction)
		authed.GET("/api/ops/devices-overview/detail/:deviceId", opsH.DeviceDetail)

		// 5 大维度独立接口
		authed.GET("/api/ops/connection/summary", opsH.ConnectionSummary)
		authed.GET("/api/ops/connection/abnormal-devices", opsH.ConnectionAbnormalDevices)
		authed.GET("/api/ops/connection/events", opsH.ConnectionEvents)
		authed.GET("/api/ops/dialog/summary", opsH.DialogSummary)
		authed.GET("/api/ops/dialog/abnormal-devices", opsH.DialogAbnormalDevices)
		authed.GET("/api/ops/dialog/failures", opsH.DialogFailures)
		authed.GET("/api/ops/action/summary", opsH.ActionSummary)
		authed.GET("/api/ops/action/failures", opsH.ActionFailures)
		authed.GET("/api/ops/latency/summary", opsH.LatencySummary)
		authed.GET("/api/ops/latency/trend", opsH.LatencyTrend)
		authed.GET("/api/ops/offline/devices", opsH.OfflineDevices)
		authed.GET("/api/ops/offline/history", opsH.OfflineHistory)
		authed.GET("/api/ops/problem-devices", opsH.ProblemDevices)

		// AI 监控
		authed.GET("/api/ops/ai-metrics/llm", opsH.AILlm)
		authed.GET("/api/ops/ai-metrics/tts", opsH.AITts)
		authed.GET("/api/ops/ai-metrics/asr", opsH.AIAsr)
		authed.GET("/api/ops/ai-metrics/rag", opsH.AIRag)
		authed.GET("/api/ops/ai-metrics/quota-trend", opsH.AIQuotaTrend)

		// 设备管理
		authed.GET("/api/device", deviceH.List)
		authed.GET("/api/device/distribution/version", deviceH.VersionDistribution)
		authed.GET("/api/device/distribution/type", deviceH.TypeDistribution)
		authed.POST("/api/device/restart/:deviceId", deviceH.Restart)
		authed.POST("/api/device/push/:deviceId", deviceH.Push)
		authed.GET("/api/device/detail/:deviceId", func(c *gin.Context) {
			// 重用 devices-overview/detail
			opsH.DeviceDetail(c)
		})

		// 告警
		authed.GET("/api/alert/rule", alertH.ListRules)
		authed.GET("/api/alert/history", alertH.ListHistory)
		authed.GET("/api/alert/active", alertH.Active)
		authed.POST("/api/alert/history/:alertId/ack", alertH.Ack)

		// 租户
		authed.GET("/api/tenant/current", tenantH.Current)
		authed.GET("/api/tenant/list", tenantH.List)
		authed.POST("/api/tenant/switch/:tenantId", tenantH.Switch)

		// 系统管理（user/role/menu/dept/dict/log）— 2026-09 从 muliti-toy 迁入
		if sysH != nil {
			authed.GET("/api/system/user/list", sysH.ListUsers)
			authed.GET("/api/system/user/:userId", sysH.GetUser)
			authed.POST("/api/system/user", sysH.CreateUser)
			authed.PUT("/api/system/user/:userId", sysH.UpdateUser)
			authed.DELETE("/api/system/user/:userId", sysH.DeleteUser)
			authed.PUT("/api/system/user/:userId/resetPwd", sysH.ResetPwd)
			authed.PUT("/api/system/user/:userId/status", sysH.ChangeStatus)

			authed.GET("/api/system/role/list", sysH.ListRoles)
			authed.POST("/api/system/role", sysH.CreateRole)
			authed.PUT("/api/system/role/:roleId", sysH.UpdateRole)
			authed.DELETE("/api/system/role/:roleId", sysH.DeleteRole)

			authed.GET("/api/system/menu/list", sysH.ListMenus)
			authed.GET("/api/system/dept/list", sysH.ListDepts)

			authed.GET("/api/system/dict/type/list", sysH.ListDictTypes)
			authed.GET("/api/system/dict/data/list", sysH.ListDictData)

			authed.GET("/api/system/log/oper", sysH.ListOperLog)
			authed.GET("/api/system/log/login", sysH.ListLoginLog)
		}
	}

	return r
}
