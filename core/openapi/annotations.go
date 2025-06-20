package openapi

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// AnnotationParser 注解解析器
type AnnotationParser struct {
	openapi    *OpenAPI
	fileSet    *token.FileSet
	modelCache map[string]*Schema // 缓存已解析的model
	packageMap map[string]string  // 包名映射
}

// NewAnnotationParser 创建注解解析器
func NewAnnotationParser(openapi *OpenAPI) *AnnotationParser {
	return &AnnotationParser{
		openapi:    openapi,
		fileSet:    token.NewFileSet(),
		modelCache: make(map[string]*Schema),
		packageMap: make(map[string]string),
	}
}

// RecursiveParseDirectory 递归解析目录中的Go文件，支持跨文件model发现
func (p *AnnotationParser) RecursiveParseDirectory(rootDir string) error {
	// 第一阶段：扫描所有Go文件，收集所有结构体定义
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过隐藏目录和vendor目录
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// 只处理Go文件
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		return p.scanFileForModels(path)
	})

	if err != nil {
		return err
	}

	// 第二阶段：解析注解和函数
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		return p.ParseFile(path)
	})
}

// scanFileForModels 扫描文件中的所有模型定义
func (p *AnnotationParser) scanFileForModels(filename string) error {
	src, err := parser.ParseFile(p.fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// 记录包名
	packageName := src.Name.Name
	p.packageMap[filename] = packageName

	// 扫描所有结构体类型
	for _, decl := range src.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						structName := typeSpec.Name.Name

						// 生成Schema并缓存
						schema := p.generateSchemaFromStruct(structName, structType)
						p.modelCache[structName] = schema

						// 注册到OpenAPI Components
						p.openapi.AddSchema(structName, schema)

						fmt.Printf("[DEBUG] 发现并注册模型: %s\n", structName)
					}
				}
			}
		}
	}

	return nil
}

// ParseDirectory 解析目录中的Go文件
func (p *AnnotationParser) ParseDirectory(dir string) error {
	return p.RecursiveParseDirectory(dir)
}

// ParseFile 解析单个Go文件
func (p *AnnotationParser) ParseFile(filename string) error {
	src, err := parser.ParseFile(p.fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// 先扫描所有结构体类型
	p.scanStructTypes(src)

	// 然后遍历所有函数
	for _, decl := range src.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			p.parseFunction(fn)
		}
	}

	return nil
}

// scanStructTypes 扫描文件中的所有结构体类型
func (p *AnnotationParser) scanStructTypes(file *ast.File) {
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						// 找到结构体定义，生成Schema并注册
						schema := p.generateSchemaFromStruct(typeSpec.Name.Name, structType)
						p.openapi.AddSchema(typeSpec.Name.Name, schema)
					}
				}
			}
		}
	}
}

// generateSchemaFromStruct 从AST结构体生成Schema
func (p *AnnotationParser) generateSchemaFromStruct(name string, structType *ast.StructType) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
		Required:   []string{},
	}

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			// 处理嵌入字段
			if ident, ok := field.Type.(*ast.Ident); ok {
				// 嵌入结构体，需要合并其字段
				if embeddedSchema, exists := p.modelCache[ident.Name]; exists && embeddedSchema.Properties != nil {
					for propName, propSchema := range embeddedSchema.Properties {
						schema.Properties[propName] = propSchema
					}
				}
			}
			continue
		}

		fieldName := field.Names[0].Name
		if !field.Names[0].IsExported() {
			continue // 跳过未导出字段
		}

		// 解析字段标签
		var jsonName string
		var description string
		var example interface{}
		var required bool

		if field.Tag != nil {
			tagValue := field.Tag.Value
			if tagValue != "" {
				// 移除反引号
				tagValue = strings.Trim(tagValue, "`")

				// 解析json标签
				if jsonMatch := regexp.MustCompile(`json:"([^"]+)"`).FindStringSubmatch(tagValue); len(jsonMatch) > 1 {
					jsonParts := strings.Split(jsonMatch[1], ",")
					jsonName = jsonParts[0]

					// 检查是否必需（没有omitempty）
					required = true
					for _, part := range jsonParts {
						if part == "omitempty" {
							required = false
							break
						}
					}
				}

				// 解析description标签
				if descMatch := regexp.MustCompile(`description:"([^"]+)"`).FindStringSubmatch(tagValue); len(descMatch) > 1 {
					description = descMatch[1]
				}

				// 解析example标签
				if exampleMatch := regexp.MustCompile(`example:"([^"]+)"`).FindStringSubmatch(tagValue); len(exampleMatch) > 1 {
					example = exampleMatch[1]
				}

				// 解析binding标签确定必需字段
				if bindingMatch := regexp.MustCompile(`binding:"([^"]+)"`).FindStringSubmatch(tagValue); len(bindingMatch) > 1 {
					if strings.Contains(bindingMatch[1], "required") {
						required = true
					}
				}
			}
		}

		if jsonName == "" {
			jsonName = fieldName
		}

		// 跳过被忽略的字段
		if jsonName == "-" {
			continue
		}

		// 生成字段Schema
		fieldSchema := p.generateSchemaFromType(field.Type)
		if description != "" {
			fieldSchema.Description = description
		}
		if example != nil {
			fieldSchema.Example = example
		}

		schema.Properties[jsonName] = fieldSchema

		// 添加到必需字段列表
		if required {
			schema.Required = append(schema.Required, jsonName)
		}
	}

	return schema
}

// generateSchemaFromType 从AST类型生成Schema
func (p *AnnotationParser) generateSchemaFromType(expr ast.Expr) *Schema {
	return p.generateSchemaFromTypeInline(expr, true)
}

// generateSchemaFromTypeInline 从AST类型生成Schema，支持内联选项
func (p *AnnotationParser) generateSchemaFromTypeInline(expr ast.Expr, inline bool) *Schema {
	switch t := expr.(type) {
	case *ast.Ident:
		// 基本类型
		switch t.Name {
		case "string":
			return &Schema{Type: "string"}
		case "int", "int8", "int16", "int32", "int64":
			return &Schema{Type: "integer", Format: "int64"}
		case "uint", "uint8", "uint16", "uint32", "uint64":
			return &Schema{Type: "integer", Format: "int64", Minimum: &[]float64{0}[0]}
		case "float32":
			return &Schema{Type: "number", Format: "float"}
		case "float64":
			return &Schema{Type: "number", Format: "double"}
		case "bool":
			return &Schema{Type: "boolean"}
		default:
			// 结构体类型处理
			if cachedSchema, exists := p.modelCache[t.Name]; exists {
				if inline {
					// 返回内联展开的schema，深拷贝以避免修改原始cache
					return p.deepCopySchema(cachedSchema)
				} else {
					// 返回引用
					return &Schema{Ref: fmt.Sprintf("#/components/schemas/%s", t.Name)}
				}
			}
			// 未知结构体类型，也尝试内联展开
			if inline {
				return &Schema{Type: "object", Properties: make(map[string]*Schema)}
			} else {
				return &Schema{Ref: fmt.Sprintf("#/components/schemas/%s", t.Name)}
			}
		}
	case *ast.ArrayType:
		// 数组类型
		itemSchema := p.generateSchemaFromTypeInline(t.Elt, inline)
		return &Schema{
			Type:  "array",
			Items: itemSchema,
		}
	case *ast.StarExpr:
		// 指针类型，递归处理
		return p.generateSchemaFromTypeInline(t.X, inline)
	case *ast.MapType:
		// Map类型
		return &Schema{
			Type:                 "object",
			AdditionalProperties: p.generateSchemaFromTypeInline(t.Value, inline),
		}
	case *ast.SelectorExpr:
		// 包.类型格式，如 time.Time
		if ident, ok := t.X.(*ast.Ident); ok {
			if ident.Name == "time" && t.Sel.Name == "Time" {
				return &Schema{Type: "string", Format: "date-time"}
			}
		}
		return &Schema{Type: "object"}
	default:
		// 未知类型，返回object
		return &Schema{Type: "object"}
	}
}

// deepCopySchema 深拷贝Schema对象
func (p *AnnotationParser) deepCopySchema(original *Schema) *Schema {
	if original == nil {
		return nil
	}

	copy := &Schema{
		Type:        original.Type,
		Format:      original.Format,
		Title:       original.Title,
		Description: original.Description,
		Default:     original.Default,
		Example:     original.Example,
		Pattern:     original.Pattern,
		Minimum:     original.Minimum,
		Maximum:     original.Maximum,
		MinLength:   original.MinLength,
		MaxLength:   original.MaxLength,
	}

	// 拷贝Required数组
	if original.Required != nil {
		copy.Required = make([]string, len(original.Required))
		copy.Required = append(copy.Required, original.Required...)
	}

	// 拷贝Enum数组
	if original.Enum != nil {
		copy.Enum = make([]interface{}, len(original.Enum))
		copy.Enum = append(copy.Enum, original.Enum...)
	}

	// 递归拷贝Properties
	if original.Properties != nil {
		copy.Properties = make(map[string]*Schema)
		for key, value := range original.Properties {
			copy.Properties[key] = p.deepCopySchema(value)
		}
	}

	// 递归拷贝Items
	if original.Items != nil {
		copy.Items = p.deepCopySchema(original.Items)
	}

	// 处理AdditionalProperties
	if original.AdditionalProperties != nil {
		if schema, ok := original.AdditionalProperties.(*Schema); ok {
			copy.AdditionalProperties = p.deepCopySchema(schema)
		} else {
			copy.AdditionalProperties = original.AdditionalProperties
		}
	}

	return copy
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
// 支持多种格式：
// @Param name query string true "参数说明"
// @Param data body User true "用户信息"
// @Param data body dto.LoginRequest true "登录请求参数"
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

	// 特殊处理body参数
	if param.In == "body" {
		// body参数需要特殊处理，直接使用内联schema
		modelName := paramType

		// 移除包名前缀
		if dotIndex := strings.LastIndex(modelName, "."); dotIndex != -1 {
			modelName = modelName[dotIndex+1:]
		}

		// 查找对应的schema并内联展开
		if cachedSchema, exists := p.modelCache[modelName]; exists {
			param.Schema = p.deepCopySchema(cachedSchema)
		} else {
			param.Schema = &Schema{Type: "object"}
		}
	} else {
		// 非body参数使用简单类型转换
		param.Schema = &Schema{Type: p.convertType(paramType)}
	}

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
		if param.Schema != nil {
			param.Schema.Default = defaultMatch[1]
		}
	}

	// 解析示例
	if exampleMatch := regexp.MustCompile(`example\(([^)]+)\)`).FindStringSubmatch(content); len(exampleMatch) > 1 {
		param.Example = exampleMatch[1]
	}

	return param
}

// parseResponse 解析响应注解
// 支持多种格式：
// @Success 200 {object} User "成功返回用户信息"
// @Success 200 {object} dto.StandardResponse{data=User} "成功"
// @Success 200 {array} User "返回用户列表"
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

	// 解析复杂的响应体类型，支持嵌套格式
	// 匹配格式如: {object} dto.StandardResponse{data=User}
	complexMatch := regexp.MustCompile(`\{([^}]+)\}\s+([^"]+)`).FindStringSubmatch(line)
	if len(complexMatch) >= 3 {
		responseType := complexMatch[1]
		modelSpec := strings.TrimSpace(complexMatch[2])

		var schema *Schema

		if strings.Contains(modelSpec, "{") && strings.Contains(modelSpec, "}") {
			// 处理嵌套结构，如 dto.StandardResponse{data=User}
			schema = p.parseNestedResponseSchema(responseType, modelSpec)
		} else {
			// 简单类型，如 User
			schema = p.parseSimpleResponseSchema(responseType, modelSpec)
		}

		response.Content["application/json"] = MediaType{
			Schema: schema,
		}
	}

	return response
}

// parseNestedResponseSchema 解析嵌套的响应Schema
// 处理如 dto.StandardResponse{data=User} 这样的格式
func (p *AnnotationParser) parseNestedResponseSchema(responseType, modelSpec string) *Schema {
	// 提取基础类型名
	baseTypeMatch := regexp.MustCompile(`^([^{]+)`).FindStringSubmatch(modelSpec)
	if len(baseTypeMatch) < 2 {
		return &Schema{Type: "object"}
	}

	baseTypeName := strings.TrimSpace(baseTypeMatch[1])
	// 移除包名前缀，如 dto.StandardResponse -> StandardResponse
	if dotIndex := strings.LastIndex(baseTypeName, "."); dotIndex != -1 {
		baseTypeName = baseTypeName[dotIndex+1:]
	}

	// 获取基础Schema
	var baseSchema *Schema
	if cachedSchema, exists := p.modelCache[baseTypeName]; exists {
		baseSchema = p.deepCopySchema(cachedSchema)
	} else {
		baseSchema = &Schema{Type: "object", Properties: make(map[string]*Schema)}
	}

	// 解析嵌套部分，如 {data=User}
	nestedMatch := regexp.MustCompile(`\{([^}]+)\}`).FindStringSubmatch(modelSpec)
	if len(nestedMatch) >= 2 {
		nestedContent := nestedMatch[1]

		// 解析键值对，如 data=User
		pairs := strings.Split(nestedContent, ",")
		for _, pair := range pairs {
			if keyValue := strings.Split(strings.TrimSpace(pair), "="); len(keyValue) == 2 {
				key := strings.TrimSpace(keyValue[0])
				valueType := strings.TrimSpace(keyValue[1])

				// 确保baseSchema有Properties
				if baseSchema.Properties == nil {
					baseSchema.Properties = make(map[string]*Schema)
				}

				// 根据valueType创建内联schema
				var valueSchema *Schema
				if responseType == "array" {
					// 如果是数组类型
					if cachedSchema, exists := p.modelCache[valueType]; exists {
						valueSchema = &Schema{
							Type:  "array",
							Items: p.deepCopySchema(cachedSchema),
						}
					} else {
						valueSchema = &Schema{Type: "array", Items: &Schema{Type: "object"}}
					}
				} else {
					// 对象类型
					if cachedSchema, exists := p.modelCache[valueType]; exists {
						valueSchema = p.deepCopySchema(cachedSchema)
					} else {
						valueSchema = &Schema{Type: "object"}
					}
				}

				baseSchema.Properties[key] = valueSchema
			}
		}
	}

	return baseSchema
}

// parseSimpleResponseSchema 解析简单的响应Schema
func (p *AnnotationParser) parseSimpleResponseSchema(responseType, modelName string) *Schema {
	// 移除包名前缀
	if dotIndex := strings.LastIndex(modelName, "."); dotIndex != -1 {
		modelName = modelName[dotIndex+1:]
	}

	if responseType == "array" {
		// 数组类型
		if cachedSchema, exists := p.modelCache[modelName]; exists {
			return &Schema{
				Type:  "array",
				Items: p.deepCopySchema(cachedSchema),
			}
		} else {
			return &Schema{
				Type:  "array",
				Items: &Schema{Type: "object"},
			}
		}
	} else if responseType == "object" {
		// 对象类型，直接展开
		if cachedSchema, exists := p.modelCache[modelName]; exists {
			return p.deepCopySchema(cachedSchema)
		} else {
			return &Schema{Type: "object"}
		}
	} else {
		// 基本类型
		return &Schema{Type: p.convertType(responseType)}
	}
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
