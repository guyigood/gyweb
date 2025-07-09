# 验证码生成功能使用指南

## 简介

gyweb 框架提供了内置的验证码生成功能，位于 `core/utils/captcha` 包中。该功能支持生成带有干扰噪点和线条的验证码图片，返回 base64 编码的图片数据，可直接在 HTML 中使用。

## 快速开始

### 基础用法

```go
import "github.com/guyigood/gyweb/core/utils/captcha"

// 最简单的用法：快速生成4位混合字符验证码
text, imageData, err := captcha.QuickGenerate()
if err != nil {
    return err
}

// text: 验证码文本，如 "A8D2"
// imageData: 完整的 data URL，可直接用于 HTML <img> 标签
```

### 自定义文本

```go
// 使用指定文本生成验证码
imageData, err := captcha.GenerateCaptcha("HELLO", nil)
```

### 生成随机文本

```go
// 生成4位数字验证码
numberCode := captcha.GenerateRandomText(4, "number")  // 如: "1234"

// 生成6位字母验证码  
letterCode := captcha.GenerateRandomText(6, "letter") // 如: "ABCDEF"

// 生成5位混合验证码
mixedCode := captcha.GenerateRandomText(5, "mixed")   // 如: "A1B2C"
```

## 在 Web 应用中使用

### 后端代码示例

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla/sessions"
    "github.com/guyigood/gyweb/core/utils/captcha"
    "strings"
)

// 生成验证码接口
func GetCaptcha(c *gin.Context) {
    text, imageData, err := captcha.QuickGenerate()
    if err != nil {
        c.JSON(500, gin.H{"error": "生成验证码失败"})
        return
    }
    
    // 将验证码存储到 session
    session := sessions.Default(c)
    session.Set("captcha", text)
    session.Save()
    
    c.JSON(200, gin.H{
        "image": imageData,
    })
}

// 验证验证码接口
func VerifyCaptcha(c *gin.Context) {
    var req struct {
        Code string `json:"code"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "参数错误"})
        return
    }
    
    session := sessions.Default(c)
    savedCode := session.Get("captcha")
    
    if savedCode == nil {
        c.JSON(400, gin.H{"error": "验证码已过期"})
        return
    }
    
    // 不区分大小写比较
    if strings.ToUpper(req.Code) != strings.ToUpper(savedCode.(string)) {
        c.JSON(400, gin.H{"error": "验证码错误"})
        return
    }
    
    // 验证成功，清除验证码
    session.Delete("captcha")
    session.Save()
    
    c.JSON(200, gin.H{"message": "验证成功"})
}
```

### 前端代码示例

```html
<!DOCTYPE html>
<html>
<head>
    <title>验证码示例</title>
</head>
<body>
    <div>
        <img id="captcha-img" src="" alt="验证码" onclick="refreshCaptcha()" style="cursor: pointer;">
        <button onclick="refreshCaptcha()">刷新</button>
    </div>
    <div>
        <input type="text" id="captcha-input" placeholder="请输入验证码">
        <button onclick="verifyCaptcha()">验证</button>
    </div>

    <script>
        // 获取验证码
        function refreshCaptcha() {
            fetch('/api/captcha')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('captcha-img').src = data.image;
                });
        }
        
        // 验证验证码
        function verifyCaptcha() {
            const code = document.getElementById('captcha-input').value;
            fetch('/api/verify-captcha', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({code: code})
            })
            .then(response => response.json())
            .then(data => {
                alert(data.message || data.error);
                if (data.message) {
                    document.getElementById('captcha-input').value = '';
                    refreshCaptcha();
                }
            });
        }
        
        // 页面加载时获取验证码
        refreshCaptcha();
    </script>
</body>
</html>
```

## 自定义配置

```go
import "image/color"

config := &captcha.CaptchaConfig{
    Width:      150,    // 图片宽度
    Height:     50,     // 图片高度  
    NoiseCount: 80,     // 噪点数量
    NoiseLevel: 0.5,    // 噪声强度
    FontSize:   28,     // 字体大小
    BgColor:    color.RGBA{R: 255, G: 255, B: 255, A: 255}, // 白色背景
    TextColor:  color.RGBA{R: 255, G: 0, B: 0, A: 255},     // 红色文字
    NoiseColor: color.RGBA{R: 200, G: 200, B: 200, A: 100}, // 浅灰噪点
}

imageData, err := captcha.GenerateCaptcha("CUSTOM", config)
```

## API 参考

### 函数列表

- `QuickGenerate() (text, imageData string, err error)` - 快速生成4位混合验证码
- `GenerateCaptcha(text string, config *CaptchaConfig) (string, error)` - 自定义文本生成验证码
- `GenerateRandomText(length int, charType string) string` - 生成随机文本
- `DefaultConfig() *CaptchaConfig` - 获取默认配置

### 字符类型

- `"number"` - 纯数字 (0-9)
- `"letter"` - 纯字母 (A-Z)  
- `"mixed"` - 数字+字母 (默认)

## 最佳实践

1. **安全性**: 验证码应存储在服务端 session 中，不要在客户端存储
2. **时效性**: 设置验证码过期时间，建议5-10分钟
3. **大小写**: 验证时建议忽略大小写
4. **频率限制**: 对验证码生成和验证接口进行频率限制
5. **错误次数**: 多次验证失败后要求重新获取验证码

## 故障排除

1. **图片不显示**: 检查返回的 imageData 是否完整，应以 `data:image/png;base64,` 开头
2. **验证失败**: 确认大小写处理是否一致
3. **性能问题**: 如需大量生成验证码，考虑使用缓存或预生成策略

更多详细信息请参考 [完整文档](./core/utils/captcha/README.md)。 