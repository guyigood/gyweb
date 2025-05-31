package openapi

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// AnnotationParser 注解解析器
type AnnotationParser struct {
	openapi *OpenAPI
	fileSet *token.FileSet
}

// NewAnnotationParser 创建注解解析器
func NewAnnotationParser(openapi *OpenAPI) *AnnotationParser {
	return &AnnotationParser{
		openapi: openapi,
		fileSet: token.NewFileSet(),
	}
}

// ParseDirectory 解析目录中的Go文件
func (p *AnnotationParser) ParseDirectory(dir string) error {
	pattern := filepath.Join(dir, "*.go")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := p.ParseFile(file); err != nil {
			return fmt.Errorf("解析文件 %s 失败: %v", file, err)
		}
	}

	return nil
}

// ParseFile 解析单个Go文件
func (p *AnnotationParser) ParseFile(filename string) error {
	src, err := parser.ParseFile(p.fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// 遍历所有函数
	for _, decl := range src.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			p.parseFunction(fn)
		}
	}

	return nil
}

// parseFunction 解析函数注释
func (p *AnnotationParser) parseFunction(fn *ast.FuncDecl) {
	if fn.Doc == nil {
		return
	}

	comments := fn.Doc.Text()
	if !strings.Contains(comments, "@Router") {
		return
	}

	doc := p.parseComments(comments)
	if doc != nil {
		// 从注释中提取路由信息
		if route := p.extractRoute(comments); route != nil {
			p.openapi.AddRoute(route.Method, route.Path, *doc)
		}
	}
}

// RouteInfo 路由信息
type RouteAnnotation struct {
	Method string
	Path   string
}

// extractRoute 从注释中提取路由信息
func (p *AnnotationParser) extractRoute(comments string) *RouteAnnotation {
	// 匹配 @Router /path [method]
	re := regexp.MustCompile(`@Router\s+([^\s]+)\s+\[([^\]]+)\]`)
	matches := re.FindStringSubmatch(comments)
	if len(matches) != 3 {
		return nil
	}

	return &RouteAnnotation{
		Path:   matches[1],
		Method: strings.ToUpper(matches[2]),
	}
}

// parseComments 解析注释生成API文档
func (p *AnnotationParser) parseComments(comments string) *APIDoc {
	doc := &APIDoc{
		Responses: make(map[string]Response),
	}

	lines := strings.Split(comments, "\n")
	var currentSection string
	var buffer []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 处理各种注解
		switch {
		case strings.HasPrefix(line, "@Summary"):
			doc.Summary = strings.TrimSpace(strings.TrimPrefix(line, "@Summary"))
		case strings.HasPrefix(line, "@Description"):
			doc.Description = strings.TrimSpace(strings.TrimPrefix(line, "@Description"))
		case strings.HasPrefix(line, "@Tags"):
			tags := strings.TrimSpace(strings.TrimPrefix(line, "@Tags"))
			doc.Tags = strings.Split(tags, ",")
			for i := range doc.Tags {
				doc.Tags[i] = strings.TrimSpace(doc.Tags[i])
			}
		case strings.HasPrefix(line, "@Deprecated"):
			doc.Deprecated = true
		case strings.HasPrefix(line, "@Param"):
			if param := p.parseParam(line); param != nil {
				doc.Parameters = append(doc.Parameters, *param)
			}
		case strings.HasPrefix(line, "@Success"):
			if response := p.parseResponse(line); response != nil {
				code := p.extractResponseCode(line)
				doc.Responses[code] = *response
			}
		case strings.HasPrefix(line, "@Failure"):
			if response := p.parseResponse(line); response != nil {
				code := p.extractResponseCode(line)
				doc.Responses[code] = *response
			}
		case strings.HasPrefix(line, "@Accept"):
			// 处理请求体类型
			contentType := strings.TrimSpace(strings.TrimPrefix(line, "@Accept"))
			if doc.RequestBody == nil {
				doc.RequestBody = &RequestBody{
					Content: make(map[string]MediaType),
				}
			}
			doc.RequestBody.Content[contentType] = MediaType{}
		case strings.HasPrefix(line, "@Produce"):
			// 处理响应类型（已在Success/Failure中处理）
			continue
		case strings.HasPrefix(line, "@Security"):
			security := strings.TrimSpace(strings.TrimPrefix(line, "@Security"))
			doc.Security = append(doc.Security, SecurityRequirement{
				security: []string{},
			})
		case strings.HasPrefix(line, "@Router"):
			// 路由信息已在其他地方处理
			continue
		default:
			// 处理多行描述
			if currentSection != "" {
				buffer = append(buffer, line)
			}
		}
	}

	return doc
}

// parseParam 解析参数注解
// @Param name query string true "参数描述" default(value)
func (p *AnnotationParser) parseParam(line string) *Parameter {
	// 移除 @Param 前缀
	content := strings.TrimSpace(strings.TrimPrefix(line, "@Param"))
	parts := strings.Fields(content)

	if len(parts) < 4 {
		return nil
	}

	param := &Parameter{
		Name: parts[0],
		In:   parts[1],
	}

	// 解析类型
	paramType := parts[2]
	param.Schema = &Schema{Type: p.convertType(paramType)}

	// 解析是否必需
	if len(parts) > 3 {
		param.Required = parts[3] == "true"
	}

	// 解析描述（在引号中）
	if descMatch := regexp.MustCompile(`"([^"]+)"`).FindStringSubmatch(content); len(descMatch) > 1 {
		param.Description = descMatch[1]
	}

	// 解析默认值
	if defaultMatch := regexp.MustCompile(`default\(([^)]+)\)`).FindStringSubmatch(content); len(defaultMatch) > 1 {
		param.Schema.Default = defaultMatch[1]
	}

	// 解析示例
	if exampleMatch := regexp.MustCompile(`example\(([^)]+)\)`).FindStringSubmatch(content); len(exampleMatch) > 1 {
		param.Example = exampleMatch[1]
	}

	return param
}

// parseResponse 解析响应注解
// @Success 200 {object} User "成功返回用户信息"
func (p *AnnotationParser) parseResponse(line string) *Response {
	// 提取描述（在引号中）
	var description string
	if descMatch := regexp.MustCompile(`"([^"]+)"`).FindStringSubmatch(line); len(descMatch) > 1 {
		description = descMatch[1]
	} else {
		description = "成功"
	}

	response := &Response{
		Description: description,
		Content:     make(map[string]MediaType),
	}

	// 解析响应体类型
	if typeMatch := regexp.MustCompile(`\{([^}]+)\}\s+(\w+)`).FindStringSubmatch(line); len(typeMatch) > 2 {
		responseType := typeMatch[1]
		modelName := typeMatch[2]

		var schema *Schema
		if responseType == "object" {
			schema = &Schema{
				Ref: fmt.Sprintf("#/components/schemas/%s", modelName),
			}
		} else {
			schema = &Schema{Type: p.convertType(responseType)}
		}

		response.Content["application/json"] = MediaType{
			Schema: schema,
		}
	}

	return response
}

// extractResponseCode 提取响应状态码
func (p *AnnotationParser) extractResponseCode(line string) string {
	parts := strings.Fields(line)
	if len(parts) > 1 {
		if code, err := strconv.Atoi(parts[1]); err == nil && code >= 100 && code < 600 {
			return parts[1]
		}
	}
	return "200"
}

// convertType 转换Go类型到OpenAPI类型
func (p *AnnotationParser) convertType(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64":
		return "integer"
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "[]string", "[]int", "[]float64":
		return "array"
	default:
		return "object"
	}
}

// GenerateFromAnnotations 从注解生成文档
func (o *OpenAPI) GenerateFromAnnotations(sourceDir string) error {
	parser := NewAnnotationParser(o)
	return parser.ParseDirectory(sourceDir)
}

// 预定义的常用响应模式
var CommonResponses = map[string]Response{
	"Success": {
		Description: "成功",
		Content: map[string]MediaType{
			"application/json": {
				Schema: &Schema{
					Type: "object",
					Properties: map[string]*Schema{
						"code":    {Type: "integer", Example: 200},
						"message": {Type: "string", Example: "success"},
						"data":    {Type: "object"},
					},
				},
			},
		},
	},
	"BadRequest": {
		Description: "请求参数错误",
		Content: map[string]MediaType{
			"application/json": {
				Schema: &Schema{
					Type: "object",
					Properties: map[string]*Schema{
						"code":    {Type: "integer", Example: 400},
						"message": {Type: "string", Example: "请求参数错误"},
					},
				},
			},
		},
	},
	"Unauthorized": {
		Description: "未授权",
		Content: map[string]MediaType{
			"application/json": {
				Schema: &Schema{
					Type: "object",
					Properties: map[string]*Schema{
						"code":    {Type: "integer", Example: 401},
						"message": {Type: "string", Example: "未授权"},
					},
				},
			},
		},
	},
	"NotFound": {
		Description: "资源不存在",
		Content: map[string]MediaType{
			"application/json": {
				Schema: &Schema{
					Type: "object",
					Properties: map[string]*Schema{
						"code":    {Type: "integer", Example: 404},
						"message": {Type: "string", Example: "资源不存在"},
					},
				},
			},
		},
	},
	"InternalServerError": {
		Description: "服务器内部错误",
		Content: map[string]MediaType{
			"application/json": {
				Schema: &Schema{
					Type: "object",
					Properties: map[string]*Schema{
						"code":    {Type: "integer", Example: 500},
						"message": {Type: "string", Example: "服务器内部错误"},
					},
				},
			},
		},
	},
}

// AddCommonResponses 添加常用响应模式
func (o *OpenAPI) AddCommonResponses() *OpenAPI {
	if o.Components.Responses == nil {
		o.Components.Responses = make(map[string]Response)
	}

	for name, response := range CommonResponses {
		o.Components.Responses[name] = response
	}

	return o
}

// AddCommonSecuritySchemes 添加常用安全方案
func (o *OpenAPI) AddCommonSecuritySchemes() *OpenAPI {
	// JWT Bearer Token
	o.AddSecurityScheme("BearerAuth", SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT Bearer Token认证",
	})

	// API Key
	o.AddSecurityScheme("ApiKeyAuth", SecurityScheme{
		Type:        "apiKey",
		In:          "header",
		Name:        "X-API-Key",
		Description: "API Key认证",
	})

	// Basic Auth
	o.AddSecurityScheme("BasicAuth", SecurityScheme{
		Type:        "http",
		Scheme:      "basic",
		Description: "HTTP Basic认证",
	})

	return o
}

// ModelRegistry 模型注册表
type ModelRegistry struct {
	models map[string]*Schema
}

// NewModelRegistry 创建模型注册表
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		models: make(map[string]*Schema),
	}
}

// RegisterModel 注册模型
func (r *ModelRegistry) RegisterModel(name string, model interface{}) {
	openapi := New()
	schema := openapi.GenerateFromStruct(model)
	r.models[name] = schema
}

// GetModels 获取所有模型
func (r *ModelRegistry) GetModels() map[string]*Schema {
	return r.models
}

// RegisterModelsToOpenAPI 将模型注册到OpenAPI
func (r *ModelRegistry) RegisterModelsToOpenAPI(openapi *OpenAPI) {
	for name, schema := range r.models {
		openapi.AddSchema(name, schema)
	}
}

// 全局模型注册表
var GlobalModelRegistry = NewModelRegistry()

// RegisterModel 注册全局模型
func RegisterModel(name string, model interface{}) {
	GlobalModelRegistry.RegisterModel(name, model)
}

// AutoRegisterModels 自动注册模型到OpenAPI
func (o *OpenAPI) AutoRegisterModels() *OpenAPI {
	GlobalModelRegistry.RegisterModelsToOpenAPI(o)
	return o
}
