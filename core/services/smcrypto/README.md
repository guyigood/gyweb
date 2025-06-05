SM Crypto Go 实现说明文档 
 
概述 
 
本包提供了国密算法(SM2, SM3)的Go语言实现，完全兼容 JavaScript库的功能。
 
功能特性 
 
- SM2加密/解密 - 与sm-crypto库行为完全一致 
  - 支持两种密文模式：
    - `C1C2C3` (默认模式)
    - `C1C3C2` 
- SM3哈希计算 
- 纯Go实现，除`github.com/tjfoc/gmsm/sm3`外无其他依赖 
 
安装使用 
 
```go 
go get github.com/your-repo/smcrypto 
```
 
核心API 
 
SM2加密 
 
```go 
func (c *SmCryptoService) SM2Encrypt(plaintext, publicKeyHex string, cipherMode int) (string, error)
```
 
参数说明:
- `plaintext`: 待加密的明文字符串 
- `publicKeyHex`: 十六进制格式的公钥 
- `cipherMode`: 加密模式，0表示C1C2C3，1表示C1C3C2 
 
返回值:
- 加密后的十六进制字符串 
- 错误信息(如有)
 
SM2解密 
 
```go 
func (c *SmCryptoService) SM2Decrypt(encryptDataHex, privateKeyHex string, cipherMode int) (string, error)
```
 
参数说明:
- `encryptDataHex`: 十六进制格式的密文 
- `privateKeyHex`: 十六进制格式的私钥 
- `cipherMode`: 解密模式，需与加密时一致 
 
返回值:
- 解密后的明文字符串 
- 错误信息(如有)
 
SM3哈希 
 
```go 
func (c *SmCryptoService) GetSM3HashString(decrypted string) string 
```
 
参数说明:
- `decrypted`: 待计算哈希的字符串 
 
返回值:
- 计算得到的SM3哈希值(十六进制字符串)
 
使用示例 
 
初始化服务 
 
```go 
smCrypto := NewSmCryptoService()
```
 
加密示例 
 
```go 
// 公钥 
publicKey := "04298364ec840088475eae92a591e01284d1abefcda348b47eb324bb521bb03b0b2a5bc393f6b71dabb8f15c99a0050818b56b23f31743b93df9cf8948f15ddb54"
 
// 加密"123456"，使用C1C3C2模式 
ciphertext, err := smCrypto.SM2Encrypt("123456", publicKey, 1)
if err != nil {
    fmt.Println("加密失败:", err)
    return 
}
fmt.Println("加密结果:", ciphertext)
```
 
解密示例 
 
```go 
// 私钥 
privateKey := "3037723d47292171677ec8bd7dc9af696c7472bc5f251b2cec07e65fdef22e25"
 
// 解密，使用C1C3C2模式 
plaintext, err := smCrypto.SM2Decrypt(ciphertext, privateKey, 1)
if err != nil {
    fmt.Println("解密失败:", err)
    return 
}
fmt.Println("解密结果:", plaintext)
```
 
SM3哈希示例 
 
```go 
hash := smCrypto.GetSM3HashString("123456")
fmt.Println("SM3哈希:", hash)
```
 
注意事项 
 
1. 密钥格式：公钥和私钥都必须是十六进制字符串格式 
2. 模式匹配：解密时必须使用与加密时相同的模式(C1C2C3或C1C3C2)
3. 性能考虑：本实现包含手动实现的椭圆曲线运算，生产环境建议使用硬件加速 
 
兼容性说明 
 
本实现已通过以下测试用例验证：
- 加密"123456"生成的密文与sm-crypto结果完全一致 
- 解密sm-crypto生成的密文能正确还原原文 
- SM3哈希值与sm-crypto计算结果一致 
 
贡献指南 
 
欢迎提交Issue和Pull Request。提交前请确保：
1. 代码通过`go test`测试 
2. 新增功能包含测试用例 
3. 代码风格符合Go语言惯例 
 
许可证 
 
Apache License 2.0