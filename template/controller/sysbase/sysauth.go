package sysbase

import (
	"encoding/json"
	"{project_name}/model"
	"{project_name}/public"
	"time"

	"github.com/guyigood/gyweb/core/middleware"

	"strings"

	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/utils/common"
	"github.com/guyigood/gyweb/core/utils/datatype"
)

/**
 * 登录验证
 */
func CheckAuth(c *gyarn.Context) bool {
	loginKey := "login"
	//使用Bearer Token认证方式
	token := GetToken(c)
	middleware.DebugVar("Token", token)
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

func GetToken(c *gyarn.Context) string {
	token := c.GetHeader("Authorization")
	if token == "" {
		token = c.GetHeader("Token")
	}
	token = strings.TrimPrefix(token, "Bearer ")
	return token
}

// Login 用户登
// Login 用户登录
// @Summary 用户登录
// @Description 用户登录接口，使用SM2解密密码并验证用户身份
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param loginRequest body model.LoginRequest true "登录参数"
// @Success 200 {object} map[string]interface{} "登录成功，返回token"
// @Failure 101 {object} map[string]interface{} "解密失败或其他错误"
// @Failure 102 {object} map[string]interface{} "无效的请求参数"
// @Failure 103 {object} map[string]interface{} "用户名或密码错误"
// @Router /api/auth/login [post]
func Login(c *gyarn.Context) {
	db := public.GetDbConnection()
	defer db.Close()
	user := new(model.LoginUser)
	var loginForm model.LoginRequest
	err := c.BindJSON(&loginForm)
	if err != nil {
		c.Error(102, "无效的请求参数")
		return
	}
	db.GetDB().Exec("update login set pass=md5('123456) where pass='123456'", nil)
	/*
		loginpass, _ := public.Sm2Encrpt("123456")
		middleware.DebugVar("loginpass", loginpass)
		password, err2 := public.Sm2Decrypt(loginForm.Password)
		if err2 != nil {
			c.Error(101, err2.Error())
			return
		}
		pass := public.Sm3Hash(password)
		middleware.DebugVar("pass:", pass+","+password)
	*/
	data, err1 := db.Table("login").Where("status=1 and code=? and pass=?", loginForm.Code, loginForm.Password).Get()
	if err1 != nil {
		c.Error(101, "登录失败,用户名或密码错误")
		return
	}
	if data == nil {
		c.Error(103, "用户名或密码错误")
		return
	}
	user.ID, _ = datatype.TypetoInt(data["id"])
	user.Username = datatype.TypetoStr(data["code"])
	user.RoleId, _ = datatype.TypetoInt(data["role_id"])
	role, _ := db.Table("role").Where("id=?", data["role_id"]).Get()
	user.RoleName = datatype.TypetoStr(role["name"])
	user.Memo = datatype.TypetoStr(role["memo"])
	user.Roles = datatype.TypetoStr(role["memo"])
	user_info, _ := json.Marshal(user)
	uuid := common.GetUUID()

	err = public.Re_Client.Set(uuid, string(user_info), 3600*time.Second)
	if err != nil {
		c.Error(101, "登录失败,缓存失败")
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
// @Success 200 {object} map[string]interface{} "用户信息"
// @Failure 101 {object} map[string]interface{} "解析用户信息失败"
// @Failure 401 {object} map[string]interface{} "未授权，请先登录"
// @Router /api/auth/userinfo [get]
func UserInfo(c *gyarn.Context) {
	uuid := GetToken(c)
	if uuid == "" {
		c.Error(101, "请先登录")
		return
	}
	user_info, err1 := public.Re_Client.Get(uuid)
	if err1 != nil {
		c.Error(101, err1.Error())
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
// @Success 200 {object} map[string]interface{} "菜单树形结构"
// @Failure 101 {object} map[string]interface{} "用户信息错误"
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
	db := public.GetDbConnection()
	defer db.Close()
	convertedData := make([]map[string]interface{}, 0)
	if role == "0" || role == "" {
		menu, _ := db.Table("nav_menu").Where("is_del=0").OrderBy("order_number").All()
		//fmt.Println(menu)
		for _, m := range menu {
			convertedData = append(convertedData, m) // 直接赋值，因为 MapModel 底层是 map[string]interface{}
		}
	} else {
		menu, _ := db.Table("nav_menu").Where("is_del=0 and id in (" + user.Memo + ")").OrderBy("order_number").All()
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
// @Success 200 {object} map[string]interface{} "退出成功"
// @Failure 101 {object} map[string]interface{} "请先登录"
// @Router /api/auth/logout [post]
func Logout(c *gyarn.Context) {
	token := GetToken(c)
	if token == "" {
		c.Error(101, "请先登录")
		return
	}
	public.Re_Client.Del(token)
	c.Success("退出成功")
}

func CheckUser(c *gyarn.Context) {
	_, ok := c.Get("login")
	if !ok {
		c.Error(101, "用户信息错误")
		return
	}
	c.Success("ok")

}
