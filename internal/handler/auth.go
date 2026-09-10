package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ykt.dev/admin/internal/auth"
	"ykt.dev/admin/internal/common"
	"ykt.dev/admin/internal/repo"
)

// AuthHandler /api/login + /api/auth/oauth/me。
type AuthHandler struct {
	Users    *repo.UserRepo
	Tenants  *repo.TenantRepo
	JWT      *auth.Manager
}

type LoginReq struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

type LoginResp struct {
	TokenName string `json:"tokenName"`
	Token     string `json:"token"`
	UserID    int64  `json:"userId"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	TenantID  int64  `json:"tenantId"`
	Roles     []string `json:"roles"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginReq
	_ = c.ShouldBind(&req)
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeBadRequest, "账号或密码不能为空"))
		return
	}

	ctx := c.Request.Context()
	u, err := h.Users.FindByUsername(ctx, req.Username)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}
	if u == nil || !h.Users.VerifyPassword(u, req.Password) {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeUnauthorized, "账号或密码错误"))
		return
	}
	if u.Status == "0" {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeForbidden, "账号已被停用"))
		return
	}

	tenants, _ := h.Tenants.List(ctx)
	tenantID := int64(0)
	if len(tenants) > 0 {
		tenantID = tenants[0].ID
	}
	if tenantID == 0 {
		// 默认租户（ykt_sys_tenant 表不存在时 fallback）
		tenantID = 1
	}
	roles := []string{"admin"}

	tok, err := h.JWT.Issue(u.UserID, u.Username, tenantID, roles)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail[any](common.CodeInternal, err.Error()))
		return
	}

	nickname := u.Nickname.String
	if nickname == "" {
		nickname = u.Username
	}
	c.JSON(http.StatusOK, common.OkMsg("登录成功", LoginResp{
		TokenName: "satoken",
		Token:     tok,
		UserID:    u.UserID,
		Username:  u.Username,
		Nickname:  nickname,
		TenantID:  tenantID,
		Roles:     roles,
	}))
}

func (h *AuthHandler) Me(c *gin.Context) {
	uid, _ := c.Get("uid")
	uname, _ := c.Get("username")
	tid, _ := c.Get("tid")
	roles, _ := c.Get("roles")

	c.JSON(http.StatusOK, common.Ok(map[string]any{
		"userId":   uid,
		"username": uname,
		"tenantId": tid,
		"roles":    roles,
		"nickname": uname,
	}))
}
