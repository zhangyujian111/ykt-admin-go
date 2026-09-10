package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"ykt.dev/admin/internal/auth"
	"ykt.dev/admin/internal/db"
)

// SystemHandler 系统管理（user/role/menu/dept/dict）— 2026-09 从 muliti-toy 迁入。
type SystemHandler struct {
	Store *db.Store
}

// ---- 用户 ----

func (h *SystemHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	username := strings.TrimSpace(c.Query("username"))
	nickname := strings.TrimSpace(c.Query("nickname"))
	status := strings.TrimSpace(c.Query("status"))
	if page < 1 { page = 1 }
	if size < 1 || size > 200 { size = 20 }

	args := []any{}
	where := "1=1"
	if username != "" { where += " AND username LIKE ?"; args = append(args, "%"+username+"%") }
	if nickname != "" { where += " AND nickname LIKE ?"; args = append(args, "%"+nickname+"%") }
	if status != "" { where += " AND status = ?"; args = append(args, status) }

	var total int
	if err := h.Store.Admin.QueryRowContext(c.Request.Context(),
		"SELECT COUNT(*) FROM ykt_sys_user WHERE "+where, args...).Scan(&total); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}

	args2 := append([]any{}, args...)
	args2 = append(args2, size, (page-1)*size)
	rows, err := db.QueryRows(c.Request.Context(), h.Store.Admin,
		`SELECT userId, deptId, username, nickname, email, phone, status, loginIp, loginDate, createTime
		 FROM ykt_sys_user WHERE `+where+` ORDER BY userId DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": rows, "total": total, "pageNum": page, "pageSize": size}})
}

func (h *SystemHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 40002, "message": "invalid userId"})
		return
	}
	rows, err := db.QueryRows(c.Request.Context(), h.Store.Admin,
		`SELECT userId, deptId, username, nickname, email, phone, status, avatar, createTime FROM ykt_sys_user WHERE userId = ?`, id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	if len(rows) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 40401, "message": "user not found"})
		return
	}
	// 角色 ID 列表
	roles, _ := db.QueryRows(c.Request.Context(), h.Store.Admin,
		`SELECT roleId FROM ykt_sys_user_role WHERE userId = ?`, id)
	roleIDs := []int64{}
	for _, r := range roles {
		if v, ok := r["roleId"].(int64); ok { roleIDs = append(roleIDs, v) }
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"user": rows[0], "roleIds": roleIDs}})
}

type userCUReq struct {
	DeptID    int64    `json:"deptId"`
	Username  string   `json:"username"`
	Nickname  string   `json:"nickname"`
	Password  string   `json:"password"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Status    string   `json:"status"`
	RoleIDs   []int64  `json:"roleIds"`
}

func (h *SystemHandler) CreateUser(c *gin.Context) {
	var req userCUReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 40002, "message": "invalid JSON: " + err.Error()})
		return
	}
	if req.Username == "" || req.Nickname == "" {
		c.JSON(http.StatusOK, gin.H{"code": 40003, "message": "username/nickname 必填"})
		return
	}
	if req.Password == "" { req.Password = "123456" }
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": "bcrypt: " + err.Error()})
		return
	}
	if req.Status == "" { req.Status = "1" }
	res, err := h.Store.Admin.ExecContext(c.Request.Context(),
		`INSERT INTO ykt_sys_user (deptId, username, nickname, password, email, phone, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		req.DeptID, req.Username, req.Nickname, string(hashed), req.Email, req.Phone, req.Status)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	uid, _ := res.LastInsertId()
	for _, rid := range req.RoleIDs {
		_, _ = h.Store.Admin.ExecContext(c.Request.Context(),
			`INSERT IGNORE INTO ykt_sys_user_role (userId, roleId) VALUES (?, ?)`, uid, rid)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"userId": uid}})
}

func (h *SystemHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 40002, "message": "invalid userId"})
		return
	}
	var req userCUReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 40002, "message": "invalid JSON: " + err.Error()})
		return
	}
	if _, err := h.Store.Admin.ExecContext(c.Request.Context(),
		`UPDATE ykt_sys_user SET deptId=?, nickname=?, email=?, phone=?, status=? WHERE userId=?`,
		req.DeptID, req.Nickname, req.Email, req.Phone, req.Status, id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	if req.RoleIDs != nil {
		_, _ = h.Store.Admin.ExecContext(c.Request.Context(), `DELETE FROM ykt_sys_user_role WHERE userId = ?`, id)
		for _, rid := range req.RoleIDs {
			_, _ = h.Store.Admin.ExecContext(c.Request.Context(),
				`INSERT IGNORE INTO ykt_sys_user_role (userId, roleId) VALUES (?, ?)`, id, rid)
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

func (h *SystemHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 40002, "message": "invalid userId"})
		return
	}
	if id == 1 {
		c.JSON(http.StatusOK, gin.H{"code": 40301, "message": "不允许删除超管"})
		return
	}
	if _, err := h.Store.Admin.ExecContext(c.Request.Context(),
		`DELETE FROM ykt_sys_user_role WHERE userId = ?`, id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	if _, err := h.Store.Admin.ExecContext(c.Request.Context(),
		`DELETE FROM ykt_sys_user WHERE userId = ?`, id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

func (h *SystemHandler) ResetPwd(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 40002, "message": "invalid userId"})
		return
	}
	var req struct{ Password string `json:"password"` }
	_ = c.ShouldBindJSON(&req)
	if req.Password == "" { req.Password = "123456" }
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if _, err := h.Store.Admin.ExecContext(c.Request.Context(),
		`UPDATE ykt_sys_user SET password = ?, pwdResetDate = NOW() WHERE userId = ?`, string(hashed), id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

func (h *SystemHandler) ChangeStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 40002, "message": "invalid userId"})
		return
	}
	var req struct{ Status string `json:"status"` }
	_ = c.ShouldBindJSON(&req)
	if req.Status == "" { req.Status = "1" }
	if _, err := h.Store.Admin.ExecContext(c.Request.Context(),
		`UPDATE ykt_sys_user SET status = ? WHERE userId = ?`, req.Status, id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// ---- 角色 ----

func (h *SystemHandler) ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	roleName := strings.TrimSpace(c.Query("roleName"))
	if page < 1 { page = 1 }
	if size < 1 || size > 200 { size = 20 }

	args := []any{}
	where := "1=1"
	if roleName != "" { where += " AND roleName LIKE ?"; args = append(args, "%"+roleName+"%") }
	var total int
	_ = h.Store.Admin.QueryRowContext(c.Request.Context(),
		"SELECT COUNT(*) FROM ykt_sys_role WHERE "+where, args...).Scan(&total)
	args2 := append([]any{}, args...)
	args2 = append(args2, size, (page-1)*size)
	rows, err := db.QueryRows(c.Request.Context(), h.Store.Admin,
		`SELECT roleId, roleName, roleKey, dataScope, status, createTime FROM ykt_sys_role WHERE `+where+` ORDER BY roleId DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": rows, "total": total}})
}

type roleCUReq struct {
	RoleName string `json:"roleName"`
	RoleKey  string `json:"roleKey"`
	DataScope string `json:"dataScope"`
	Status   string `json:"status"`
	MenuIDs  []int64 `json:"menuIds"`
}

func (h *SystemHandler) CreateRole(c *gin.Context) {
	var req roleCUReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 40002, "message": "invalid JSON"})
		return
	}
	if req.RoleName == "" || req.RoleKey == "" {
		c.JSON(http.StatusOK, gin.H{"code": 40003, "message": "roleName/roleKey 必填"})
		return
	}
	if req.Status == "" { req.Status = "1" }
	if req.DataScope == "" { req.DataScope = "1" }
	res, err := h.Store.Admin.ExecContext(c.Request.Context(),
		`INSERT INTO ykt_sys_role (roleName, roleKey, dataScope, status) VALUES (?, ?, ?, ?)`,
		req.RoleName, req.RoleKey, req.DataScope, req.Status)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	rid, _ := res.LastInsertId()
	for _, mid := range req.MenuIDs {
		_, _ = h.Store.Admin.ExecContext(c.Request.Context(),
			`INSERT IGNORE INTO ykt_sys_role_menu (roleId, menuId) VALUES (?, ?)`, rid, mid)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"roleId": rid}})
}

func (h *SystemHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("roleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 40002, "message": "invalid roleId"})
		return
	}
	var req roleCUReq
	_ = c.ShouldBindJSON(&req)
	if _, err := h.Store.Admin.ExecContext(c.Request.Context(),
		`UPDATE ykt_sys_role SET roleName=?, roleKey=?, dataScope=?, status=? WHERE roleId=?`,
		req.RoleName, req.RoleKey, req.DataScope, req.Status, id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	if req.MenuIDs != nil {
		_, _ = h.Store.Admin.ExecContext(c.Request.Context(), `DELETE FROM ykt_sys_role_menu WHERE roleId = ?`, id)
		for _, mid := range req.MenuIDs {
			_, _ = h.Store.Admin.ExecContext(c.Request.Context(),
				`INSERT IGNORE INTO ykt_sys_role_menu (roleId, menuId) VALUES (?, ?)`, id, mid)
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

func (h *SystemHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("roleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 40002, "message": "invalid roleId"})
		return
	}
	_, _ = h.Store.Admin.ExecContext(c.Request.Context(), `DELETE FROM ykt_sys_role_menu WHERE roleId = ?`, id)
	_, _ = h.Store.Admin.ExecContext(c.Request.Context(), `DELETE FROM ykt_sys_user_role WHERE roleId = ?`, id)
	if _, err := h.Store.Admin.ExecContext(c.Request.Context(),
		`DELETE FROM ykt_sys_role WHERE roleId = ?`, id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// ---- 菜单 ----

func (h *SystemHandler) ListMenus(c *gin.Context) {
	rows, err := db.QueryRows(c.Request.Context(), h.Store.Admin,
		`SELECT menuId, parentId, menuName, menuType, orderNum, path, component, perms, icon, visible, status
		 FROM ykt_sys_menu ORDER BY parentId, orderNum`)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rows})
}

// ---- 部门 ----

func (h *SystemHandler) ListDepts(c *gin.Context) {
	rows, err := db.QueryRows(c.Request.Context(), h.Store.Admin,
		`SELECT deptId, parentId, ancestors, deptName, orderNum, leader, phone, status
		 FROM ykt_sys_dept ORDER BY parentId, orderNum`)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rows})
}

// ---- 字典 ----

func (h *SystemHandler) ListDictTypes(c *gin.Context) {
	rows, err := db.QueryRows(c.Request.Context(), h.Store.Admin,
		`SELECT dictId, dictName, dictType, status, remark FROM ykt_sys_dict_type ORDER BY dictId DESC`)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rows})
}

func (h *SystemHandler) ListDictData(c *gin.Context) {
	dictType := c.Query("dictType")
	args := []any{}
	where := "1=1"
	if dictType != "" { where = "dictType = ?"; args = append(args, dictType) }
	rows, err := db.QueryRows(c.Request.Context(), h.Store.Admin,
		`SELECT dictCode, dictType, dictLabel, dictValue, dictSort, cssClass, listClass, isDefault, status
		 FROM ykt_sys_dict_data WHERE `+where+` ORDER BY dictSort`, args...)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rows})
}

// ---- 操作日志 ----

func (h *SystemHandler) ListOperLog(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 { page = 1 }
	if size < 1 || size > 200 { size = 20 }
	var total int
	_ = h.Store.Admin.QueryRowContext(c.Request.Context(),
		"SELECT COUNT(*) FROM ykt_sys_oper_log").Scan(&total)
	rows, err := db.QueryRows(c.Request.Context(), h.Store.Admin,
		`SELECT operId, title, businessType, method, requestMethod, operName, deptName, operUrl, operIp, operParam, jsonResult, status, errorMsg, operTime
		 FROM ykt_sys_oper_log ORDER BY operId DESC LIMIT ? OFFSET ?`, size, (page-1)*size)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": rows, "total": total}})
}

// ---- 登录日志 ----

func (h *SystemHandler) ListLoginLog(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 { page = 1 }
	if size < 1 || size > 200 { size = 20 }
	var total int
	_ = h.Store.Admin.QueryRowContext(c.Request.Context(),
		"SELECT COUNT(*) FROM ykt_sys_login_log").Scan(&total)
	rows, err := db.QueryRows(c.Request.Context(), h.Store.Admin,
		`SELECT infoId, loginName, ipaddr, loginLocation, browser, os, status, msg, loginTime
		 FROM ykt_sys_login_log ORDER BY infoId DESC LIMIT ? OFFSET ?`, size, (page-1)*size)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50001, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": rows, "total": total}})
}

// 引用保持
var _ = auth.New