package sysbase

import (
	"encoding/json"
	"net/http"
	"time"
	"{project_name}/model"
	"{project_name}/public"

	"github.com/guyigood/gyweb/core/middleware"

	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/utils/common"
	"github.com/guyigood/gyweb/core/utils/datatype"
)

/**
 * 登录验证
 */
func CheckAuth(c *gyarn.Context) bool {
	loginKey := "login"
	token := c.GetHeader("Token")

	user := new(model.LoginUser)
	loginFlag, err := public.Re_Client.Exists(token)
	if err != nil {
		return false
	}
	if loginFlag {
		userinfo, err1 := public.Re_Client.Get(token)
		if err1 != nil {
			return false
		}
		err = json.Unmarshal([]byte(userinfo), &user)
		if err != nil {
			return false
		}
		public.Re_Client.Expire(token, 3600*time.Second)
		c.Set(loginKey, *user)
		return true
	}

	return false
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Code     string `json:"code" example:"admin" binding:"required"`      // 用户名
	Password string `json:"password" example:"123456" binding:"required"` // 密码(SM2加密)
}

// LoginResponse 登录响应
type LoginResponse struct {
	Code int    `json:"code" example:"200"`        // 响应码
	Msg  string `json:"msg" example:"操作成功"`        // 响应消息
	Data string `json:"data" example:"token-uuid"` // 返回的token
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	Code int             `json:"code" example:"200"` // 响应码
	Msg  string          `json:"msg" example:"操作成功"` // 响应消息
	Data LoginUserDetail `json:"data"`               // 用户详细信息
}

// LoginUserDetail 用户详细信息
type LoginUserDetail struct {
	ID       int    `json:"id" example:"1"`           // 用户ID
	Username string `json:"username" example:"admin"` // 用户名
	RoleId   int    `json:"role_id" example:"1"`      // 角色ID
	RoleName string `json:"role_name" example:"管理员"`  // 角色名称
	Memo     string `json:"memo" example:"权限备注"`      // 权限备注
}

// MenuResponse 菜单响应
type MenuResponse struct {
	Code int                      `json:"code" example:"200"` // 响应码
	Msg  string                   `json:"msg" example:"操作成功"` // 响应消息
	Data []map[string]interface{} `json:"data"`               // 菜单树形结构
}

// LogoutResponse 登出响应
type LogoutResponse struct {
	Code int    `json:"code" example:"200"`  // 响应码
	Msg  string `json:"msg" example:"操作成功"`  // 响应消息
	Data string `json:"data" example:"退出成功"` // 响应数据
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code int    `json:"code" example:"101"`  // 错误码
	Msg  string `json:"msg" example:"错误信息"`  // 错误消息
	Data string `json:"data" example:"null"` // 错误数据
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录接口，使用SM2解密密码并验证用户身份
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "登录参数"
// @Success 200 {object} LoginResponse "登录成功，返回token"
// @Failure 101 {object} ErrorResponse "解密失败或其他错误"
// @Failure 102 {object} ErrorResponse "无效的请求参数"
// @Failure 103 {object} ErrorResponse "用户名或密码错误"
// @Router /api/auth/login [post]
func Login(c *gyarn.Context) {
	db := public.Db
	user := new(model.LoginUser)
	var loginForm LoginRequest
	err := c.BindJSON(&loginForm)
	if err != nil {
		c.Error(102, "无效的请求参数")
		return
	}
	loginpass, _ := public.Sm2Encrpt("123456")
	middleware.DebugVar("loginpass", loginpass)
	password, err2 := public.Sm2Decrypt(loginForm.Password)
	if err2 != nil {
		c.Error(101, err2.Error())
		return
	}
	data, err1 := db.Table("sl_login").Where("status=1 and code=? and pass=?", loginForm.Code, public.Sm3Hash(password)).Get()
	if err1 != nil {
		c.Error(101, err1.Error())
		return
	}
	if data == nil {
		c.Error(103, "用户名或密码错误")
		return
	}
	user.ID, _ = datatype.TypetoInt(data["id"])
	user.Username = datatype.TypetoStr(data["code"])
	user.RoleId, _ = datatype.TypetoInt(data["role_id"])
	role, _ := db.Table("sl_role").Where("id=?", data["role_id"]).Get()
	user.RoleName = datatype.TypetoStr(role["name"])
	user.Memo = datatype.TypetoStr(role["memo"])
	user_info, _ := json.Marshal(user)
	uuid := common.GetUUID()

	err = public.Re_Client.Set(uuid, string(user_info), 3600*time.Second)
	if err != nil {
		c.Error(101, err2.Error())
		return
	}
	c.Set("login", *user)
	c.Success(uuid)
}

/**
 * 获取用户信息
 */
// UserInfo 获取用户信息
// @Summary 获取用户信息
// @Description 根据token获取当前登录用户的详细信息
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param Token header string true "用户token"
// @Success 200 {object} UserInfoResponse "用户信息"
// @Failure 101 {object} ErrorResponse "解析用户信息失败"
// @Failure 401 {object} ErrorResponse "未授权，请先登录"
// @Router /api/auth/userinfo [get]
func UserInfo(c *gyarn.Context) {
	uuid := c.GetHeader("Token")
	if uuid == "" {
		c.Fail(http.StatusUnauthorized, "请先登录")
		return
	}
	user_info, err1 := public.Re_Client.Get(uuid)
	if err1 != nil {
		c.Fail(http.StatusUnauthorized, err1.Error())
		return
	}

	user := new(model.LoginUser)
	err := json.Unmarshal([]byte(user_info), &user)
	if err != nil {
		c.Error(101, err.Error())
		return
	}
	c.Success(user)
}

/**
 * 获取用户权限菜单
 */
// GetRoleMenu 获取用户权限菜单
// @Summary 获取用户权限菜单
// @Description 根据用户角色获取相应的权限菜单，支持树形结构
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param Token header string true "用户token"
// @Param role query string false "角色ID，为空或0时返回全部菜单"
// @Success 200 {object} MenuResponse "菜单树形结构"
// @Failure 101 {object} ErrorResponse "用户信息错误"
// @Router /api/auth/getmenu [get]
func GetRoleMenu(c *gyarn.Context) {
	role := c.Query("role")
	login, ok := c.Get("login")
	if !ok {
		c.Error(101, "用户信息错误")
		return
	}
	user, u_flag := login.(model.LoginUser)
	if !u_flag {
		c.Error(101, "用户信息错误")
		return
	}
	db := public.Db
	convertedData := make([]map[string]interface{}, 0)
	if role == "0" || role == "" {
		menu, _ := db.Table("sl_nav").All()
		//fmt.Println(menu)
		for _, m := range menu {
			convertedData = append(convertedData, m) // 直接赋值，因为 MapModel 底层是 map[string]interface{}
		}
	} else {
		menu, _ := db.Table("sl_nav").Where("id in (" + user.Memo + ")").All()
		for _, m := range menu {
			convertedData = append(convertedData, m) // 直接赋值，因为 MapModel 底层是 map[string]interface{}
		}
	}
	c.Success(common.BuildTree(convertedData, "0", "id", "parent_id"))

}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出接口，清除用户token
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param Token header string true "用户token"
// @Success 200 {object} LogoutResponse "退出成功"
// @Failure 101 {object} ErrorResponse "请先登录"
// @Router /api/auth/logout [post]
func Logout(c *gyarn.Context) {
	token := c.GetHeader("Token")
	if token == "" {
		c.Error(101, "请先登录")
		return
	}
	public.Re_Client.Del(token)
	c.Success("退出成功")
}
