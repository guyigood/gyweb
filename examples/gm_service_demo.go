package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/guyigood/gyweb/core/services/gm"
)

func main() {
	fmt.Println("=== 国密服务演示 ===")

	// 1. 创建国密服务
	service, err := gm.NewGMServiceDefault()
	if err != nil {
		log.Fatal("创建国密服务失败:", err)
	}

	// 2. SM2 演示
	fmt.Println("\n--- SM2 加解密演示 ---")
	sm2Demo(service)

	// 3. SM4 演示
	fmt.Println("\n--- SM4 加解密演示 ---")
	sm4Demo(service)

	// 4. SM3 演示
	fmt.Println("\n--- SM3 哈希演示 ---")
	sm3Demo(service)

	// 5. 批量操作演示
	fmt.Println("\n--- 批量操作演示 ---")
	batchDemo(service)

	// 6. 工具函数演示
	fmt.Println("\n--- 工具函数演示 ---")
	utilsDemo()

	// 7. JSON 加解密演示
	fmt.Println("\n--- JSON 加解密演示 ---")
	jsonDemo(service)
}

// SM2 加解密演示
func sm2Demo(service *gm.GMService) {
	originalText := "这是一段需要用SM2加密的敏感信息"
	fmt.Printf("原文: %s\n", originalText)

	// 加密
	encryptResp, err := service.SM2Encrypt([]byte(originalText))
	if err != nil {
		log.Printf("SM2加密失败: %v", err)
		return
	}
	fmt.Printf("密文: %s\n", encryptResp.EncryptedData)

	// 解密
	decryptResp, err := service.SM2Decrypt(encryptResp.EncryptedData)
	if err != nil {
		log.Printf("SM2解密失败: %v", err)
		return
	}
	fmt.Printf("解密后: %s\n", string(decryptResp.Data))

	// 验证
	if string(decryptResp.Data) == originalText {
		fmt.Println("✓ SM2 加解密验证成功")
	} else {
		fmt.Println("✗ SM2 加解密验证失败")
	}
}

// SM4 加解密演示
func sm4Demo(service *gm.GMService) {
	originalText := "这是一段需要用SM4加密的数据"
	fmt.Printf("原文: %s\n", originalText)

	// 生成SM4密钥
	sm4Key, err := service.GenerateSM4Key()
	if err != nil {
		log.Printf("生成SM4密钥失败: %v", err)
		return
	}
	fmt.Printf("SM4密钥: %s\n", sm4Key)

	// 解析密钥
	keyBytes, err := hex.DecodeString(sm4Key)
	if err != nil {
		log.Printf("解析密钥失败: %v", err)
		return
	}

	// 加密
	encryptResp, err := service.SM4Encrypt([]byte(originalText), keyBytes)
	if err != nil {
		log.Printf("SM4加密失败: %v", err)
		return
	}
	fmt.Printf("密文: %s\n", encryptResp.EncryptedData)

	// 解密
	decryptResp, err := service.SM4Decrypt(encryptResp.EncryptedData, keyBytes)
	if err != nil {
		log.Printf("SM4解密失败: %v", err)
		return
	}
	fmt.Printf("解密后: %s\n", string(decryptResp.Data))

	// 验证
	if string(decryptResp.Data) == originalText {
		fmt.Println("✓ SM4 加解密验证成功")
	} else {
		fmt.Println("✗ SM4 加解密验证失败")
	}
}

// SM3 哈希演示
func sm3Demo(service *gm.GMService) {
	originalText := "这是一段需要计算SM3哈希的数据"
	fmt.Printf("原文: %s\n", originalText)

	// 计算哈希
	hashResp, err := service.SM3Hash([]byte(originalText), "hex")
	if err != nil {
		log.Printf("SM3哈希计算失败: %v", err)
		return
	}
	fmt.Printf("SM3哈希(hex): %s\n", hashResp.Hash)

	// 验证哈希
	verifyResp, err := service.SM3Verify([]byte(originalText), hashResp.Hash, "hex")
	if err != nil {
		log.Printf("SM3哈希验证失败: %v", err)
		return
	}

	if verifyResp.Valid {
		fmt.Println("✓ SM3 哈希验证成功")
	} else {
		fmt.Println("✗ SM3 哈希验证失败")
	}

	// 字符串哈希
	stringHashResp, err := service.SM3HashString("Hello, 国密!", "base64")
	if err != nil {
		log.Printf("字符串哈希失败: %v", err)
		return
	}
	fmt.Printf("字符串哈希(base64): %s\n", stringHashResp.Hash)
}

// 批量操作演示
func batchDemo(service *gm.GMService) {
	// 准备数据
	texts := []string{
		"第一段数据",
		"第二段数据",
		"第三段数据",
	}

	// 批量加密请求
	encryptRequests := make([]*gm.EncryptRequest, len(texts))
	for i, text := range texts {
		encryptRequests[i] = &gm.EncryptRequest{
			Data:      []byte(text),
			Algorithm: "SM2",
		}
	}

	// 执行批量加密
	encryptResponses, err := service.BatchEncrypt(encryptRequests)
	if err != nil {
		log.Printf("批量加密失败: %v", err)
		return
	}
	fmt.Printf("批量加密成功，共处理 %d 项\n", len(encryptResponses))

	// 批量解密请求
	decryptRequests := make([]*gm.DecryptRequest, len(encryptResponses))
	for i, resp := range encryptResponses {
		decryptRequests[i] = &gm.DecryptRequest{
			EncryptedData: resp.EncryptedData,
			Algorithm:     "SM2",
		}
	}

	// 执行批量解密
	decryptResponses, err := service.BatchDecrypt(decryptRequests)
	if err != nil {
		log.Printf("批量解密失败: %v", err)
		return
	}

	// 验证结果
	allValid := true
	for i, resp := range decryptResponses {
		if string(resp.Data) != texts[i] {
			allValid = false
			break
		}
	}

	if allValid {
		fmt.Println("✓ 批量加解密验证成功")
	} else {
		fmt.Println("✗ 批量加解密验证失败")
	}

	// 批量哈希
	dataList := make([][]byte, len(texts))
	for i, text := range texts {
		dataList[i] = []byte(text)
	}

	hashResponses, err := service.BatchHash(dataList, "hex")
	if err != nil {
		log.Printf("批量哈希失败: %v", err)
		return
	}

	fmt.Printf("批量哈希成功，共处理 %d 项:\n", len(hashResponses))
	for i, hashResp := range hashResponses {
		fmt.Printf("  %d: %s -> %s\n", i+1, texts[i], hashResp.Hash)
	}
}

// 工具函数演示
func utilsDemo() {
	// 快速加密
	encrypted, err := gm.QuickSM2EncryptString("Hello, World!")
	if err != nil {
		log.Printf("快速加密失败: %v", err)
		return
	}
	fmt.Printf("快速SM2加密结果: %s\n", encrypted)

	// 快速哈希
	hash, err := gm.QuickSM3HashString("Hello, 国密!", "hex")
	if err != nil {
		log.Printf("快速哈希失败: %v", err)
		return
	}
	fmt.Printf("快速SM3哈希结果: %s\n", hash)

	// 生成随机密钥
	randomKey, err := gm.GenerateRandomKeyHex(16)
	if err != nil {
		log.Printf("生成随机密钥失败: %v", err)
		return
	}
	fmt.Printf("随机SM4密钥: %s\n", randomKey)

	// 格式转换
	base64Key, err := gm.ConvertFormat(randomKey, "hex", "base64")
	if err != nil {
		log.Printf("格式转换失败: %v", err)
		return
	}
	fmt.Printf("Base64格式密钥: %s\n", base64Key)

	// 验证密钥
	keyBytes, _ := hex.DecodeString(randomKey)
	err = gm.ValidateKeyLength(keyBytes, "SM4")
	if err != nil {
		fmt.Printf("密钥验证失败: %v\n", err)
	} else {
		fmt.Println("✓ 密钥验证成功")
	}
}

// JSON 加解密演示
func jsonDemo(service *gm.GMService) {
	// 准备JSON数据
	userData := map[string]interface{}{
		"name":  "张三",
		"age":   30,
		"email": "zhangsan@example.com",
		"phone": "13800138000",
	}
	fmt.Printf("原始JSON数据: %+v\n", userData)

	// 加密JSON
	encryptResp, err := service.EncryptJSON(userData, "SM2")
	if err != nil {
		log.Printf("JSON加密失败: %v", err)
		return
	}
	fmt.Printf("加密后的JSON: %s\n", encryptResp.EncryptedData)

	// 解密JSON
	var decryptedData map[string]interface{}
	err = service.DecryptJSON(encryptResp.EncryptedData, "SM2", &decryptedData)
	if err != nil {
		log.Printf("JSON解密失败: %v", err)
		return
	}
	fmt.Printf("解密后的JSON: %+v\n", decryptedData)

	// 验证
	if decryptedData["name"] == userData["name"] &&
		decryptedData["email"] == userData["email"] {
		fmt.Println("✓ JSON 加解密验证成功")
	} else {
		fmt.Println("✗ JSON 加解密验证失败")
	}
}

// 配置演示
func configDemo() {
	// 默认配置
	defaultConfig := gm.GetGMConfigDefault()
	fmt.Printf("默认配置: %+v\n", defaultConfig)

	// 安全配置
	secureConfig := gm.GetGMConfigSecure()
	fmt.Printf("安全配置: %+v\n", secureConfig)

	// 性能配置
	perfConfig := gm.GetGMConfigPerformance()
	fmt.Printf("性能配置: %+v\n", perfConfig)

	// 自定义配置
	customConfig := gm.GetGMConfigCustom("hex", "abcdef1234567890abcdef1234567890")
	fmt.Printf("自定义配置: %+v\n", customConfig)
}
