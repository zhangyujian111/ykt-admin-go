package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"ykt.dev/admin/internal/common"
	"ykt.dev/admin/internal/repo"
)

// AlertHandler /api/alert/*  告警规则 + 历史 + 活跃告警 + 确认。
type AlertHandler struct {
	Alerts *repo.AlertRepo
}

func (h *AlertHandler) ListRules(c *gin.Context) {
	ctx := c.Request.Context()
	rules, err := h.Alerts.ListRules(ctx)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rules))
}

func (h *AlertHandler) ListHistory(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	rows, total, err := h.Alerts.ListHistory(ctx, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(common.Page(total, page, pageSize, rows)))
}

func (h *AlertHandler) Active(c *gin.Context) {
	ctx := c.Request.Context()
	rows, err := h.Alerts.ActiveAlerts(ctx)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *AlertHandler) Ack(c *gin.Context) {
	idStr := c.Param("alertId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeBadRequest, "alertId 非法"))
		return
	}
	if err := h.Alerts.Ack(c.Request.Context(), id, 1); err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OkMsg[any]("确认成功", nil))
}
