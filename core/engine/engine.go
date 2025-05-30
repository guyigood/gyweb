package engine

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"html/template"
	"log"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"

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

// TLSConfig TLS证书配置
type TLSConfig struct {
	CertFile string // 证书文件路径
	KeyFile  string // 私钥文件路径
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

// RunTLS 启动 HTTPS 服务器
func (engine *Engine) RunTLS(addr string, tlsConfig *TLSConfig) (err error) {
	log.Printf("Server is running on https://%s", addr)
	return http.ListenAndServeTLS(addr, tlsConfig.CertFile, tlsConfig.KeyFile, engine)
}

// RunAutoTLS 启动自动证书的 HTTPS 服务器（用于开发环境）
func (engine *Engine) RunAutoTLS(addr string) (err error) {
	log.Printf("Server is running on https://%s with self-signed certificate", addr)

	// 创建自签名证书（仅用于开发环境）
	certPEM, keyPEM, err := generateSelfSignedCert()
	if err != nil {
		return err
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:    addr,
		Handler: engine,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	return server.ListenAndServeTLS("", "")
}

// generateSelfSignedCert 生成自签名证书（仅用于开发环境）
func generateSelfSignedCert() ([]byte, []byte, error) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"GyWeb Development"},
			Country:       []string{"CN"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour), // 1年有效期
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost"},
	}

	// 创建证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	// 编码证书
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// 编码私钥
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return certPEM, keyPEM, nil
}
