# 附加服务集成指南

GyWeb 框架新增了微信公众号和Excel导入导出服务，进一步扩展了框架的功能范围。

## 概述

本框架新增以下服务：

- **微信公众号服务**: 用户管理、消息推送、菜单管理、二维码生成、网页授权
- **Excel服务**: Excel文件的导入导出、数据映射、样式配置、模板生成

## 微信公众号集成

### 配置

```go
import "github.com/guyigood/gyweb/core/services/wechat"

config := &wechat.WechatConfig{
    AppID:          "your_wechat_app_id",     // 公众号AppID
    AppSecret:      "your_wechat_app_secret", // 公众号AppSecret
    Token:          "your_wechat_token",      // 消息校验Token
    EncodingAESKey: "your_encoding_aes_key",  // 消息加解密密钥（可选）
}

wechatClient := wechat.NewWechat(config)
```

### 功能说明

#### 1. 服务器验证

```go
// 验证微信服务器
func verifyWechatServer(c *gyarn.Context) {
    signature := c.Query("signature")
    timestamp := c.Query("timestamp")
    nonce := c.Query("nonce")
    echostr := c.Query("echostr")

    if wechatClient.VerifySignature(signature, timestamp, nonce) {
        c.String(200, echostr)
    } else {
        c.String(403, "验证失败")
    }
}
```

#### 2. 用户管理

```go
// 获取用户信息
userInfo, err := wechatClient.GetUserInfo("user_openid")
if err != nil {
    log.Printf("获取用户信息失败: %v", err)
} else {
    fmt.Printf("用户昵称: %s\n", userInfo.Nickname)
    fmt.Printf("用户性别: %d\n", userInfo.Sex)
    fmt.Printf("用户城市: %s\n", userInfo.City)
}

// 获取用户列表
userList, err := wechatClient.GetUserList("")
if err != nil {
    log.Printf("获取用户列表失败: %v", err)
} else {
    fmt.Printf("总用户数: %d\n", userList.Total)
    for _, openid := range userList.Data.OpenID {
        fmt.Printf("用户OpenID: %s\n", openid)
    }
}
```

#### 3. 消息推送

**模板消息**

```go
msg := &wechat.TemplateMessage{
    ToUser:     "user_openid",
    TemplateID: "template_id",
    URL:        "https://example.com/detail",
    Data: map[string]interface{}{
        "first": map[string]string{
            "value": "您有新的订单",
            "color": "#173177",
        },
        "keyword1": map[string]string{
            "value": "订单号123456",
            "color": "#173177",
        },
        "keyword2": map[string]string{
            "value": "2024-01-01 12:00:00",
            "color": "#173177",
        },
        "remark": map[string]string{
            "value": "请及时处理",
            "color": "#173177",
        },
    },
}

resp, err := wechatClient.SendTemplateMessage(msg)
if err != nil {
    log.Printf("发送模板消息失败: %v", err)
} else {
    fmt.Printf("消息ID: %d\n", resp.MsgID)
}
```

**客服消息**

```go
// 发送文本消息
resp, err := wechatClient.SendTextMessage("user_openid", "您好，欢迎关注我们的公众号！")

// 发送图文消息
articles := []wechat.Article{
    {
        Title:       "文章标题",
        Description: "文章描述",
        URL:         "https://example.com/article",
        PicURL:      "https://example.com/image.jpg",
    },
}
resp, err = wechatClient.SendNewsMessage("user_openid", articles)
```

#### 4. 二维码生成

```go
// 临时二维码
qrReq := &wechat.QRCodeRequest{
    ExpireSeconds: 604800, // 7天
    ActionName:    "QR_SCENE",
    ActionInfo: &wechat.ActionInfo{
        Scene: &wechat.Scene{
            SceneID: 123,
        },
    },
}

qrResp, err := wechatClient.CreateQRCode(qrReq)
if err != nil {
    log.Printf("生成二维码失败: %v", err)
} else {
    // 获取二维码图片
    qrImage, err := wechatClient.GetQRCodeImage(qrResp.Ticket)
    if err == nil {
        // 保存或返回二维码图片
        fmt.Printf("二维码URL: %s\n", qrResp.URL)
    }
}

// 永久二维码
qrReq = &wechat.QRCodeRequest{
    ActionName: "QR_LIMIT_STR_SCENE",
    ActionInfo: &wechat.ActionInfo{
        Scene: &wechat.Scene{
            SceneStr: "user_123",
        },
    },
}
```

#### 5. 菜单管理

```go
// 创建自定义菜单
menu := &wechat.Menu{
    Button: []wechat.Button{
        {
            Name: "公司介绍",
            Type: "view",
            URL:  "https://example.com/about",
        },
        {
            Name: "产品服务",
            SubButton: []wechat.Button{
                {
                    Name: "产品A",
                    Type: "view",
                    URL:  "https://example.com/product-a",
                },
                {
                    Name: "产品B",
                    Type: "click",
                    Key:  "PRODUCT_B",
                },
            },
        },
        {
            Name:     "小程序",
            Type:     "miniprogram",
            AppID:    "miniprogram_app_id",
            PagePath: "pages/index/index",
            URL:      "https://example.com/fallback",
        },
    },
}

resp, err := wechatClient.CreateMenu(menu)
if err != nil {
    log.Printf("创建菜单失败: %v", err)
}

// 删除菜单
resp, err = wechatClient.DeleteMenu()
```

#### 6. 网页授权

```go
// 第一步：生成授权链接
redirectURI := "https://your-domain.com/oauth/callback"
scope := "snsapi_userinfo" // 或 "snsapi_base"
state := "custom_state"
oauthURL := wechatClient.GetOAuthURL(redirectURI, state, scope)

// 第二步：处理授权回调
func handleOAuthCallback(c *gyarn.Context) {
    code := c.Query("code")
    state := c.Query("state")
    
    // 获取access_token
    oauthToken, err := wechatClient.GetOAuthAccessToken(code)
    if err != nil {
        c.Error(500, "获取access token失败")
        return
    }
    
    // 获取用户信息（仅snsapi_userinfo作用域）
    if oauthToken.Scope == "snsapi_userinfo" {
        userInfo, err := wechatClient.GetOAuthUserInfo(oauthToken.AccessToken, oauthToken.OpenID)
        if err != nil {
            c.Error(500, "获取用户信息失败")
            return
        }
        
        c.Success(gyarn.H{
            "openid":   userInfo.OpenID,
            "nickname": userInfo.Nickname,
            "avatar":   userInfo.Headimgurl,
        })
    } else {
        c.Success(gyarn.H{
            "openid": oauthToken.OpenID,
        })
    }
}
```

## Excel服务集成

### 配置

Excel服务无需特殊配置，直接使用即可：

```go
import "github.com/guyigood/gyweb/core/services/excel"

// 创建新的Excel服务
excelService := excel.NewExcelService()
defer excelService.Close()

// 从文件创建
excelService, err := excel.NewExcelServiceWithFile("data.xlsx")
if err != nil {
    log.Fatal(err)
}
defer excelService.Close()

// 从字节数据创建
excelService, err := excel.NewExcelServiceWithReader(fileData)
if err != nil {
    log.Fatal(err)
}
defer excelService.Close()
```

### 功能说明

#### 1. 数据导入

**定义数据结构**

```go
type User struct {
    ID       int       `json:"id" excel:"ID"`
    Name     string    `json:"name" excel:"姓名"`
    Email    string    `json:"email" excel:"邮箱"`
    Age      int       `json:"age" excel:"年龄"`
    IsActive bool      `json:"is_active" excel:"是否激活"`
    CreateAt time.Time `json:"create_at" excel:"创建时间"`
}
```

**配置导入选项**

```go
importOptions := &excel.ImportOptions{
    SheetName: "用户数据",  // 工作表名称，空值表示第一个工作表
    StartRow:  2,          // 数据开始行号（跳过标题行）
    HeaderRow: 1,          // 标题行号
    MaxRows:   1000,       // 最大导入行数，0表示无限制
    ColumnMaps: []excel.ColumnMap{
        {
            Name:     "ID",       // Excel列名
            Field:    "ID",       // 结构体字段名
            Required: true,       // 是否必填
            DataType: "int",      // 数据类型
        },
        {
            Name:     "姓名",
            Field:    "Name",
            Required: true,
            DataType: "string",
        },
        {
            Name:     "邮箱",
            Field:    "Email",
            Required: false,
            DataType: "string",
        },
        {
            Name:     "年龄",
            Field:    "Age",
            Required: false,
            DataType: "int",
            Default:  18,         // 默认值
        },
        {
            Name:     "是否激活",
            Field:    "IsActive",
            Required: false,
            DataType: "bool",
        },
        {
            Name:     "创建时间",
            Field:    "CreateAt",
            Required: false,
            DataType: "time",
            Format:   "2006-01-02", // 时间格式
        },
    },
    ValidateFunc: func(data interface{}) error {
        user := data.(*User)
        if user.Age < 0 || user.Age > 150 {
            return fmt.Errorf("年龄必须在0-150之间")
        }
        if user.Email != "" && !strings.Contains(user.Email, "@") {
            return fmt.Errorf("邮箱格式不正确")
        }
        return nil
    },
}
```

**执行导入**

```go
var users []User
result, err := excelService.ImportData(importOptions, &users)
if err != nil {
    log.Printf("导入失败: %v", err)
    return
}

fmt.Printf("成功导入: %d 条\n", result.SuccessCount)
fmt.Printf("失败: %d 条\n", result.ErrorCount)

// 输出错误信息
for _, validationError := range result.Errors {
    fmt.Printf("第%d行，列%s：%s\n", 
        validationError.Row, 
        validationError.Column, 
        validationError.Message)
}

// 使用导入的数据
for _, user := range users {
    fmt.Printf("用户: %s, 邮箱: %s\n", user.Name, user.Email)
}
```

#### 2. 数据导出

**配置导出选项**

```go
exportOptions := &excel.ExportOptions{
    SheetName: "用户列表",
    Headers:   []string{"ID", "姓名", "邮箱", "年龄", "是否激活", "创建时间"},
    ColumnMaps: []excel.ColumnMap{
        {Name: "ID", Field: "ID", DataType: "int"},
        {Name: "姓名", Field: "Name", DataType: "string"},
        {Name: "邮箱", Field: "Email", DataType: "string"},
        {Name: "年龄", Field: "Age", DataType: "int"},
        {Name: "是否激活", Field: "IsActive", DataType: "bool"},
        {Name: "创建时间", Field: "CreateAt", DataType: "time", Format: "2006-01-02 15:04:05"},
    },
    StyleConfig: &excel.StyleConfig{
        HeaderStyle: &excel.CellStyle{
            Font: &excel.FontStyle{
                Bold:  true,
                Size:  12,
                Color: "#FFFFFF",
            },
            Fill: &excel.FillStyle{
                Type:    "pattern",
                Pattern: 1,
                Color:   "#4472C4",
            },
            Alignment: &excel.AlignmentStyle{
                Horizontal: "center",
                Vertical:   "center",
            },
        },
        DataStyle: &excel.CellStyle{
            Alignment: &excel.AlignmentStyle{
                Horizontal: "left",
                Vertical:   "center",
            },
            Border: &excel.BorderStyle{
                Type:  "thin",
                Color: "#000000",
            },
        },
        ColumnWidths: []float64{10, 15, 25, 10, 12, 20}, // 列宽
    },
    FormatFunc: func(data interface{}) interface{} {
        // 自定义数据格式化
        if user, ok := data.(User); ok {
            // 可以在这里对数据进行格式化处理
            return user
        }
        return data
    },
}
```

**执行导出**

```go
users := []User{
    {
        ID:       1,
        Name:     "张三",
        Email:    "zhangsan@example.com",
        Age:      25,
        IsActive: true,
        CreateAt: time.Now(),
    },
    {
        ID:       2,
        Name:     "李四",
        Email:    "lisi@example.com",
        Age:      30,
        IsActive: false,
        CreateAt: time.Now().AddDate(0, -1, 0),
    },
}

// 导出数据
err := excelService.ExportData(users, exportOptions)
if err != nil {
    log.Printf("导出失败: %v", err)
    return
}

// 保存到文件
err = excelService.SaveToFile("users.xlsx")
if err != nil {
    log.Printf("保存文件失败: %v", err)
    return
}

// 或获取字节数据
fileData, err := excelService.GetBytes()
if err != nil {
    log.Printf("获取数据失败: %v", err)
    return
}

// 用于HTTP响应
c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
c.Header("Content-Disposition", "attachment; filename=users.xlsx")
c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)
```

#### 3. 工作表操作

```go
// 获取所有工作表名称
sheetNames := excelService.GetSheetNames()
for _, name := range sheetNames {
    fmt.Printf("工作表: %s\n", name)
}

// 添加工作表
err := excelService.AddSheet("新工作表")

// 删除工作表
err = excelService.DeleteSheet("Sheet1")

// 设置活动工作表
err = excelService.SetActiveSheet("用户数据")

// 获取单元格值
value, err := excelService.GetCellValue("Sheet1", "A1")

// 设置单元格值
err = excelService.SetCellValue("Sheet1", "A1", "标题")

// 获取所有行数据
rows, err := excelService.GetRows("Sheet1")
for i, row := range rows {
    fmt.Printf("第%d行: %v\n", i+1, row)
}

// 获取所有列数据
cols, err := excelService.GetCols("Sheet1")
for i, col := range cols {
    fmt.Printf("第%d列: %v\n", i+1, col)
}
```

#### 4. 样式配置详解

**字体样式**

```go
fontStyle := &excel.FontStyle{
    Bold:   true,           // 粗体
    Italic: false,          // 斜体
    Size:   12,             // 字体大小
    Color:  "#FF0000",      // 字体颜色（十六进制）
    Family: "微软雅黑",        // 字体族
}
```

**填充样式**

```go
fillStyle := &excel.FillStyle{
    Type:    "pattern",     // 填充类型
    Pattern: 1,             // 图案（1为实心）
    Color:   "#FFFF00",     // 填充颜色
}
```

**对齐样式**

```go
alignmentStyle := &excel.AlignmentStyle{
    Horizontal: "center",   // 水平对齐：left, center, right
    Vertical:   "center",   // 垂直对齐：top, center, bottom
    WrapText:   true,       // 文本换行
}
```

**边框样式**

```go
borderStyle := &excel.BorderStyle{
    Type:  "thin",          // 边框类型：thin, thick, double
    Color: "#000000",       // 边框颜色
}
```

## API 路由示例

### 微信公众号路由

```go
wechatGroup := r.Group("/api/wechat")
{
    // 服务器验证
    wechatGroup.GET("/verify", verifyWechatServer)
    wechatGroup.POST("/verify", handleWechatMessage)
    
    // 用户管理
    wechatGroup.GET("/user/:openid", getUserInfo)
    wechatGroup.GET("/users", getUserList)
    
    // 消息推送
    wechatGroup.POST("/message/template", sendTemplateMessage)
    wechatGroup.POST("/message/text", sendTextMessage)
    wechatGroup.POST("/message/news", sendNewsMessage)
    
    // 二维码
    wechatGroup.POST("/qrcode", createQRCode)
    wechatGroup.GET("/qrcode/image", getQRCodeImage)
    
    // 菜单管理
    wechatGroup.POST("/menu", createMenu)
    wechatGroup.DELETE("/menu", deleteMenu)
    
    // 网页授权
    wechatGroup.GET("/oauth", handleOAuth)
    wechatGroup.GET("/oauth/callback", handleOAuthCallback)
}
```

### Excel路由

```go
excelGroup := r.Group("/api/excel")
{
    // 数据导入导出
    excelGroup.POST("/import", importExcelData)
    excelGroup.GET("/export", exportExcelData)
    
    // 模板管理
    excelGroup.POST("/template", generateTemplate)
    excelGroup.GET("/template/:name", downloadTemplate)
    
    // 文件操作
    excelGroup.POST("/upload", uploadExcelFile)
    excelGroup.GET("/download/:id", downloadExcelFile)
    
    // 数据预览
    excelGroup.POST("/preview", previewExcelData)
    excelGroup.GET("/sheets", getSheetNames)
}
```

## 安全建议

### 微信公众号安全

1. **Token安全**
   - 使用复杂的Token值
   - 定期更换Token
   - 验证消息来源

2. **用户隐私**
   - 合规获取用户信息
   - 安全存储用户数据
   - 遵循数据保护法规

3. **接口安全**
   - 实现访问频率限制
   - 验证请求参数
   - 记录操作日志

### Excel服务安全

1. **文件上传安全**
   - 限制文件大小
   - 验证文件类型
   - 扫描恶意内容

2. **数据处理安全**
   - 验证导入数据
   - 防止内存溢出
   - 处理异常情况

3. **权限控制**
   - 验证用户权限
   - 限制访问范围
   - 审计操作记录

## 错误处理

### 微信公众号错误

```go
// 统一错误处理
func handleWechatError(err error, operation string) {
    if err != nil {
        log.Printf("微信API错误 [%s]: %v", operation, err)
        // 根据错误类型进行相应处理
        if strings.Contains(err.Error(), "access_token") {
            // Access Token过期，重新获取
            wechatClient.GetAccessToken()
        }
    }
}
```

### Excel服务错误

```go
// Excel操作错误处理
func handleExcelError(err error, filename string) {
    if err != nil {
        log.Printf("Excel操作错误 [%s]: %v", filename, err)
        // 根据错误类型处理
        if strings.Contains(err.Error(), "not found") {
            // 文件不存在
        } else if strings.Contains(err.Error(), "invalid") {
            // 文件格式错误
        }
    }
}
```

## 最佳实践

1. **性能优化**
   - 使用连接池
   - 缓存访问令牌
   - 异步处理大文件

2. **错误恢复**
   - 实现重试机制
   - 记录详细日志
   - 提供友好错误信息

3. **监控告警**
   - 监控API调用频率
   - 跟踪错误率
   - 设置性能指标

4. **测试策略**
   - 单元测试覆盖
   - 集成测试验证
   - 压力测试评估

## 示例项目

完整的示例代码请参考：`examples/additional_services_example.go`

该示例包含了微信公众号和Excel服务的完整集成代码，展示了各种功能的具体使用方法。 