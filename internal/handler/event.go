package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"ykt.dev/admin/internal/common"
	"ykt.dev/admin/internal/db"
)

// EventHandler /api/events  事件接入（xiaozhi-server 调用，写入 ykt_admin 库）。
type EventHandler struct {
	Store *db.Store
}

type EventItem struct {
	EventType  string          `json:"eventType"`  // CONNECT / DISCONNECT / DIALOG_FAIL / ACTION_FAIL / OFFLINE
	DeviceID   string          `json:"deviceId"`
	Stage      string          `json:"stage,omitempty"`
	ActionName string          `json:"actionName,omitempty"`
	Reason     string          `json:"reason,omitempty"`
	ErrorMsg   string          `json:"errorMsg,omitempty"`
	DurationMs int             `json:"durationMs,omitempty"`
	OccurredAt string          `json:"occurredAt"` // ISO8601
	Extra      json.RawMessage `json:"extra,omitempty"`
}

func (h *EventHandler) Ingest(c *gin.Context) {
	var e EventItem
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeBadRequest, "事件格式错误"))
		return
	}
	ctx := c.Request.Context()

	switch e.EventType {
	case "CONNECT", "DISCONNECT":
		ts := e.OccurredAt
		if ts == "" {
			ts = time.Now().Format("2006-01-02 15:04:05")
		}
		_, err := h.Store.Admin.ExecContext(ctx, `
			INSERT INTO ykt_event_connection (deviceId, reason, connectTime, disconnectTime, durationSec)
			VALUES (?, ?, ?, ?, 0)`,
			e.DeviceID, e.Reason, ts, ts)
		if err != nil {
			c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
			return
		}
	case "DIALOG_FAIL":
		_, err := h.Store.Admin.ExecContext(ctx, `
			INSERT INTO ykt_metrics_device_daily (deviceId, statDate, dialogFailedTurns, sttFailedCount, llmFailedCount, ttsFailedCount, lastFailureReason, lastFailureTime, hasProblem)
			VALUES (?, COALESCE(NULLIF(?, ''), CURDATE()), 1, 0, 0, 0, ?, COALESCE(NULLIF(?, ''), NOW()), 1)`,
			e.DeviceID, e.OccurredAt, e.ErrorMsg, e.OccurredAt)
		if err != nil {
			c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
			return
		}
	case "ACTION_FAIL":
		_, err := h.Store.Admin.ExecContext(ctx, `
			INSERT INTO ykt_metrics_device_daily (deviceId, statDate, actionFailedCount, actionDispatchCount, actionFailureRate, lastActionFailureTime, hasProblem)
			VALUES (?, COALESCE(NULLIF(?, ''), CURDATE()), 1, 1, 1.0, COALESCE(NULLIF(?, ''), NOW()), 1)`,
			e.DeviceID, e.OccurredAt, e.OccurredAt)
		if err != nil {
			c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
			return
		}
	case "OFFLINE":
		// OFFLINE 事件：当前表是 daily 聚合，写入 offlineDurationSec=1 表示一次离线事件
		_, err := h.Store.Admin.ExecContext(ctx, `
			INSERT INTO ykt_metrics_device_daily (deviceId, statDate, offlineDurationSec, offlineEventCount, hasProblem)
			VALUES (?, COALESCE(NULLIF(?, ''), CURDATE()), 0, 1, 1)
			ON DUPLICATE KEY UPDATE offlineEventCount = offlineEventCount + 1`,
			e.DeviceID, e.OccurredAt)
		if err != nil {
			c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
			return
		}
	default:
		c.JSON(http.StatusOK, common.Fail[any](common.CodeBadRequest, "未知 eventType: "+e.EventType))
		return
	}
	c.JSON(http.StatusOK, common.OkMsg[any]("event accepted", nil))
}

func (h *EventHandler) Batch(c *gin.Context) {
	var batch struct {
		Events []EventItem `json:"events"`
	}
	if err := c.ShouldBindJSON(&batch); err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeBadRequest, "events 格式错误"))
		return
	}
	ctx := c.Request.Context()
	accepted := 0
	for _, e := range batch.Events {
		ok := true
		switch e.EventType {
		case "CONNECT", "DISCONNECT":
			_, err := h.Store.Admin.ExecContext(ctx, `
				INSERT INTO ykt_event_connection (deviceId, reason, connectTime)
				VALUES (?, ?, COALESCE(NULLIF(?, ''), NOW()))`,
				e.DeviceID, e.Reason, e.OccurredAt)
			ok = err == nil
		case "DIALOG_FAIL":
			_, err := h.Store.Admin.ExecContext(ctx, `
				INSERT INTO ykt_metrics_device_daily (deviceId, statDate, dialogFailedTurns, sttFailedCount, llmFailedCount, ttsFailedCount, lastFailureReason, lastFailureTime, hasProblem)
				VALUES (?, COALESCE(NULLIF(?, ''), CURDATE()), 1, 0, 0, 0, ?, COALESCE(NULLIF(?, ''), NOW()), 1)`,
				e.DeviceID, e.OccurredAt, e.ErrorMsg, e.OccurredAt)
			ok = err == nil
		case "ACTION_FAIL":
			_, err := h.Store.Admin.ExecContext(ctx, `
				INSERT INTO ykt_metrics_device_daily (deviceId, statDate, actionFailedCount, actionDispatchCount, actionFailureRate, lastActionFailureTime, hasProblem)
				VALUES (?, COALESCE(NULLIF(?, ''), CURDATE()), 1, 1, 1.0, COALESCE(NULLIF(?, ''), NOW()), 1)`,
				e.DeviceID, e.OccurredAt, e.OccurredAt)
			ok = err == nil
		case "OFFLINE":
			_, err := h.Store.Admin.ExecContext(ctx, `
				INSERT INTO ykt_metrics_device_daily (deviceId, statDate, offlineDurationSec, offlineEventCount, hasProblem)
				VALUES (?, COALESCE(NULLIF(?, ''), CURDATE()), 0, 1, 1)
				ON DUPLICATE KEY UPDATE offlineEventCount = offlineEventCount + 1`,
				e.DeviceID, e.OccurredAt)
			ok = err == nil
		}
		if ok {
			accepted++
		}
	}
	c.JSON(http.StatusOK, common.Ok(map[string]any{"accepted": accepted, "total": len(batch.Events)}))
}
