# 国密加密服务模块

本模块提供了中国国密算法的Go语言实现，包括SM2、SM3、SM4三种主要算法。

## 功能特性

### SM2 椭圆曲线加密算法
- 密钥对生成
- 公钥加密/私钥解密
- 私钥签名/公钥验签

### SM3 密码杂凑算法
- 字符串哈希
- 字节数组哈希

### SM4 对称加密算法
- ECB/CBC加密模式
- PKCS7填充
- 密钥和IV生成

## 使用方法

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/guyigood/gyweb/core/services/smcrypto"
)

func main() {
    // 创建服务实例
    service := smcrypto.NewSmCryptoService()
    
    // SM2示例
    keyPair, _ := service.GenerateSM2KeyPair()
    ciphertext, _ := service.SM2Encrypt("Hello World", keyPair.PublicKey)
    plaintext, _ := service.SM2Decrypt(ciphertext, keyPair.PrivateKey)
    
    // SM3示例
    hash := service.SM3Hash("Hello World")
    
    // SM4示例
    key, _ := service.GenerateSM4Key()
    options := &smcrypto.SM4Options{
        Mode:    "ECB",
        Padding: "PKCS7",
    }
    encrypted, _ := service.SM4Encrypt("Hello World", key, options)
    decrypted, _ := service.SM4Decrypt(encrypted, key, options)
}
```

### API 文档

#### SM2 相关方法

##### GenerateSM2KeyPair() (*SM2KeyPair, error)
生成SM2密钥对

**返回值:**
- `SM2KeyPair`: 包含PrivateKey和PublicKey的结构体
- `error`: 错误信息

##### SM2Encrypt(plaintext, publicKeyHex string) (string, error)
SM2公钥加密

**参数:**
- `plaintext`: 明文字符串
- `publicKeyHex`: 公钥（十六进制字符串）

**返回值:**
- `string`: 密文（十六进制字符串）
- `error`: 错误信息

##### SM2Decrypt(ciphertextHex, privateKeyHex string) (string, error)
SM2私钥解密

**参数:**
- `ciphertextHex`: 密文（十六进制字符串）
- `privateKeyHex`: 私钥（十六进制字符串）

**返回值:**
- `string`: 明文字符串
- `error`: 错误信息

##### SM2Sign(message, privateKeyHex string) (string, error)
SM2数字签名

**参数:**
- `message`: 待签名消息
- `privateKeyHex`: 私钥（十六进制字符串）

**返回值:**
- `string`: 签名（十六进制字符串）
- `error`: 错误信息

##### SM2Verify(message, signatureHex, publicKeyHex string) (bool, error)
SM2签名验证

**参数:**
- `message`: 原始消息
- `signatureHex`: 签名（十六进制字符串）
- `publicKeyHex`: 公钥（十六进制字符串）

**返回值:**
- `bool`: 验证结果
- `error`: 错误信息

#### SM3 相关方法

##### SM3Hash(data string) string
计算字符串的SM3哈希值

**参数:**
- `data`: 输入字符串

**返回值:**
- `string`: 哈希值（十六进制字符串）

##### SM3HashBytes(data []byte) string
计算字节数组的SM3哈希值

**参数:**
- `data`: 输入字节数组

**返回值:**
- `string`: 哈希值（十六进制字符串）

#### SM4 相关方法

##### SM4Encrypt(plaintext, keyHex string, options *SM4Options) (string, error)
SM4对称加密

**参数:**
- `plaintext`: 明文字符串
- `keyHex`: 密钥（十六进制字符串，32个字符）
- `options`: 加密选项

**返回值:**
- `string`: 密文（十六进制字符串）
- `error`: 错误信息

##### SM4Decrypt(ciphertextHex, keyHex string, options *SM4Options) (string, error)
SM4对称解密

**参数:**
- `ciphertextHex`: 密文（十六进制字符串）
- `keyHex`: 密钥（十六进制字符串，32个字符）
- `options`: 解密选项

**返回值:**
- `string`: 明文字符串
- `error`: 错误信息

##### GenerateSM4Key() (string, error)
生成SM4密钥

**返回值:**
- `string`: 密钥（十六进制字符串，32个字符）
- `error`: 错误信息

##### GenerateIV() (string, error)
生成初始向量

**返回值:**
- `string`: IV（十六进制字符串，32个字符）
- `error`: 错误信息

### 数据结构

#### SM2KeyPair
```go
type SM2KeyPair struct {
    PrivateKey string `json:"private_key"` // 私钥（十六进制字符串）
    PublicKey  string `json:"public_key"`  // 公钥（十六进制字符串）
}
```

#### SM4Options
```go
type SM4Options struct {
    Mode    string `json:"mode"`    // 加密模式：ECB、CBC
    Padding string `json:"padding"` // 填充模式：PKCS7、ZERO
    IV      string `json:"iv"`      // 初始向量（CBC模式需要）
}
```

## 运行测试

```bash
go test -v ./core/services/smcrypto/
```

## 注意事项

1. 所有密钥和加密结果都使用十六进制字符串表示
2. SM4密钥长度固定为16字节（32个十六进制字符）
3. CBC模式需要提供16字节的初始向量IV
4. 本模块基于 `github.com/ZZMarquis/gm` 库实现
5. 符合中国国密标准：GM/T 0003-2012 (SM2)、GM/T 0004-2012 (SM3)、GM/T 0002-2012 (SM4)

## 示例程序

运行示例：
```go
package main

import "github.com/guyigood/gyweb/core/services/smcrypto"

func main() {
    smcrypto.Example()
}
``` 