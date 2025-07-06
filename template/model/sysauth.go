package model

// LoginRequest 登录请求参数
type LoginRequest struct {
	Code     string `json:"username" example:"admin" binding:"required"`  // 用户名
	Password string `json:"password" example:"123456" binding:"required"` // 密码(SM2加密)
}
 