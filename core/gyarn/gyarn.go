package gyarn

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"strings"
)

// HandlerFunc 定义处理函数类型
type HandlerFunc func(*Context)

// H 是一个便捷的 map 类型，用于 JSON 响应
type H map[string]interface{}

// Context 封装了请求和响应
type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	Path       string
	Method     string
	Params     map[string]string // 路由参数
	StatusCode int
	Handlers   []HandlerFunc
	index      int
	aborted    bool
	// 用于存储请求级别的数据
	Keys map[string]interface{}
}

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`           // 状态码
	Message string      `json:"message"`        // 提示信息
	Data    interface{} `json:"data,omitempty"` // 数据，可选
}

// NewContext 创建新的上下文
func NewContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer:     w,
		Request:    req,
		Path:       req.URL.Path,
		Method:     req.Method,
		Params:     make(map[string]string), // 初始化路由参数map
		StatusCode: http.StatusOK,           // 设置默认状态码
		Handlers:   make([]HandlerFunc, 0),  // 初始化处理器切片
		index:      -1,
		aborted:    false,
		Keys:       make(map[string]interface{}), // 初始化Keys map
	}
}

// Set 存储键值对
func (c *Context) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

// Get 获取存储的值
func (c *Context) Get(key string) (value interface{}, exists bool) {
	if c.Keys == nil {
		return nil, false
	}
	value, exists = c.Keys[key]
	return
}

// MustGet 获取存储的值，如果不存在则panic
func (c *Context) MustGet(key string) interface{} {
	if value, exists := c.Get(key); exists {
		return value
	}
	fmt.Println("Key \"" + key + "\" does not exist")
	return nil

}

// Next 执行下一个中间件
func (c *Context) Next() {
	c.index++
	s := len(c.Handlers)
	for ; c.index < s && !c.aborted; c.index++ {
		c.Handlers[c.index](c)
	}
}

// Abort 中止后续中间件的执行
func (c *Context) Abort() {
	c.aborted = true
}

// IsAborted 检查是否已中止
func (c *Context) IsAborted() bool {
	return c.aborted
}

// JSON 发送 JSON 响应
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// XML 发送 XML 响应
func (c *Context) XML(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/xml")
	c.Status(code)
	encoder := xml.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// String 发送字符串响应
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(format))
}

// Data 发送数据响应，支持设置Content-Type
func (c *Context) Data(code int, contentType string, data []byte) {
	if contentType != "" {
		c.SetHeader("Content-Type", contentType)
	}
	c.Status(code)
	c.Writer.Write(data)
}

// HTML 发送 HTML 响应
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

// Status 设置状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// SetHeader 设置响应头
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// Param 获取路由参数
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

// Query 获取查询参数
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// DefaultQuery 获取查询参数，如果不存在则返回默认值
func (c *Context) DefaultQuery(key, defaultValue string) string {
	if value := c.Query(key); value != "" {
		return value
	}
	return defaultValue
}

// PostForm 获取表单参数
func (c *Context) PostForm(key string) string {
	return c.Request.FormValue(key)
}

// Fail 返回错误响应
func (c *Context) Fail(code int, err string) {
	c.index = len(c.Handlers)
	c.JSON(code, H{"error": err})
}

// GetHeader 获取请求头
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// SetCookie 设置Cookie
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

// Cookie 获取Cookie
func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// BindJSON 绑定 JSON 数据到结构体
func (c *Context) BindJSON(obj interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(obj)
}

// ShouldBindJSON 绑定 JSON 数据到结构体（别名方法）
func (c *Context) ShouldBindJSON(obj interface{}) error {
	return c.BindJSON(obj)
}

// BindXML 绑定XML请求体
func (c *Context) BindXML(obj interface{}) error {
	decoder := xml.NewDecoder(c.Request.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}

// Success 成功响应
func (c *Context) Success(data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 带自定义消息的成功响应
func (c *Context) SuccessWithMessage(message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func (c *Context) Error(code int, message string) {
	if code == 0 {
		code = 500 // 默认错误码
	}
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// ErrorWithData 带数据的错误响应
func (c *Context) ErrorWithData(code int, message string, data interface{}) {
	if code == 0 {
		code = 500 // 默认错误码
	}
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// 预定义一些常用的错误码
const (
	ErrCodeSuccess        = 200 // 成功
	ErrCodeBadRequest     = 400 // 请求参数错误
	ErrCodeUnauthorized   = 401 // 未授权
	ErrCodeForbidden      = 403 // 禁止访问
	ErrCodeNotFound       = 404 // 资源不存在
	ErrCodeInternalServer = 500 // 服务器内部错误
)

// BadRequest 400错误响应
func (c *Context) BadRequest(message string) {
	if message == "" {
		message = "请求参数错误"
	}
	c.Error(ErrCodeBadRequest, message)
}

// Unauthorized 401错误响应
func (c *Context) Unauthorized(message string) {
	if message == "" {
		message = "未授权访问"
	}
	c.Error(ErrCodeUnauthorized, message)
}

// Forbidden 403错误响应
func (c *Context) Forbidden(message string) {
	if message == "" {
		message = "禁止访问"
	}
	c.Error(ErrCodeForbidden, message)
}

// NotFound 404错误响应
func (c *Context) NotFound(message string) {
	if message == "" {
		message = "资源不存在"
	}
	c.Error(ErrCodeNotFound, message)
}

// InternalServerError 500错误响应
func (c *Context) InternalServerError(message string) {
	if message == "" {
		message = "服务器内部错误"
	}
	c.Error(ErrCodeInternalServer, message)
}

// FormFile 获取上传的文件
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if c.Request.MultipartForm == nil {
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}
	}
	f, fh, err := c.Request.FormFile(name)
	if err != nil {
		return nil, err
	}
	f.Close()
	return fh, nil
}

// GetRawData 获取请求的原始数据
func (c *Context) GetRawData() ([]byte, error) {
	return io.ReadAll(c.Request.Body)
}

// ClientIP 获取客户端IP地址
func (c *Context) ClientIP() string {
	clientIP := c.Request.Header.Get("X-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP == "" {
		clientIP = strings.TrimSpace(c.Request.Header.Get("X-Real-Ip"))
	}
	if clientIP != "" {
		return clientIP
	}
	if addr := c.Request.Header.Get("X-Appengine-Remote-Addr"); addr != "" {
		return addr
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

// Header 设置响应头
func (c *Context) Header(key, value string) {
	c.Writer.Header().Set(key, value)
}
