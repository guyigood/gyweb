package context

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
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
	Params     map[string]string
	StatusCode int
	Handlers   []HandlerFunc
	index      int
	aborted    bool // 添加 aborted 标志
}

// NewContext 创建新的上下文
func NewContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: req,
		Path:    req.URL.Path,
		Method:  req.Method,
		index:   -1,
	}
}

// Next 执行下一个中间件
func (c *Context) Next() {
	c.index++
	s := len(c.Handlers)
	for ; c.index < s && !c.aborted; c.index++ {
		c.Handlers[c.index](c)
	}
}

// JSON 发送 JSON 响应
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// XML 发送 XML 响应
func (c *Context) XML(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/xml")
	c.Status(code)
	encoder := xml.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// String 发送字符串响应
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(format))
}

// Data 发送数据响应
func (c *Context) Data(code int, data []byte) {
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

// PostForm 获取表单参数
func (c *Context) PostForm(key string) string {
	return c.Request.FormValue(key)
}

// Fail 返回错误响应
func (c *Context) Fail(code int, err string) {
	c.index = len(c.Handlers)
	c.JSON(code, H{"message": err})
}

// Abort 中止后续中间件的执行
func (c *Context) Abort() {
	c.aborted = true
}

// IsAborted 检查是否已中止
func (c *Context) IsAborted() bool {
	return c.aborted
}
