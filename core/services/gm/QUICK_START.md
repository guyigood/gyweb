# 国密服务快速入门

## 5分钟快速体验

### 1. 基本安装

```go
// 直接导入使用，无需额外安装
import "gyweb/core/services/gm"
```

### 2. 最简示例

```go
package main

import (
    "fmt"
    "log"
    "gyweb/core/services/gm"
)

func main() {
    // 创建服务
    service, err := gm.NewGMServiceDefault()
    if err != nil {
        log.Fatal(err)
    }

    // SM2 加密
    resp, _ := service.SM2Encrypt([]byte("Hello 国密"))
    fmt.Printf("SM2密文: %s\n", resp.EncryptedData)

    // SM3 哈希
    hash, _ := service.SM3HashString("Hello 国密", "hex")
    fmt.Printf("SM3哈希: %s\n", hash.Hash)
}
```

### 3. 运行测试

```bash
cd core/services/gm
go test -v
```

## 常用场景

### 场景1: API 数据加密

```go
// 在 HTTP 处理器中加密敏感数据
func HandleUserData(c *gin.Context) {
    service, _ := gm.NewGMServiceDefault()
    
    userData := map[string]interface{}{
        "phone": "13800138000",
        "email": "user@example.com",
    }
    
    // 加密 JSON
    encrypted, err := service.EncryptJSON(userData, "SM2")
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"data": encrypted.EncryptedData})
}
```

### 场景2: 密码哈希存储

```go
// 用户密码哈希
func HashPassword(password string) (string, error) {
    hash, err := gm.QuickSM3HashString(password, "hex")
    return hash, err
}

// 验证密码
func VerifyPassword(password, hashedPassword string) bool {
    service, _ := gm.NewGMServiceDefault()
    result, _ := service.SM3Verify([]byte(password), hashedPassword, "hex")
    return result.Valid
}
```

### 场景3: 文件完整性检查

```go
// 计算文件哈希
func CalculateFileHash(filePath string) (string, error) {
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return "", err
    }
    
    return gm.QuickSM3Hash(data, "hex")
}
```

### 场景4: 配置文件加密

```go
// 加密配置
func EncryptConfig(config map[string]interface{}) (string, error) {
    service, _ := gm.NewGMService(&gm.GMConfig{
        DefaultSM4Key: "your-config-key-32-chars-long",
        OutputFormat:  "base64",
    })
    
    resp, err := service.EncryptJSON(config, "SM4")
    return resp.EncryptedData, err
}
```

## 性能对比

| 算法 | 用途 | 性能特点 |
|------|------|----------|
| SM2  | 公钥加密 | 适合小数据，密钥交换 |
| SM3  | 哈希 | 高性能，适合大数据 |
| SM4  | 对称加密 | 高性能，适合大数据 |

## 配置选择指南

```go
// 🔥 生产环境 - 安全优先
config := gm.GetGMConfigSecure()

// ⚡ 开发环境 - 性能优先  
config := gm.GetGMConfigPerformance()

// 🛠️ 自定义环境
config := gm.GetGMConfigCustom("hex", "your-key")
```

## 错误处理模板

```go
func SafeEncrypt(data []byte) (string, error) {
    service, err := gm.NewGMServiceDefault()
    if err != nil {
        return "", fmt.Errorf("创建服务失败: %w", err)
    }
    
    resp, err := service.SM2Encrypt(data)
    if err != nil {
        return "", fmt.Errorf("加密失败: %w", err)
    }
    
    return resp.EncryptedData, nil
}
```

## 最佳实践清单

### ✅ 推荐做法

- 始终检查错误返回值
- 使用配置管理密钥，不要硬编码
- 选择合适的输出格式（hex/base64）
- 定期轮换密钥

### ❌ 避免做法

- 忽略错误处理
- 使用弱密钥（如"123456"）
- 在日志中输出密钥
- 混合使用不同的输出格式

## 集成模板

### Gin 中间件模板

```go
func GMCryptoMiddleware() gin.HandlerFunc {
    service, _ := gm.NewGMServiceDefault()
    
    return func(c *gin.Context) {
        // 将服务注入到上下文
        c.Set("gm_service", service)
        c.Next()
    }
}

// 使用
func SomeHandler(c *gin.Context) {
    service := c.MustGet("gm_service").(*gm.GMService)
    // 使用 service 进行加密操作
}
```

### GORM 模型模板

```go
type User struct {
    ID       uint   `gorm:"primarykey"`
    Name     string
    Phone    string `gorm:"column:phone_encrypted"`
    service  *gm.GMService `gorm:"-"`
}

func (u *User) BeforeSave(tx *gorm.DB) error {
    if u.service == nil {
        u.service, _ = gm.NewGMServiceDefault()
    }
    
    if u.Phone != "" {
        resp, err := u.service.SM4Encrypt([]byte(u.Phone), defaultKey)
        if err != nil {
            return err
        }
        u.Phone = resp.EncryptedData
    }
    return nil
}
```

## 快速调试

### 问题诊断

```go
// 检查服务是否正常
func DiagnoseGMService() {
    service, err := gm.NewGMServiceDefault()
    if err != nil {
        fmt.Printf("❌ 服务创建失败: %v\n", err)
        return
    }
    
    // 测试 SM3
    hash, err := service.SM3HashString("test", "hex")
    if err != nil {
        fmt.Printf("❌ SM3测试失败: %v\n", err)
    } else {
        fmt.Printf("✅ SM3测试通过: %s\n", hash.Hash)
    }
    
    // 测试 SM2
    resp, err := service.SM2Encrypt([]byte("test"))
    if err != nil {
        fmt.Printf("❌ SM2测试失败: %v\n", err)
    } else {
        fmt.Printf("✅ SM2测试通过\n")
    }
    
    fmt.Println("🎉 服务诊断完成")
}
```

### 性能测试

```go
// 简单性能测试
func QuickBenchmark() {
    config := &gm.BenchmarkConfig{
        DataSize:   1024,
        Iterations: 100,
        Algorithm:  "SM4",
    }
    
    result, _ := gm.RunBenchmark(config)
    fmt.Printf("性能测试结果: %.2f MB/s\n", result.ThroughputMBps)
}
```

## 下一步

- 查看完整文档: [README.md](README.md)
- 运行完整示例: `go run examples/gm_service_demo.go`
- 查看测试用例: `go test -v core/services/gm`

## 常见问题

**Q: 密钥长度错误怎么办？**
A: SM4要求16字节密钥，检查密钥长度和格式

**Q: 加密后无法解密？**
A: 确保使用相同的密钥和输出格式

**Q: 性能如何优化？**
A: 使用SM4处理大数据，SM2处理小数据和密钥交换

**Q: 如何在生产环境部署？**
A: 使用环境变量管理密钥，启用安全配置 