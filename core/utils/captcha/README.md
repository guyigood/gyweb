# 验证码生成工具包

这个工具包提供了简单易用的验证码生成功能，可以根据输入的字符串生成验证码图片。

## 功能特性

- 🖼️ 支持自定义文本生成验证码图片
- 🎨 可配置图片尺寸、颜色、噪点等
- 🔤 支持数字、字母、混合字符类型
- 📋 返回base64编码的图片数据，可直接在HTML中使用
- ⚡ 提供快速生成接口，开箱即用

## 基本用法

### 1. 使用自定义文本生成验证码

```go
import "github.com/guyigood/gyweb/core/utils/captcha"

// 使用默认配置生成验证码
imageData, err := captcha.GenerateCaptcha("AB12", nil)
if err != nil {
    log.Fatal(err)
}

// imageData 是完整的 data URL，可以直接在HTML中使用
// 格式: "data:image/png;base64,iVBORw0KGgoAAAANSUhE..."
```

### 2. 快速生成随机验证码

```go
// 生成4位随机混合字符验证码
text, imageData, err := captcha.QuickGenerate()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("验证码文本: %s\n", text)
fmt.Printf("图片数据: %s\n", imageData)
```

### 3. 生成随机文本

```go
// 生成4位数字验证码
numberCode := captcha.GenerateRandomText(4, "number")

// 生成6位字母验证码
letterCode := captcha.GenerateRandomText(6, "letter")

// 生成5位混合验证码
mixedCode := captcha.GenerateRandomText(5, "mixed")
```

### 4. 自定义配置

```go
import "image/color"

// 创建自定义配置
config := &captcha.CaptchaConfig{
    Width:      150,
    Height:     50,
    NoiseCount: 80,
    NoiseLevel: 0.5,
    FontSize:   28,
    BgColor:    color.RGBA{R: 255, G: 255, B: 255, A: 255}, // 白色背景
    TextColor:  color.RGBA{R: 255, G: 0, B: 0, A: 255},     // 红色文字
    NoiseColor: color.RGBA{R: 200, G: 200, B: 200, A: 100}, // 浅灰噪点
}

// 使用自定义配置生成验证码
imageData, err := captcha.GenerateCaptcha("HELLO", config)
```

## 在Web应用中使用

### 1. 在控制器中生成验证码

```go
func GetCaptcha(c *gin.Context) {
    // 生成验证码
    text, imageData, err := captcha.QuickGenerate()
    if err != nil {
        c.JSON(500, gin.H{"error": "生成验证码失败"})
        return
    }
    
    // 将验证码文本存储到session中
    session := sessions.Default(c)
    session.Set("captcha", text)
    session.Save()
    
    // 返回图片数据
    c.JSON(200, gin.H{
        "image": imageData,
    })
}

func VerifyCaptcha(c *gin.Context) {
    var req struct {
        Code string `json:"code"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "参数错误"})
        return
    }
    
    // 从session获取验证码
    session := sessions.Default(c)
    savedCode := session.Get("captcha")
    
    if savedCode == nil || strings.ToUpper(req.Code) != strings.ToUpper(savedCode.(string)) {
        c.JSON(400, gin.H{"error": "验证码错误"})
        return
    }
    
    // 验证成功，清除session中的验证码
    session.Delete("captcha")
    session.Save()
    
    c.JSON(200, gin.H{"message": "验证成功"})
}
```

### 2. 在HTML中显示验证码

```html
<!DOCTYPE html>
<html>
<head>
    <title>验证码示例</title>
</head>
<body>
    <div>
        <img id="captcha-img" src="" alt="验证码" style="cursor: pointer;" onclick="refreshCaptcha()">
        <button onclick="refreshCaptcha()">刷新验证码</button>
    </div>
    
    <div>
        <input type="text" id="captcha-input" placeholder="请输入验证码">
        <button onclick="verifyCaptcha()">验证</button>
    </div>

    <script>
        // 加载验证码
        function refreshCaptcha() {
            fetch('/api/captcha')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('captcha-img').src = data.image;
                })
                .catch(error => console.error('Error:', error));
        }
        
        // 验证验证码
        function verifyCaptcha() {
            const code = document.getElementById('captcha-input').value;
            fetch('/api/verify-captcha', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({code: code})
            })
            .then(response => response.json())
            .then(data => {
                alert(data.message || data.error);
            });
        }
        
        // 页面加载时获取验证码
        refreshCaptcha();
    </script>
</body>
</html>
```

## API 参考

### 函数

#### `GenerateCaptcha(text string, config *CaptchaConfig) (string, error)`
根据指定文本生成验证码图片

- `text`: 要显示的验证码文本
- `config`: 配置参数，传nil使用默认配置
- 返回: base64编码的图片数据和错误信息

#### `GenerateRandomText(length int, charType string) string`
生成随机验证码文本

- `length`: 验证码长度
- `charType`: 字符类型，可选值：
  - `"number"`: 纯数字
  - `"letter"`: 纯字母
  - `"mixed"`: 数字+字母（默认）

#### `QuickGenerate() (text string, imageData string, err error)`
快速生成4位混合字符验证码

#### `DefaultConfig() *CaptchaConfig`
获取默认配置

### 配置结构

```go
type CaptchaConfig struct {
    Width       int        // 图片宽度 (默认: 120)
    Height      int        // 图片高度 (默认: 40)
    NoiseCount  int        // 噪点数量 (默认: 50)
    NoiseLevel  float64    // 噪声强度 (默认: 0.3)
    FontSize    int        // 字体大小 (默认: 24)
    BgColor     color.RGBA // 背景颜色
    TextColor   color.RGBA // 文字颜色
    NoiseColor  color.RGBA // 噪点颜色
}
```

## 注意事项

1. 验证码文本建议控制在6位以内，过长可能影响显示效果
2. 生成的图片数据是完整的data URL格式，可以直接在HTML的`<img>`标签中使用
3. 验证码应该存储在session中，并设置合理的过期时间
4. 验证时建议忽略大小写
5. 为防止暴力破解，建议对验证码验证频率进行限制

## 许可证

本工具包随 gyweb 框架一起发布，遵循相同的许可证。 