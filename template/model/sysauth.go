package model

// LoginRequest 登录请求参数
type LoginRequest struct {
	Code     string `json:"username" example:"admin" binding:"required"`  // 用户名
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
