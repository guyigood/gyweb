package gyarn

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// HandlerFunc 定义处理函数类型
type HandlerFunc func(*Ctx)

// H 是一个便捷的 map 类型，用于 JSON 响应
type H map[string]interface{}

// Context 封装了请求和响应
type Ctx struct {
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
func NewContext(w http.ResponseWriter, req *http.Request) *Ctx {
	return &Ctx{
		Writer:  w,
		Request: req,
		Path:    req.URL.Path,
		Method:  req.Method,
		index:   -1,
	}
}

// Next 执行下一个中间件
func (c *Ctx) Next() {
	c.index++
	s := len(c.Handlers)
	for ; c.index < s && !c.aborted; c.index++ {
		c.Handlers[c.index](c)
	}
}

// JSON 发送 JSON 响应
func (c *Ctx) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// XML 发送 XML 响应
func (c *Ctx) XML(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/xml")
	c.Status(code)
	encoder := xml.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// String 发送字符串响应
func (c *Ctx) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(format))
}

// Data 发送数据响应
func (c *Ctx) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// HTML 发送 HTML 响应
func (c *Ctx) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

// Status 设置状态码
func (c *Ctx) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// SetHeader 设置响应头
func (c *Ctx) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// Param 获取路由参数
func (c *Ctx) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

// Query 获取查询参数
func (c *Ctx) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// PostForm 获取表单参数
func (c *Ctx) PostForm(key string) string {
	return c.Request.FormValue(key)
}

// Fail 返回错误响应
func (c *Ctx) Fail(code int, err string) {
	c.index = len(c.Handlers)
	c.JSON(code, H{"message": err})
}

// Abort 中止后续中间件的执行
func (c *Ctx) Abort() {
	c.aborted = true
}

// IsAborted 检查是否已中止
func (c *Ctx) IsAborted() bool {
	return c.aborted
}
