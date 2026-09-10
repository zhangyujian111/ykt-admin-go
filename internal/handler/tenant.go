package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"ykt.dev/admin/internal/common"
	"ykt.dev/admin/internal/middleware"
	"ykt.dev/admin/internal/repo"
)

// TenantHandler /api/tenant/*  当前租户 + 列表 + 切换。
type TenantHandler struct {
	Tenants *repo.TenantRepo
}

func (h *TenantHandler) Current(c *gin.Context) {
	tidAny, _ := c.Get(middleware.CtxTenantID)
	tid, _ := tidAny.(int64)
	t, err := h.Tenants.GetByID(c.Request.Context(), tid)
	if err != nil || t == nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeNotFound, "租户不存在"))
		return
	}
	c.JSON(http.StatusOK, common.Ok(t))
}

func (h *TenantHandler) List(c *gin.Context) {
	ts, err := h.Tenants.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Ok(ts))
}

// Switch 仅返回成功，前端需要重新登录或更新 token（这里给个新 token 占位）。
func (h *TenantHandler) Switch(c *gin.Context) {
	tidStr := c.Param("tenantId")
	tid, err := strconv.ParseInt(tidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeBadRequest, "tenantId 非法"))
		return
	}
	t, err := h.Tenants.GetByID(c.Request.Context(), tid)
	if err != nil || t == nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeNotFound, "租户不存在"))
		return
	}
	// 提示：完整切换需要重新签发 token；这里返回成功让前端知道切换生效
	c.JSON(http.StatusOK, common.Ok(map[string]any{
		"tenantId":   t.ID,
		"tenantName": t.Name,
		"message":    "切换成功，请重新登录以刷新 token",
	}))
}
