package router

import (
	"strings"

	"github.com/guyigood/gyweb/core/context"
)

// HandlerFunc 使用 context 包中的 HandlerFunc 类型
type HandlerFunc = context.HandlerFunc

// node 路由树节点
type node struct {
	Pattern  string        // 导出字段
	part     string        // 路由中的一部分
	children []*node       // 子节点
	Handlers []HandlerFunc // 导出字段
	isWild   bool          // 是否模糊匹配（包含:或*）
}

// Router 路由接口
type Router interface {
	AddRoute(method string, pattern string, handlers ...HandlerFunc)
	GetRoute(method string, path string) (*node, map[string]string)
	GetHandlers(key string) []HandlerFunc
}

// router 路由实现
type router struct {
	roots    map[string]*node         // 每种请求方法的根节点
	handlers map[string][]HandlerFunc // 路由处理函数
}

// New 创建路由实例
func New() Router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string][]HandlerFunc),
	}
}

// parsePattern 解析路由模式
func parsePattern(pattern string) []string {
	parts := strings.Split(pattern, "/")
	result := make([]string, 0)
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
			if part[0] == '*' {
				break
			}
		}
	}
	return result
}

// AddRoute 添加路由
func (r *router) AddRoute(method string, pattern string, handlers ...HandlerFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern

	// 获取或创建方法对应的根节点
	root, exists := r.roots[method]
	if !exists {
		root = &node{}
		r.roots[method] = root
	}

	// 插入节点
	root.insert(pattern, parts, 0)
	r.handlers[key] = handlers
}

// GetRoute 获取路由
func (r *router) GetRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)
	if n != nil {
		parts := parsePattern(n.Pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}

// insert 插入节点
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.Pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

// search 搜索节点
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.Pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}

// matchChild 匹配单个子节点
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// matchChildren 匹配所有子节点
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// GetHandlers 获取处理函数
func (r *router) GetHandlers(key string) []HandlerFunc {
	return r.handlers[key]
}
