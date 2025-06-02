# 国密服务 (GM Service) 调用说明文档

## 概述

国密服务是 gyweb 框架中的第三方服务模块，提供中国国产密码算法（国密算法）的完整实现，包括 SM2（椭圆曲线公钥算法）、SM3（密码杂凑算法）和 SM4（分组密码算法）。

## 快速开始

### 1. 导入服务

```go
import "gyweb/core/services/gm"
```

### 2. 创建服务实例

```go
// 使用默认配置
service, err := gm.NewGMServiceDefault()
if err != nil {
    log.Fatal("创建国密服务失败:", err)
}

// 使用自定义配置
config := &gm.GMConfig{
    OutputFormat: "hex",
    DefaultSM4Key: "1234567890abcdef1234567890abcdef",
}
service, err := gm.NewGMService(config)
```

### 3. 基本使用

```go
// SM2 加密
text := "敏感数据"
resp, err := service.SM2Encrypt([]byte(text))

// SM4 加密
key := []byte("1234567890abcdef") // 16字节密钥
resp, err := service.SM4Encrypt([]byte(text), key)

// SM3 哈希
hashResp, err := service.SM3Hash([]byte(text))
```

## 详细 API 文档

### 服务配置

#### GMConfig 结构体

```go
type GMConfig struct {
    SM2PrivateKey         string `json:"sm2_private_key,omitempty"`  // SM2私钥
    SM2PublicKey          string `json:"sm2_public_key,omitempty"`   // SM2公钥
    DefaultSM4Key         string `json:"default_sm4_key,omitempty"`  // 默认SM4密钥
    OutputFormat          string `json:"output_format"`              // 输出格式: "base64", "hex"
    EnableSignatureVerify bool   `json:"enable_signature_verify"`    // 启用签名验证
    EnableIntegrityCheck  bool   `json:"enable_integrity_check"`     // 启用完整性检查
}
```

#### 预设配置

```go
// 默认配置
config := gm.GetGMConfigDefault()

// 安全配置 (启用签名验证和完整性检查)
config := gm.GetGMConfigSecure()

// 性能配置 (禁用额外验证)
config := gm.GetGMConfigPerformance()

// 自定义配置
config := gm.GetGMConfigCustom("hex", "your-sm4-key")
```

### SM2 椭圆曲线公钥算法

#### 密钥生成

```go
// 生成新的SM2密钥对
keyPair, err := service.GenerateSM2KeyPair()
if err != nil {
    return err
}

fmt.Printf("私钥: %s\n", keyPair.PrivateKey)
fmt.Printf("公钥: %s\n", keyPair.PublicKey)

// 获取当前公钥
publicKey := service.GetSM2PublicKey()
```

#### 加密解密

```go
// SM2 加密
plaintext := "需要加密的数据"
encryptResp, err := service.SM2Encrypt([]byte(plaintext))
if err != nil {
    return err
}

fmt.Printf("密文: %s\n", encryptResp.EncryptedData)
fmt.Printf("算法: %s\n", encryptResp.Algorithm)
fmt.Printf("时间戳: %d\n", encryptResp.Timestamp)

// SM2 解密
decryptResp, err := service.SM2Decrypt(encryptResp.EncryptedData)
if err != nil {
    return err
}

fmt.Printf("明文: %s\n", string(decryptResp.Data))
```

#### 使用指定公钥加密

```go
// 使用其他公钥加密 (需要解析公钥)
// 这里是概念性演示，实际需要实现公钥解析
resp, err := service.SM2Encrypt([]byte(data), customPublicKey)
```

### SM4 分组密码算法

#### 密钥生成

```go
// 生成随机SM4密钥
sm4Key, err := service.GenerateSM4Key()
if err != nil {
    return err
}

fmt.Printf("SM4密钥: %s\n", sm4Key) // 32字符的十六进制字符串

// 解析密钥字符串
keyBytes, err := hex.DecodeString(sm4Key)
if err != nil {
    return err
}
```

#### 加密解密

```go
import "encoding/hex"

// 准备密钥
keyHex := "1234567890abcdef1234567890abcdef" // 32字符十六进制
keyBytes, _ := hex.DecodeString(keyHex)

// SM4 加密
plaintext := "需要加密的数据"
encryptResp, err := service.SM4Encrypt([]byte(plaintext), keyBytes)
if err != nil {
    return err
}

fmt.Printf("密文: %s\n", encryptResp.EncryptedData)

// SM4 解密
decryptResp, err := service.SM4Decrypt(encryptResp.EncryptedData, keyBytes)
if err != nil {
    return err
}

fmt.Printf("明文: %s\n", string(decryptResp.Data))
```

#### 使用默认密钥

```go
// 在配置中设置默认密钥
config := &gm.GMConfig{
    DefaultSM4Key: "1234567890abcdef1234567890abcdef",
    OutputFormat:  "base64",
}
service, _ := gm.NewGMService(config)

// 不提供密钥时使用默认密钥
encryptResp, err := service.SM4Encrypt([]byte("数据"))
decryptResp, err := service.SM4Decrypt(encryptResp.EncryptedData)
```

### SM3 密码杂凑算法

#### 基本哈希

```go
// 计算数据哈希
data := []byte("需要计算哈希的数据")
hashResp, err := service.SM3Hash(data, "hex")
if err != nil {
    return err
}

fmt.Printf("哈希值: %s\n", hashResp.Hash)
fmt.Printf("格式: %s\n", hashResp.Format)

// 字符串哈希
stringHashResp, err := service.SM3HashString("Hello World", "base64")
if err != nil {
    return err
}

fmt.Printf("字符串哈希: %s\n", stringHashResp.Hash)
```

#### 哈希验证

```go
// 验证哈希值
data := []byte("原始数据")
expectedHash := "预期的哈希值"

verifyResp, err := service.SM3Verify(data, expectedHash, "hex")
if err != nil {
    return err
}

if verifyResp.Valid {
    fmt.Println("哈希验证通过")
} else {
    fmt.Println("哈希验证失败")
}
```

### 批量操作

#### 批量加密

```go
// 准备加密请求
requests := []*gm.EncryptRequest{
    {Data: []byte("数据1"), Algorithm: "SM2"},
    {Data: []byte("数据2"), Algorithm: "SM4", Key: sm4Key},
    {Data: []byte("数据3"), Algorithm: "SM2"},
}

// 执行批量加密
responses, err := service.BatchEncrypt(requests)
if err != nil {
    return err
}

for i, resp := range responses {
    fmt.Printf("第%d项加密结果: %s\n", i+1, resp.EncryptedData)
}
```

#### 批量解密

```go
// 准备解密请求
requests := []*gm.DecryptRequest{
    {EncryptedData: "密文1", Algorithm: "SM2"},
    {EncryptedData: "密文2", Algorithm: "SM4", Key: sm4Key},
}

// 执行批量解密
responses, err := service.BatchDecrypt(requests)
if err != nil {
    return err
}

for i, resp := range responses {
    fmt.Printf("第%d项解密结果: %s\n", i+1, string(resp.Data))
}
```

#### 批量哈希

```go
// 准备数据列表
dataList := [][]byte{
    []byte("数据1"),
    []byte("数据2"),
    []byte("数据3"),
}

// 执行批量哈希
responses, err := service.BatchHash(dataList, "hex")
if err != nil {
    return err
}

for i, resp := range responses {
    fmt.Printf("第%d项哈希: %s\n", i+1, resp.Hash)
}
```

### JSON 数据处理

#### JSON 加密

```go
// 准备 JSON 数据
userData := map[string]interface{}{
    "name":  "张三",
    "age":   30,
    "email": "zhang@example.com",
}

// 加密 JSON 数据
encryptResp, err := service.EncryptJSON(userData, "SM2")
if err != nil {
    return err
}

fmt.Printf("加密的JSON: %s\n", encryptResp.EncryptedData)
```

#### JSON 解密

```go
// 解密 JSON 数据
var decryptedData map[string]interface{}
err := service.DecryptJSON(encryptResp.EncryptedData, "SM2", &decryptedData)
if err != nil {
    return err
}

fmt.Printf("解密的JSON: %+v\n", decryptedData)

// 解密到结构体
type User struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

var user User
err = service.DecryptJSON(encryptResp.EncryptedData, "SM2", &user)
```

### 工具函数

#### 快速操作

```go
// 快速 SM2 加密
encrypted, err := gm.QuickSM2EncryptString("Hello World")

// 快速 SM4 加密
key := []byte("1234567890abcdef")
encrypted, err := gm.QuickSM4EncryptString("Hello World", key)

// 快速 SM3 哈希
hash, err := gm.QuickSM3HashString("Hello World", "hex")
```

#### 密钥工具

```go
// 生成随机密钥
randomKey, err := gm.GenerateRandomKeyHex(16)      // 十六进制格式
randomKey, err := gm.GenerateRandomKeyBase64(16)   // Base64格式

// 验证密钥
keyBytes, _ := hex.DecodeString("1234567890abcdef1234567890abcdef")
err := gm.ValidateKeyLength(keyBytes, "SM4")

// 验证密钥字符串
err := gm.ValidateSM4KeyString("1234567890abcdef1234567890abcdef", "hex")

// 解析密钥
keyBytes, err := gm.ParseSM4Key("1234567890abcdef1234567890abcdef", "hex")
```

#### 格式转换

```go
// 编码解码
hexStr := gm.EncodeHex([]byte("data"))
data, err := gm.DecodeHex(hexStr)

base64Str := gm.EncodeBase64([]byte("data"))
data, err := gm.DecodeBase64(base64Str)

// 格式转换
base64Str, err := gm.ConvertFormat(hexStr, "hex", "base64")
hexStr, err := gm.ConvertFormat(base64Str, "base64", "hex")
```

### 性能测试

```go
// 配置性能测试
config := &gm.BenchmarkConfig{
    DataSize:   1024,  // 1KB 数据
    Iterations: 1000,  // 1000 次迭代
    Algorithm:  "SM4", // 测试算法
}

// 运行测试
result, err := gm.RunBenchmark(config)
if err != nil {
    return err
}

fmt.Printf("算法: %s\n", result.Algorithm)
fmt.Printf("数据大小: %d 字节\n", result.DataSize)
fmt.Printf("迭代次数: %d\n", result.Iterations)
fmt.Printf("总耗时: %v\n", result.TotalDuration)
fmt.Printf("平均耗时: %v\n", result.AvgDuration)
fmt.Printf("吞吐量: %.2f MB/s\n", result.ThroughputMBps)
```

## 错误处理

### 常见错误

```go
// 密钥长度错误
if err != nil {
    if strings.Contains(err.Error(), "密钥长度") {
        // 处理密钥长度错误
    }
}

// 解码错误
if err != nil {
    if strings.Contains(err.Error(), "解码") {
        // 处理数据解码错误
    }
}

// 算法不支持
if err != nil {
    if strings.Contains(err.Error(), "不支持") {
        // 处理不支持的算法
    }
}
```

### 最佳实践

1. **密钥管理**
   ```go
   // ✓ 使用安全的密钥生成
   key, err := gm.GenerateRandomKeyHex(16)
   
   // ✗ 不要使用固定的弱密钥
   // key := "1111111111111111"
   ```

2. **错误处理**
   ```go
   // ✓ 始终检查错误
   resp, err := service.SM2Encrypt(data)
   if err != nil {
       log.Printf("加密失败: %v", err)
       return err
   }
   
   // ✗ 忽略错误可能导致安全问题
   // resp, _ := service.SM2Encrypt(data)
   ```

3. **配置选择**
   ```go
   // ✓ 生产环境使用安全配置
   config := gm.GetGMConfigSecure()
   
   // ✓ 开发环境可使用性能配置
   config := gm.GetGMConfigPerformance()
   ```

## 集成示例

### 在 HTTP 处理器中使用

```go
import (
    "gyweb/core/services/gm"
    "github.com/gin-gonic/gin"
)

type CryptoHandler struct {
    gmService *gm.GMService
}

func NewCryptoHandler() *CryptoHandler {
    service, _ := gm.NewGMServiceDefault()
    return &CryptoHandler{gmService: service}
}

func (h *CryptoHandler) EncryptData(c *gin.Context) {
    var req struct {
        Data      string `json:"data"`
        Algorithm string `json:"algorithm"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    var resp interface{}
    var err error
    
    switch req.Algorithm {
    case "SM2":
        resp, err = h.gmService.SM2Encrypt([]byte(req.Data))
    case "SM4":
        key, _ := hex.DecodeString("your-default-key")
        resp, err = h.gmService.SM4Encrypt([]byte(req.Data), key)
    default:
        c.JSON(400, gin.H{"error": "不支持的算法"})
        return
    }
    
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, resp)
}
```

### 在数据库模型中使用

```go
type UserSensitive struct {
    ID       uint   `gorm:"primarykey"`
    Name     string `gorm:"column:name"`
    Phone    string `gorm:"column:phone_encrypted"`
    Email    string `gorm:"column:email_encrypted"`
}

func (u *UserSensitive) BeforeSave(tx *gorm.DB) error {
    service, _ := gm.NewGMServiceDefault()
    
    // 加密手机号
    if u.Phone != "" {
        resp, err := service.SM4Encrypt([]byte(u.Phone), defaultKey)
        if err != nil {
            return err
        }
        u.Phone = resp.EncryptedData
    }
    
    // 加密邮箱
    if u.Email != "" {
        resp, err := service.SM4Encrypt([]byte(u.Email), defaultKey)
        if err != nil {
            return err
        }
        u.Email = resp.EncryptedData
    }
    
    return nil
}

func (u *UserSensitive) AfterFind(tx *gorm.DB) error {
    service, _ := gm.NewGMServiceDefault()
    
    // 解密手机号
    if u.Phone != "" {
        resp, err := service.SM4Decrypt(u.Phone, defaultKey)
        if err != nil {
            return err
        }
        u.Phone = string(resp.Data)
    }
    
    // 解密邮箱
    if u.Email != "" {
        resp, err := service.SM4Decrypt(u.Email, defaultKey)
        if err != nil {
            return err
        }
        u.Email = string(resp.Data)
    }
    
    return nil
}
```

## 注意事项

1. **安全性**
   - 密钥应该通过安全的方式生成和存储
   - 不要在代码中硬编码密钥
   - 定期轮换密钥

2. **性能**
   - SM2 适用于小数据量的加密（如密钥交换）
   - SM4 适用于大数据量的对称加密
   - 批量操作可以提高性能

3. **兼容性**
   - 输出格式（hex/base64）需要在加密和解密时保持一致
   - 不同版本之间的密钥格式可能不兼容

4. **错误处理**
   - 始终检查并适当处理错误
   - 不要泄露敏感的错误信息给客户端

## 完整示例

参考 `examples/gm_service_demo.go` 文件，其中包含了所有功能的完整演示代码。

```bash
# 运行演示
go run examples/gm_service_demo.go
``` 