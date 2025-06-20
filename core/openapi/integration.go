package openapi

import (
	"fmt"
	"net/http"

	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
)

// OpenAPIConfig OpenAPI配置
type OpenAPIConfig struct {
	Title       string
	Description string
	Version     string
	DocsPath    string // 文档路径，默认 /swagger
	JSONPath    string // JSON API路径，默认 /swagger/openapi.json
	Enabled     bool   // 是否启用，默认 true
}

// DefaultConfig 默认配置
func DefaultConfig() OpenAPIConfig {
	return OpenAPIConfig{
		Title:       "GyWeb API",
		Description: "基于GyWeb框架的API文档",
		Version:     "1.0.0",
		DocsPath:    "/swagger",
		JSONPath:    "/swagger/openapi.json",
		Enabled:     true,
	}
}

// EngineExtension 引擎扩展
type EngineExtension struct {
	engine  *engine.Engine
	openapi *OpenAPI
	config  OpenAPIConfig
}

// NewEngineExtension 创建引擎扩展
func NewEngineExtension(e *engine.Engine, config ...OpenAPIConfig) *EngineExtension {
	cfg := DefaultConfig()
	if len(config) > 0 {
		// 合并配置而不是完全替换
		userConfig := config[0]
		if userConfig.Title != "" {
			cfg.Title = userConfig.Title
		}
		if userConfig.Description != "" {
			cfg.Description = userConfig.Description
		}
		if userConfig.Version != "" {
			cfg.Version = userConfig.Version
		}
		if userConfig.DocsPath != "" {
			cfg.DocsPath = userConfig.DocsPath
		}
		if userConfig.JSONPath != "" {
			cfg.JSONPath = userConfig.JSONPath
		}
		// 保持默认的Enabled=true，除非用户显式传入了完整的禁用配置
		// 这里简化逻辑：只要用户设置了基本信息，就认为想要启用OpenAPI
	}

	fmt.Printf("[DEBUG OpenAPI] 配置: Enabled=%v, DocsPath=%s\n", cfg.Enabled, cfg.DocsPath)

	openapi := New().
		SetInfo(Info{
			Title:       cfg.Title,
			Description: cfg.Description,
			Version:     cfg.Version,
		}).
		AddCommonSecuritySchemes().
		AddCommonResponses().
		AutoRegisterModels()

	ext := &EngineExtension{
		engine:  e,
		openapi: openapi,
		config:  cfg,
	}

	// 注册文档路由
	if cfg.Enabled {
		ext.registerRoutes()
	} else {
		fmt.Printf("[DEBUG OpenAPI] OpenAPI已禁用，跳过路由注册\n")
	}

	return ext
}

// registerRoutes 注册文档路由
func (ext *EngineExtension) registerRoutes() {
	fmt.Printf("[DEBUG OpenAPI] 开始注册文档路由\n")
	fmt.Printf("[DEBUG OpenAPI] DocsPath: %s\n", ext.config.DocsPath)
	fmt.Printf("[DEBUG OpenAPI] JSONPath: %s\n", ext.config.JSONPath)

	// 注册OpenAPI JSON端点
	ext.engine.GET(ext.config.JSONPath, func(c *gyarn.Context) {
		c.Header("Content-Type", "application/json")
		ext.openapi.buildPaths() // 确保路径已构建
		jsonData, err := ext.openapi.ToJSON()
		if err != nil {
			c.String(http.StatusInternalServerError, "生成OpenAPI文档失败: %v", err)
			return
		}
		c.Data(http.StatusOK, "application/json", jsonData)
	})

	// 注册Swagger UI主页面
	ext.engine.GET(ext.config.DocsPath, func(c *gyarn.Context) {
		ext.openapi.buildPaths() // 确保路径已构建
		html := ext.openapi.generateSwaggerHTML()
		c.HTML(http.StatusOK, html)
	})

	fmt.Printf("[DEBUG OpenAPI] 文档路由注册完成\n")
}

// GetOpenAPI 获取OpenAPI实例
func (ext *EngineExtension) GetOpenAPI() *OpenAPI {
	return ext.openapi
}

// AddRoute 添加路由文档
func (ext *EngineExtension) AddRoute(method, path string, doc APIDoc) *EngineExtension {
	ext.openapi.AddRoute(method, path, doc)
	return ext
}

// AddServer 添加服务器信息
func (ext *EngineExtension) AddServer(server Server) *EngineExtension {
	ext.openapi.AddServer(server)
	return ext
}

// AddTag 添加标签
func (ext *EngineExtension) AddTag(tag Tag) *EngineExtension {
	ext.openapi.AddTag(tag)
	return ext
}

// AddSecurityScheme 添加安全方案
func (ext *EngineExtension) AddSecurityScheme(name string, scheme SecurityScheme) *EngineExtension {
	ext.openapi.AddSecurityScheme(name, scheme)
	return ext
}

// AddSchema 添加模式定义
func (ext *EngineExtension) AddSchema(name string, schema *Schema) *EngineExtension {
	ext.openapi.AddSchema(name, schema)
	return ext
}

// GenerateFromAnnotations 从注解生成文档
func (ext *EngineExtension) GenerateFromAnnotations(sourceDir string) error {
	fmt.Printf("[DEBUG OpenAPI] 开始解析注解，源目录: %s\n", sourceDir)

	// 使用增强的注解解析器
	parser := NewAnnotationParser(ext.openapi)

	// 递归解析目录，自动发现所有模型
	err := parser.RecursiveParseDirectory(sourceDir)
	if err != nil {
		return fmt.Errorf("解析注解失败: %v", err)
	}

	fmt.Printf("[DEBUG OpenAPI] 注解解析完成\n")
	return nil
}

// RegisterModel 注册单个模型到OpenAPI schema
func (ext *EngineExtension) RegisterModel(name string, model interface{}) *EngineExtension {
	schema := ext.openapi.GenerateFromStruct(model)
	ext.openapi.AddSchema(name, schema)
	fmt.Printf("[DEBUG OpenAPI] 手动注册模型: %s\n", name)
	return ext
}

// RegisterModels 批量注册模型
func (ext *EngineExtension) RegisterModels(models map[string]interface{}) *EngineExtension {
	for name, model := range models {
		ext.RegisterModel(name, model)
	}
	return ext
}

// AutoDiscoverModels 自动发现并注册指定包中的所有模型
func (ext *EngineExtension) AutoDiscoverModels(packagePaths ...string) *EngineExtension {
	for _, packagePath := range packagePaths {
		err := ext.GenerateFromAnnotations(packagePath)
		if err != nil {
			fmt.Printf("[WARNING OpenAPI] 自动发现模型失败 %s: %v\n", packagePath, err)
		}
	}
	return ext
}

// 为引擎添加OpenAPI支持的便捷方法

// EnableOpenAPI 为引擎启用OpenAPI支持
func EnableOpenAPI(e *engine.Engine, config ...OpenAPIConfig) *EngineExtension {
	return NewEngineExtension(e, config...)
}

// RouteDoc 路由文档装饰器
func RouteDoc(doc APIDoc) middleware.HandlerFunc {
	return func(c *gyarn.Context) {
		// 这个中间件主要用于标记路由文档信息
		// 实际的文档生成在其他地方处理
		c.Next()
	}
}

// WithDoc 为路由添加文档的便捷方法
func WithDoc(handler middleware.HandlerFunc, doc APIDoc) middleware.HandlerFunc {
	return func(c *gyarn.Context) {
		// 可以在这里添加文档相关的元数据到context
		handler(c)
	}
}

// 常用的文档构建器

// DocBuilder 文档构建器
type DocBuilder struct {
	doc APIDoc
}

// NewDocBuilder 创建文档构建器
func NewDocBuilder() *DocBuilder {
	return &DocBuilder{
		doc: APIDoc{
			Responses: make(map[string]Response),
		},
	}
}

// Summary 设置摘要
func (b *DocBuilder) Summary(summary string) *DocBuilder {
	b.doc.Summary = summary
	return b
}

// Description 设置描述
func (b *DocBuilder) Description(description string) *DocBuilder {
	b.doc.Description = description
	return b
}

// Tags 设置标签
func (b *DocBuilder) Tags(tags ...string) *DocBuilder {
	b.doc.Tags = tags
	return b
}

// Param 添加参数
func (b *DocBuilder) Param(name, in, paramType string, required bool, description string) *DocBuilder {
	param := Parameter{
		Name:        name,
		In:          in,
		Description: description,
		Required:    required,
		Schema:      &Schema{Type: paramType},
	}
	b.doc.Parameters = append(b.doc.Parameters, param)
	return b
}

// QueryParam 添加查询参数
func (b *DocBuilder) QueryParam(name, paramType string, required bool, description string) *DocBuilder {
	return b.Param(name, "query", paramType, required, description)
}

// PathParam 添加路径参数
func (b *DocBuilder) PathParam(name, paramType string, description string) *DocBuilder {
	return b.Param(name, "path", paramType, true, description)
}

// HeaderParam 添加头部参数
func (b *DocBuilder) HeaderParam(name, paramType string, required bool, description string) *DocBuilder {
	return b.Param(name, "header", paramType, required, description)
}

// RequestBody 设置请求体
func (b *DocBuilder) RequestBody(description string, contentType string, schema *Schema, required bool) *DocBuilder {
	if b.doc.RequestBody == nil {
		b.doc.RequestBody = &RequestBody{
			Content: make(map[string]MediaType),
		}
	}
	b.doc.RequestBody.Description = description
	b.doc.RequestBody.Required = required
	b.doc.RequestBody.Content[contentType] = MediaType{
		Schema: schema,
	}
	return b
}

// JSONRequestBody 设置JSON请求体
func (b *DocBuilder) JSONRequestBody(description string, schema *Schema, required bool) *DocBuilder {
	return b.RequestBody(description, "application/json", schema, required)
}

// Response 添加响应
func (b *DocBuilder) Response(code, description string, contentType string, schema *Schema) *DocBuilder {
	response := Response{
		Description: description,
		Content:     make(map[string]MediaType),
	}
	if schema != nil {
		response.Content[contentType] = MediaType{
			Schema: schema,
		}
	}
	b.doc.Responses[code] = response
	return b
}

// JSONResponse 添加JSON响应
func (b *DocBuilder) JSONResponse(code, description string, schema *Schema) *DocBuilder {
	return b.Response(code, description, "application/json", schema)
}

// SuccessResponse 添加成功响应
func (b *DocBuilder) SuccessResponse(description string, schema *Schema) *DocBuilder {
	return b.JSONResponse("200", description, schema)
}

// ErrorResponse 添加错误响应
func (b *DocBuilder) ErrorResponse(code, description string) *DocBuilder {
	return b.JSONResponse(code, description, &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"code":    {Type: "integer"},
			"message": {Type: "string"},
		},
	})
}

// Security 添加安全要求
func (b *DocBuilder) Security(name string, scopes ...string) *DocBuilder {
	b.doc.Security = append(b.doc.Security, SecurityRequirement{
		name: scopes,
	})
	return b
}

// Deprecated 标记为已弃用
func (b *DocBuilder) Deprecated() *DocBuilder {
	b.doc.Deprecated = true
	return b
}

// Build 构建文档
func (b *DocBuilder) Build() APIDoc {
	return b.doc
}

// 常用的Schema构建器

// SchemaBuilder Schema构建器
type SchemaBuilder struct {
	schema *Schema
}

// NewSchemaBuilder 创建Schema构建器
func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		schema: &Schema{
			Properties: make(map[string]*Schema),
		},
	}
}

// Type 设置类型
func (b *SchemaBuilder) Type(schemaType string) *SchemaBuilder {
	b.schema.Type = schemaType
	return b
}

// Format 设置格式
func (b *SchemaBuilder) Format(format string) *SchemaBuilder {
	b.schema.Format = format
	return b
}

// Description 设置描述
func (b *SchemaBuilder) Description(description string) *SchemaBuilder {
	b.schema.Description = description
	return b
}

// Example 设置示例
func (b *SchemaBuilder) Example(example interface{}) *SchemaBuilder {
	b.schema.Example = example
	return b
}

// Property 添加属性
func (b *SchemaBuilder) Property(name string, schema *Schema) *SchemaBuilder {
	if b.schema.Properties == nil {
		b.schema.Properties = make(map[string]*Schema)
	}
	b.schema.Properties[name] = schema
	return b
}

// StringProperty 添加字符串属性
func (b *SchemaBuilder) StringProperty(name, description string, example ...string) *SchemaBuilder {
	schema := &Schema{
		Type:        "string",
		Description: description,
	}
	if len(example) > 0 {
		schema.Example = example[0]
	}
	return b.Property(name, schema)
}

// IntegerProperty 添加整数属性
func (b *SchemaBuilder) IntegerProperty(name, description string, example ...int) *SchemaBuilder {
	schema := &Schema{
		Type:        "integer",
		Description: description,
	}
	if len(example) > 0 {
		schema.Example = example[0]
	}
	return b.Property(name, schema)
}

// NumberProperty 添加数字属性
func (b *SchemaBuilder) NumberProperty(name, description string, example ...float64) *SchemaBuilder {
	schema := &Schema{
		Type:        "number",
		Description: description,
	}
	if len(example) > 0 {
		schema.Example = example[0]
	}
	return b.Property(name, schema)
}

// BooleanProperty 添加布尔属性
func (b *SchemaBuilder) BooleanProperty(name, description string, example ...bool) *SchemaBuilder {
	schema := &Schema{
		Type:        "boolean",
		Description: description,
	}
	if len(example) > 0 {
		schema.Example = example[0]
	}
	return b.Property(name, schema)
}

// ArrayProperty 添加数组属性
func (b *SchemaBuilder) ArrayProperty(name, description string, items *Schema) *SchemaBuilder {
	schema := &Schema{
		Type:        "array",
		Description: description,
		Items:       items,
	}
	return b.Property(name, schema)
}

// ObjectProperty 添加对象属性
func (b *SchemaBuilder) ObjectProperty(name, description string, properties map[string]*Schema) *SchemaBuilder {
	schema := &Schema{
		Type:        "object",
		Description: description,
		Properties:  properties,
	}
	return b.Property(name, schema)
}

// Required 设置必需字段
func (b *SchemaBuilder) Required(fields ...string) *SchemaBuilder {
	b.schema.Required = fields
	return b
}

// Items 设置数组项类型
func (b *SchemaBuilder) Items(items *Schema) *SchemaBuilder {
	b.schema.Items = items
	return b
}

// Ref 设置引用
func (b *SchemaBuilder) Ref(ref string) *SchemaBuilder {
	b.schema.Ref = ref
	return b
}

// Build 构建Schema
func (b *SchemaBuilder) Build() *Schema {
	return b.schema
}

// 常用的预定义Schema

// StringSchema 字符串Schema
func StringSchema(description string, example ...string) *Schema {
	schema := &Schema{
		Type:        "string",
		Description: description,
	}
	if len(example) > 0 {
		schema.Example = example[0]
	}
	return schema
}

// IntegerSchema 整数Schema
func IntegerSchema(description string, example ...int) *Schema {
	schema := &Schema{
		Type:        "integer",
		Description: description,
	}
	if len(example) > 0 {
		schema.Example = example[0]
	}
	return schema
}

// NumberSchema 数字Schema
func NumberSchema(description string, example ...float64) *Schema {
	schema := &Schema{
		Type:        "number",
		Description: description,
	}
	if len(example) > 0 {
		schema.Example = example[0]
	}
	return schema
}

// BooleanSchema 布尔Schema
func BooleanSchema(description string, example ...bool) *Schema {
	schema := &Schema{
		Type:        "boolean",
		Description: description,
	}
	if len(example) > 0 {
		schema.Example = example[0]
	}
	return schema
}

// ArraySchema 数组Schema
func ArraySchema(description string, items *Schema) *Schema {
	return &Schema{
		Type:        "array",
		Description: description,
		Items:       items,
	}
}

// ObjectSchema 对象Schema
func ObjectSchema(description string, properties map[string]*Schema, required ...string) *Schema {
	return &Schema{
		Type:        "object",
		Description: description,
		Properties:  properties,
		Required:    required,
	}
}

// RefSchema 引用Schema
func RefSchema(ref string) *Schema {
	return &Schema{
		Ref: ref,
	}
}
