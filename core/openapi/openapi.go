package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
)

// OpenAPI OpenAPI文档生成器
type OpenAPI struct {
	Info       Info                  `json:"info"`
	OpenAPIVer string                `json:"openapi"`
	Servers    []Server              `json:"servers,omitempty"`
	Paths      map[string]PathItem   `json:"paths"`
	Components Components            `json:"components,omitempty"`
	Tags       []Tag                 `json:"tags,omitempty"`
	Security   []SecurityRequirement `json:"security,omitempty"`
	routes     []RouteInfo           // 内部路由信息
}

// Info API基本信息
type Info struct {
	Title          string  `json:"title"`
	Description    string  `json:"description,omitempty"`
	Version        string  `json:"version"`
	TermsOfService string  `json:"termsOfService,omitempty"`
	Contact        Contact `json:"contact,omitempty"`
	License        License `json:"license,omitempty"`
}

// Contact 联系信息
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License 许可证信息
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Server 服务器信息
type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable 服务器变量
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// PathItem 路径项
type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Options *Operation `json:"options,omitempty"`
	Head    *Operation `json:"head,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Trace   *Operation `json:"trace,omitempty"`
}

// Operation 操作信息
type Operation struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []SecurityRequirement `json:"security,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty"`
}

// Parameter 参数信息
type Parameter struct {
	Name            string      `json:"name"`
	In              string      `json:"in"` // query, header, path, cookie
	Description     string      `json:"description,omitempty"`
	Required        bool        `json:"required,omitempty"`
	Deprecated      bool        `json:"deprecated,omitempty"`
	AllowEmptyValue bool        `json:"allowEmptyValue,omitempty"`
	Schema          *Schema     `json:"schema,omitempty"`
	Example         interface{} `json:"example,omitempty"`
}

// RequestBody 请求体
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Content     map[string]MediaType `json:"content"`
	Required    bool                 `json:"required,omitempty"`
}

// Response 响应信息
type Response struct {
	Description string               `json:"description"`
	Headers     map[string]Header    `json:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// MediaType 媒体类型
type MediaType struct {
	Schema   *Schema            `json:"schema,omitempty"`
	Example  interface{}        `json:"example,omitempty"`
	Examples map[string]Example `json:"examples,omitempty"`
}

// Header 头部信息
type Header struct {
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Deprecated  bool        `json:"deprecated,omitempty"`
	Schema      *Schema     `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// Example 示例
type Example struct {
	Summary       string      `json:"summary,omitempty"`
	Description   string      `json:"description,omitempty"`
	Value         interface{} `json:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty"`
}

// Schema 模式定义
type Schema struct {
	Type                 string             `json:"type,omitempty"`
	Format               string             `json:"format,omitempty"`
	Title                string             `json:"title,omitempty"`
	Description          string             `json:"description,omitempty"`
	Default              interface{}        `json:"default,omitempty"`
	Example              interface{}        `json:"example,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	AdditionalProperties interface{}        `json:"additionalProperties,omitempty"`
	Enum                 []interface{}      `json:"enum,omitempty"`
	Minimum              *float64           `json:"minimum,omitempty"`
	Maximum              *float64           `json:"maximum,omitempty"`
	MinLength            *int               `json:"minLength,omitempty"`
	MaxLength            *int               `json:"maxLength,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	Ref                  string             `json:"$ref,omitempty"`
}

// Components 组件定义
type Components struct {
	Schemas         map[string]*Schema        `json:"schemas,omitempty"`
	Responses       map[string]Response       `json:"responses,omitempty"`
	Parameters      map[string]Parameter      `json:"parameters,omitempty"`
	Examples        map[string]Example        `json:"examples,omitempty"`
	RequestBodies   map[string]RequestBody    `json:"requestBodies,omitempty"`
	Headers         map[string]Header         `json:"headers,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme 安全方案
type SecurityScheme struct {
	Type             string     `json:"type"`
	Description      string     `json:"description,omitempty"`
	Name             string     `json:"name,omitempty"`
	In               string     `json:"in,omitempty"`
	Scheme           string     `json:"scheme,omitempty"`
	BearerFormat     string     `json:"bearerFormat,omitempty"`
	Flows            OAuthFlows `json:"flows,omitempty"`
	OpenIDConnectURL string     `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows OAuth流程
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow OAuth流程详情
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// SecurityRequirement 安全要求
type SecurityRequirement map[string][]string

// Tag 标签
type Tag struct {
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
}

// ExternalDocs 外部文档
type ExternalDocs struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

// RouteInfo 路由信息（内部使用）
type RouteInfo struct {
	Method      string
	Path        string
	Handler     middleware.HandlerFunc
	Summary     string
	Description string
	Tags        []string
	Parameters  []Parameter
	RequestBody *RequestBody
	Responses   map[string]Response
	Security    []SecurityRequirement
	Deprecated  bool
}

// APIDoc API文档注解
type APIDoc struct {
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses,omitempty"`
	Security    []SecurityRequirement `json:"security,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty"`
}

// New 创建OpenAPI文档生成器
func New() *OpenAPI {
	return &OpenAPI{
		OpenAPIVer: "3.0.3",
		Info: Info{
			Title:   "GyWeb API",
			Version: "1.0.0",
		},
		Paths: make(map[string]PathItem),
		Components: Components{
			Schemas:         make(map[string]*Schema),
			SecuritySchemes: make(map[string]SecurityScheme),
		},
		routes: make([]RouteInfo, 0),
	}
}

// SetInfo 设置API基本信息
func (o *OpenAPI) SetInfo(info Info) *OpenAPI {
	o.Info = info
	return o
}

// AddServer 添加服务器信息
func (o *OpenAPI) AddServer(server Server) *OpenAPI {
	o.Servers = append(o.Servers, server)
	return o
}

// AddTag 添加标签
func (o *OpenAPI) AddTag(tag Tag) *OpenAPI {
	o.Tags = append(o.Tags, tag)
	return o
}

// AddSecurityScheme 添加安全方案
func (o *OpenAPI) AddSecurityScheme(name string, scheme SecurityScheme) *OpenAPI {
	if o.Components.SecuritySchemes == nil {
		o.Components.SecuritySchemes = make(map[string]SecurityScheme)
	}
	o.Components.SecuritySchemes[name] = scheme
	return o
}

// AddSchema 添加模式定义
func (o *OpenAPI) AddSchema(name string, schema *Schema) *OpenAPI {
	if o.Components.Schemas == nil {
		o.Components.Schemas = make(map[string]*Schema)
	}
	o.Components.Schemas[name] = schema
	return o
}

// AddRoute 添加路由文档
func (o *OpenAPI) AddRoute(method, path string, doc APIDoc) *OpenAPI {
	route := RouteInfo{
		Method:      strings.ToUpper(method),
		Path:        path,
		Summary:     doc.Summary,
		Description: doc.Description,
		Tags:        doc.Tags,
		Parameters:  doc.Parameters,
		RequestBody: doc.RequestBody,
		Responses:   doc.Responses,
		Security:    doc.Security,
		Deprecated:  doc.Deprecated,
	}

	o.routes = append(o.routes, route)
	o.buildPaths()
	return o
}

// buildPaths 构建路径信息
func (o *OpenAPI) buildPaths() {
	o.Paths = make(map[string]PathItem)

	for _, route := range o.routes {
		pathItem, exists := o.Paths[route.Path]
		if !exists {
			pathItem = PathItem{}
		}

		operation := &Operation{
			Tags:        route.Tags,
			Summary:     route.Summary,
			Description: route.Description,
			OperationID: generateOperationID(route.Method, route.Path),
			Parameters:  route.Parameters,
			RequestBody: route.RequestBody,
			Responses:   route.Responses,
			Security:    route.Security,
			Deprecated:  route.Deprecated,
		}

		// 如果没有响应定义，添加默认响应
		if operation.Responses == nil {
			operation.Responses = map[string]Response{
				"200": {
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
			}
		}

		switch route.Method {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		case "OPTIONS":
			pathItem.Options = operation
		case "HEAD":
			pathItem.Head = operation
		case "PATCH":
			pathItem.Patch = operation
		case "TRACE":
			pathItem.Trace = operation
		}

		o.Paths[route.Path] = pathItem
	}
}

// generateOperationID 生成操作ID
func generateOperationID(method, path string) string {
	// 移除路径参数的冒号和星号
	cleanPath := strings.ReplaceAll(path, ":", "")
	cleanPath = strings.ReplaceAll(cleanPath, "*", "")
	cleanPath = strings.ReplaceAll(cleanPath, "/", "_")
	cleanPath = strings.Trim(cleanPath, "_")

	if cleanPath == "" {
		cleanPath = "root"
	}

	return strings.ToLower(method) + "_" + cleanPath
}

// GenerateFromStruct 从结构体生成Schema
func (o *OpenAPI) GenerateFromStruct(v interface{}) *Schema {
	return o.generateSchemaFromType(reflect.TypeOf(v))
}

// generateSchemaFromType 从类型生成Schema
func (o *OpenAPI) generateSchemaFromType(t reflect.Type) *Schema {
	if t == nil {
		return &Schema{Type: "object"}
	}

	// 处理指针类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Schema{Type: "integer"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer", Minimum: &[]float64{0}[0]}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Slice, reflect.Array:
		return &Schema{
			Type:  "array",
			Items: o.generateSchemaFromType(t.Elem()),
		}
	case reflect.Map:
		return &Schema{
			Type:                 "object",
			AdditionalProperties: o.generateSchemaFromType(t.Elem()),
		}
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return &Schema{Type: "string", Format: "date-time"}
		}

		schema := &Schema{
			Type:       "object",
			Properties: make(map[string]*Schema),
		}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			fieldName := field.Name
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" {
					fieldName = parts[0]
				}
			}

			fieldSchema := o.generateSchemaFromType(field.Type)

			// 添加描述
			if desc := field.Tag.Get("description"); desc != "" {
				fieldSchema.Description = desc
			}

			// 添加示例
			if example := field.Tag.Get("example"); example != "" {
				fieldSchema.Example = example
			}

			schema.Properties[fieldName] = fieldSchema
		}

		return schema
	default:
		return &Schema{Type: "object"}
	}
}

// ToJSON 转换为JSON
func (o *OpenAPI) ToJSON() ([]byte, error) {
	return json.MarshalIndent(o, "", "  ")
}

// ServeSwaggerUI 提供Swagger UI服务
func (o *OpenAPI) ServeSwaggerUI() middleware.HandlerFunc {
	return func(c *gyarn.Context) {
		path := c.Request.URL.Path

		// 处理OpenAPI JSON请求
		if strings.HasSuffix(path, "/openapi.json") {
			c.Header("Content-Type", "application/json")
			o.buildPaths() // 关键：确保路径已构建
			jsonData, err := o.ToJSON()
			if err != nil {
				c.String(http.StatusInternalServerError, "生成OpenAPI文档失败: %v", err)
				return
			}
			c.Data(http.StatusOK, "application/json", jsonData)
			return
		}

		// 提供Swagger UI HTML
		o.buildPaths() // 关键：确保路径已构建
		html := o.generateSwaggerHTML()
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
	}
}

// generateSwaggerHTML 生成Swagger UI HTML
func (o *OpenAPI) generateSwaggerHTML() string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>%s - API文档</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger/openapi.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                docExpansion: "list",
                defaultModelsExpandDepth: 1,
                defaultModelExpandDepth: 1
            });
        };
    </script>
</body>
</html>`, o.Info.Title)
}
