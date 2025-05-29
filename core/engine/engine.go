package engine

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
	"github.com/guyigood/gyweb/core/router"
)

// Engine 是框架的核心结构
type Engine struct {
	*RouterGroup
	router        router.Router
	groups        []*RouterGroup
	htmlTemplates *template.Template
	funcMap       template.FuncMap
}

// RouterGroup 路由组
type RouterGroup struct {
	prefix      string
	middlewares []middleware.HandlerFunc
	parent      *RouterGroup
	engine      *Engine
}

// New 创建引擎实例
func New() *Engine {
	engine := &Engine{
		router: router.New(),
	}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Group 创建路由组
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// Any 注册所有请求方法的路由
func (group *RouterGroup) Any(pattern string, handler middleware.HandlerFunc) {
	group.addRoute("GET", pattern, handler)
	group.addRoute("POST", pattern, handler)
	group.addRoute("PUT", pattern, handler)
	group.addRoute("DELETE", pattern, handler)
}

// Use 添加中间件
func (group *RouterGroup) Use(middlewares ...middleware.HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// addRoute 添加路由
func (group *RouterGroup) addRoute(method string, comp string, handler middleware.HandlerFunc) {
	pattern := group.prefix + comp
	group.engine.router.AddRoute(method, pattern, handler)
}

// GET 注册 GET 请求
func (group *RouterGroup) GET(pattern string, handler middleware.HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST 注册 POST 请求
func (group *RouterGroup) POST(pattern string, handler middleware.HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// PUT 注册 PUT 请求
func (group *RouterGroup) PUT(pattern string, handler middleware.HandlerFunc) {
	group.addRoute("PUT", pattern, handler)
}

// DELETE 注册 DELETE 请求
func (group *RouterGroup) DELETE(pattern string, handler middleware.HandlerFunc) {
	group.addRoute("DELETE", pattern, handler)
}

// SetFuncMap 设置模板函数
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

// LoadHTMLGlob 加载 HTML 模板
func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

// ServeHTTP 实现 http.Handler 接口
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []middleware.HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	c := gyarn.NewContext(w, req)
	c.Handlers = middlewares
	engine.router.GetRoute(req.Method, req.URL.Path)

	if node, params := engine.router.GetRoute(req.Method, req.URL.Path); node != nil {
		c.Params = params
		key := req.Method + "-" + node.Pattern
		c.Handlers = append(c.Handlers, engine.router.GetHandlers(key)...)
	} else {
		c.Handlers = append(c.Handlers, func(c *gyarn.Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}

	c.Next()
}

// Run 启动 HTTP 服务器
func (engine *Engine) Run(addr string) (err error) {
	log.Printf("Server is running on %s", addr)
	return http.ListenAndServe(addr, engine)
}
