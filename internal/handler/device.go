package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"ykt.dev/admin/internal/common"
	"ykt.dev/admin/internal/repo"
)

// DeviceHandler /api/device/*  设备列表 + 分布 + 重启/推送（推送打 xiaozhi HTTP，stub）。
type DeviceHandler struct {
	Devices *repo.DeviceRepo
}

func (h *DeviceHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	keyword := c.Query("keyword")
	state := c.Query("state")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	rows, total, err := h.Devices.List(ctx, keyword, state, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(common.Page(total, page, pageSize, rows)))
}

func (h *DeviceHandler) VersionDistribution(c *gin.Context) {
	ctx := c.Request.Context()
	rows, err := h.Devices.VersionDistribution(ctx)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

func (h *DeviceHandler) TypeDistribution(c *gin.Context) {
	ctx := c.Request.Context()
	rows, err := h.Devices.TypeDistribution(ctx)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(rows))
}

// Restart stub — 真实实现需要调 xiaozhi-server HTTP API。
func (h *DeviceHandler) Restart(c *gin.Context) {
	deviceID := c.Param("deviceId")
	c.JSON(http.StatusOK, common.Ok(map[string]any{
		"deviceId": deviceID,
		"action":   "restart",
		"status":   "queued",
		"message":  "TODO: call xiaozhi-server /api/device/:id/restart",
	}))
}

// Push stub — 真实实现需要调 xiaozhi-server HTTP API。
func (h *DeviceHandler) Push(c *gin.Context) {
	deviceID := c.Param("deviceId")
	var body struct {
		Content string `json:"content"`
		Type    string `json:"type"`
	}
	_ = c.ShouldBind(&body)
	c.JSON(http.StatusOK, common.Ok(map[string]any{
		"deviceId": deviceID,
		"content":  body.Content,
		"type":     body.Type,
		"status":   "queued",
		"message":  "TODO: call xiaozhi-server push",
	}))
}
