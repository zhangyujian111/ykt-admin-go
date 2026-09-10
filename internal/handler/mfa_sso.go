package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ykt.dev/admin/internal/common"
)

// MfaHandler /api/mfa/status  MFA 状态（未实现 MFA）。
type MfaHandler struct{}

func (h *MfaHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, common.Ok(map[string]any{
		"enabled":      false,
		"method":       "",
		"requiredNext": false,
	}))
}

// SsoHandler /api/sso/providers  SSO 提供商（未接入）。
type SsoHandler struct{}

func (h *SsoHandler) Providers(c *gin.Context) {
	c.JSON(http.StatusOK, common.Ok(map[string]any{
		"providers": []any{},
	}))
}
