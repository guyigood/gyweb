package gm

import (
	"encoding/hex"
	"testing"
)

func TestGMService_SM2(t *testing.T) {
	service, err := NewGMServiceDefault()
	if err != nil {
		t.Fatalf("创建服务失败: %v", err)
	}

	originalText := "测试SM2加密的数据"

	// 测试加密
	encryptResp, err := service.SM2Encrypt([]byte(originalText))
	if err != nil {
		t.Fatalf("SM2加密失败: %v", err)
	}

	if encryptResp.EncryptedData == "" {
		t.Fatal("加密数据为空")
	}

	if encryptResp.Algorithm != "SM2" {
		t.Fatalf("算法不正确, 期望: SM2, 实际: %s", encryptResp.Algorithm)
	}

	// 测试解密
	decryptResp, err := service.SM2Decrypt(encryptResp.EncryptedData)
	if err != nil {
		t.Fatalf("SM2解密失败: %v", err)
	}

	if string(decryptResp.Data) != originalText {
		t.Fatalf("解密结果不正确, 期望: %s, 实际: %s", originalText, string(decryptResp.Data))
	}
}

func TestGMService_SM4(t *testing.T) {
	service, err := NewGMServiceDefault()
	if err != nil {
		t.Fatalf("创建服务失败: %v", err)
	}

	originalText := "测试SM4加密的数据"
	key := []byte("1234567890abcdef") // 16字节密钥

	// 测试加密
	encryptResp, err := service.SM4Encrypt([]byte(originalText), key)
	if err != nil {
		t.Fatalf("SM4加密失败: %v", err)
	}

	if encryptResp.EncryptedData == "" {
		t.Fatal("加密数据为空")
	}

	if encryptResp.Algorithm != "SM4" {
		t.Fatalf("算法不正确, 期望: SM4, 实际: %s", encryptResp.Algorithm)
	}

	// 测试解密
	decryptResp, err := service.SM4Decrypt(encryptResp.EncryptedData, key)
	if err != nil {
		t.Fatalf("SM4解密失败: %v", err)
	}

	if string(decryptResp.Data) != originalText {
		t.Fatalf("解密结果不正确, 期望: %s, 实际: %s", originalText, string(decryptResp.Data))
	}
}

func TestGMService_SM3(t *testing.T) {
	service, err := NewGMServiceDefault()
	if err != nil {
		t.Fatalf("创建服务失败: %v", err)
	}

	originalText := "测试SM3哈希的数据"

	// 测试哈希计算
	hashResp, err := service.SM3Hash([]byte(originalText), "hex")
	if err != nil {
		t.Fatalf("SM3哈希失败: %v", err)
	}

	if hashResp.Hash == "" {
		t.Fatal("哈希值为空")
	}

	if hashResp.Algorithm != "SM3" {
		t.Fatalf("算法不正确, 期望: SM3, 实际: %s", hashResp.Algorithm)
	}

	if hashResp.Format != "hex" {
		t.Fatalf("格式不正确, 期望: hex, 实际: %s", hashResp.Format)
	}

	// 测试哈希验证
	verifyResp, err := service.SM3Verify([]byte(originalText), hashResp.Hash, "hex")
	if err != nil {
		t.Fatalf("SM3验证失败: %v", err)
	}

	if !verifyResp.Valid {
		t.Fatal("哈希验证失败")
	}
}

func TestGMService_BatchOperations(t *testing.T) {
	service, err := NewGMServiceDefault()
	if err != nil {
		t.Fatalf("创建服务失败: %v", err)
	}

	// 测试批量加密
	texts := []string{"数据1", "数据2", "数据3"}
	requests := make([]*EncryptRequest, len(texts))
	for i, text := range texts {
		requests[i] = &EncryptRequest{
			Data:      []byte(text),
			Algorithm: "SM2",
		}
	}

	encryptResponses, err := service.BatchEncrypt(requests)
	if err != nil {
		t.Fatalf("批量加密失败: %v", err)
	}

	if len(encryptResponses) != len(texts) {
		t.Fatalf("加密响应数量不正确, 期望: %d, 实际: %d", len(texts), len(encryptResponses))
	}

	// 测试批量解密
	decryptRequests := make([]*DecryptRequest, len(encryptResponses))
	for i, resp := range encryptResponses {
		decryptRequests[i] = &DecryptRequest{
			EncryptedData: resp.EncryptedData,
			Algorithm:     "SM2",
		}
	}

	decryptResponses, err := service.BatchDecrypt(decryptRequests)
	if err != nil {
		t.Fatalf("批量解密失败: %v", err)
	}

	if len(decryptResponses) != len(texts) {
		t.Fatalf("解密响应数量不正确, 期望: %d, 实际: %d", len(texts), len(decryptResponses))
	}

	// 验证结果
	for i, resp := range decryptResponses {
		if string(resp.Data) != texts[i] {
			t.Fatalf("批量解密结果不正确, 索引: %d, 期望: %s, 实际: %s", i, texts[i], string(resp.Data))
		}
	}
}

func TestGMService_JSON(t *testing.T) {
	service, err := NewGMServiceDefault()
	if err != nil {
		t.Fatalf("创建服务失败: %v", err)
	}

	// 测试JSON加密
	userData := map[string]interface{}{
		"name":  "张三",
		"age":   30,
		"email": "test@example.com",
	}

	encryptResp, err := service.EncryptJSON(userData, "SM2")
	if err != nil {
		t.Fatalf("JSON加密失败: %v", err)
	}

	if encryptResp.EncryptedData == "" {
		t.Fatal("加密的JSON数据为空")
	}

	// 测试JSON解密
	var decryptedData map[string]interface{}
	err = service.DecryptJSON(encryptResp.EncryptedData, "SM2", &decryptedData)
	if err != nil {
		t.Fatalf("JSON解密失败: %v", err)
	}

	if decryptedData["name"] != userData["name"] {
		t.Fatalf("解密后的名称不正确, 期望: %v, 实际: %v", userData["name"], decryptedData["name"])
	}

	if decryptedData["email"] != userData["email"] {
		t.Fatalf("解密后的邮箱不正确, 期望: %v, 实际: %v", userData["email"], decryptedData["email"])
	}
}

func TestQuickFunctions(t *testing.T) {
	// 测试快速SM2加密
	originalText := "Hello, World!"
	encrypted, err := QuickSM2EncryptString(originalText)
	if err != nil {
		t.Fatalf("快速SM2加密失败: %v", err)
	}

	if encrypted == "" {
		t.Fatal("快速加密结果为空")
	}

	// 测试快速SM3哈希
	hash, err := QuickSM3HashString(originalText, "hex")
	if err != nil {
		t.Fatalf("快速SM3哈希失败: %v", err)
	}

	if hash == "" {
		t.Fatal("快速哈希结果为空")
	}

	// 验证哈希长度 (SM3产生256位/32字节/64字符的十六进制)
	if len(hash) != 64 {
		t.Fatalf("哈希长度不正确, 期望: 64, 实际: %d", len(hash))
	}
}

func TestGMUtils(t *testing.T) {
	// 测试密钥生成
	randomKey, err := GenerateRandomKeyHex(16)
	if err != nil {
		t.Fatalf("生成随机密钥失败: %v", err)
	}

	if len(randomKey) != 32 { // 16字节 = 32个十六进制字符
		t.Fatalf("随机密钥长度不正确, 期望: 32, 实际: %d", len(randomKey))
	}

	// 测试密钥验证
	keyBytes, err := hex.DecodeString(randomKey)
	if err != nil {
		t.Fatalf("解析密钥失败: %v", err)
	}

	err = ValidateKeyLength(keyBytes, "SM4")
	if err != nil {
		t.Fatalf("密钥验证失败: %v", err)
	}

	// 测试格式转换
	base64Key, err := ConvertFormat(randomKey, "hex", "base64")
	if err != nil {
		t.Fatalf("格式转换失败: %v", err)
	}

	// 转换回来验证
	hexKey, err := ConvertFormat(base64Key, "base64", "hex")
	if err != nil {
		t.Fatalf("格式转换回hex失败: %v", err)
	}

	if hexKey != randomKey {
		t.Fatalf("格式转换不一致, 原始: %s, 转换后: %s", randomKey, hexKey)
	}
}

func TestGMConfigs(t *testing.T) {
	// 测试默认配置
	defaultConfig := GetGMConfigDefault()
	if defaultConfig.OutputFormat != "base64" {
		t.Fatalf("默认配置输出格式不正确, 期望: base64, 实际: %s", defaultConfig.OutputFormat)
	}

	// 测试安全配置
	secureConfig := GetGMConfigSecure()
	if secureConfig.OutputFormat != "hex" {
		t.Fatalf("安全配置输出格式不正确, 期望: hex, 实际: %s", secureConfig.OutputFormat)
	}

	if !secureConfig.EnableSignatureVerify {
		t.Fatal("安全配置应该启用签名验证")
	}

	// 测试自定义配置
	customConfig := GetGMConfigCustom("hex", "testkey")
	if customConfig.OutputFormat != "hex" {
		t.Fatalf("自定义配置输出格式不正确, 期望: hex, 实际: %s", customConfig.OutputFormat)
	}

	if customConfig.DefaultSM4Key != "testkey" {
		t.Fatalf("自定义配置SM4密钥不正确, 期望: testkey, 实际: %s", customConfig.DefaultSM4Key)
	}
}

func TestSM4KeyValidation(t *testing.T) {
	// 测试有效密钥
	validKey := []byte("1234567890abcdef") // 16字节
	err := ValidateSM4Key(validKey)
	if err != nil {
		t.Fatalf("有效密钥验证失败: %v", err)
	}

	// 测试无效密钥长度
	invalidKey := []byte("short") // 少于16字节
	err = ValidateSM4Key(invalidKey)
	if err == nil {
		t.Fatal("无效密钥应该验证失败")
	}

	// 测试密钥字符串验证
	validKeyHex := "1234567890abcdef1234567890abcdef"
	err = ValidateSM4KeyString(validKeyHex, "hex")
	if err != nil {
		t.Fatalf("有效密钥字符串验证失败: %v", err)
	}

	// 测试无效密钥字符串
	invalidKeyHex := "invalid"
	err = ValidateSM4KeyString(invalidKeyHex, "hex")
	if err == nil {
		t.Fatal("无效密钥字符串应该验证失败")
	}
}

// 基准测试
func BenchmarkSM2Encrypt(b *testing.B) {
	service, _ := NewGMServiceDefault()
	data := []byte("benchmark test data for SM2 encryption")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.SM2Encrypt(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSM4Encrypt(b *testing.B) {
	service, _ := NewGMServiceDefault()
	data := []byte("benchmark test data for SM4 encryption")
	key := []byte("1234567890abcdef")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.SM4Encrypt(data, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSM3Hash(b *testing.B) {
	service, _ := NewGMServiceDefault()
	data := []byte("benchmark test data for SM3 hash")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.SM3Hash(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
