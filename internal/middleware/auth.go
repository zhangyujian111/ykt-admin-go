package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"ykt.dev/admin/internal/auth"
	"ykt.dev/admin/internal/common"
)

const (
	CtxUserID   = "uid"
	CtxUsername = "username"
	CtxTenantID = "tid"
	CtxRoles    = "roles"
)

// Auth 登录态校验（替换 sa-token 的 SaInterceptor）。
func Auth(jwtMgr *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从 header 拿 tokenName + tokenValue（兼容 sa-token 协议）
		tokName := c.GetHeader("tokenName")
		if tokName == "" {
			tokName = "satoken"
		}
		token := c.GetHeader(tokName)
		if token == "" {
			token = c.GetHeader("Authorization")
			token = strings.TrimPrefix(token, "Bearer ")
		}
		if token == "" {
			c.JSON(http.StatusUnauthorized, common.Fail[any](common.CodeUnauthorized, "未登录"))
			c.Abort()
			return
		}
		claims, err := jwtMgr.Parse(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, common.Fail[any](common.CodeTokenInvalid, err.Error()))
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUsername, claims.Username)
		c.Set(CtxTenantID, claims.TenantID)
		c.Set(CtxRoles, claims.Roles)
		c.Next()
	}
}
