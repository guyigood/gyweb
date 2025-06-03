package smcrypto

import (
	"testing"
)

func TestSmCryptoService(t *testing.T) {
	service := NewSmCryptoService()

	// 测试SM2密钥生成
	t.Run("SM2KeyPair", func(t *testing.T) {
		keyPair, err := service.GenerateSM2KeyPair()
		if err != nil {
			t.Fatalf("生成SM2密钥对失败: %v", err)
		}
		if keyPair.PrivateKey == "" || keyPair.PublicKey == "" {
			t.Fatal("生成的密钥对为空")
		}
		t.Logf("私钥: %s", keyPair.PrivateKey)
		t.Logf("公钥: %s", keyPair.PublicKey)
	})

	// 测试SM2加密解密
	t.Run("SM2EncryptDecrypt", func(t *testing.T) {
		keyPair, err := service.GenerateSM2KeyPair()
		if err != nil {
			t.Fatalf("生成SM2密钥对失败: %v", err)
		}

		plaintext := "Hello, 国密SM2加密测试!"

		// 加密
		ciphertext, err := service.SM2Encrypt(plaintext, keyPair.PublicKey)
		if err != nil {
			t.Fatalf("SM2加密失败: %v", err)
		}
		t.Logf("加密结果: %s", ciphertext)

		// 解密
		decrypted, err := service.SM2Decrypt(ciphertext, keyPair.PrivateKey)
		if err != nil {
			t.Fatalf("SM2解密失败: %v", err)
		}

		if decrypted != plaintext {
			t.Fatalf("解密结果不匹配，期望: %s，实际: %s", plaintext, decrypted)
		}
		t.Logf("解密成功: %s", decrypted)
	})

	// 测试SM2签名验签
	t.Run("SM2SignVerify", func(t *testing.T) {
		keyPair, err := service.GenerateSM2KeyPair()
		if err != nil {
			t.Fatalf("生成SM2密钥对失败: %v", err)
		}

		message := "Hello, 国密SM2签名测试!"

		// 签名
		signature, err := service.SM2Sign(message, keyPair.PrivateKey)
		if err != nil {
			t.Fatalf("SM2签名失败: %v", err)
		}
		t.Logf("签名结果: %s", signature)

		// 验签
		valid, err := service.SM2Verify(message, signature, keyPair.PublicKey)
		if err != nil {
			t.Fatalf("SM2验签失败: %v", err)
		}

		if !valid {
			t.Fatal("签名验证失败")
		}
		t.Logf("签名验证成功")
	})

	// 测试SM3哈希
	t.Run("SM3Hash", func(t *testing.T) {
		data := "Hello, 国密SM3哈希测试!"
		hash := service.SM3Hash(data)
		if hash == "" {
			t.Fatal("SM3哈希结果为空")
		}
		if len(hash) != 64 { // SM3产生256位（32字节）哈希值，十六进制为64个字符
			t.Fatalf("SM3哈希长度错误，期望64个字符，实际: %d", len(hash))
		}
		t.Logf("SM3哈希: %s", hash)
	})

	// 测试SM4密钥生成
	t.Run("SM4KeyGeneration", func(t *testing.T) {
		key, err := service.GenerateSM4Key()
		if err != nil {
			t.Fatalf("生成SM4密钥失败: %v", err)
		}
		if len(key) != 32 { // 16字节的十六进制表示
			t.Fatalf("SM4密钥长度错误，期望32个字符，实际: %d", len(key))
		}
		t.Logf("SM4密钥: %s", key)

		iv, err := service.GenerateIV()
		if err != nil {
			t.Fatalf("生成IV失败: %v", err)
		}
		if len(iv) != 32 { // 16字节的十六进制表示
			t.Fatalf("IV长度错误，期望32个字符，实际: %d", len(iv))
		}
		t.Logf("IV: %s", iv)
	})

	// 测试SM4 ECB模式加密解密
	t.Run("SM4ECBEncryptDecrypt", func(t *testing.T) {
		key, err := service.GenerateSM4Key()
		if err != nil {
			t.Fatalf("生成SM4密钥失败: %v", err)
		}

		plaintext := "Hello, 国密SM4 ECB加密测试!"
		options := &SM4Options{
			Mode:    "ECB",
			Padding: "PKCS7",
		}

		// 加密
		ciphertext, err := service.SM4Encrypt(plaintext, key, options)
		if err != nil {
			t.Fatalf("SM4 ECB加密失败: %v", err)
		}
		t.Logf("ECB加密结果: %s", ciphertext)

		// 解密
		decrypted, err := service.SM4Decrypt(ciphertext, key, options)
		if err != nil {
			t.Fatalf("SM4 ECB解密失败: %v", err)
		}

		if decrypted != plaintext {
			t.Fatalf("解密结果不匹配，期望: %s，实际: %s", plaintext, decrypted)
		}
		t.Logf("ECB解密成功: %s", decrypted)
	})

	// 测试SM4 CBC模式加密解密
	t.Run("SM4CBCEncryptDecrypt", func(t *testing.T) {
		key, err := service.GenerateSM4Key()
		if err != nil {
			t.Fatalf("生成SM4密钥失败: %v", err)
		}

		iv, err := service.GenerateIV()
		if err != nil {
			t.Fatalf("生成IV失败: %v", err)
		}

		plaintext := "Hello, 国密SM4 CBC加密测试!"
		options := &SM4Options{
			Mode:    "CBC",
			Padding: "PKCS7",
			IV:      iv,
		}

		// 加密
		ciphertext, err := service.SM4Encrypt(plaintext, key, options)
		if err != nil {
			t.Fatalf("SM4 CBC加密失败: %v", err)
		}
		t.Logf("CBC加密结果: %s", ciphertext)

		// 解密
		decrypted, err := service.SM4Decrypt(ciphertext, key, options)
		if err != nil {
			t.Fatalf("SM4 CBC解密失败: %v", err)
		}

		if decrypted != plaintext {
			t.Fatalf("解密结果不匹配，期望: %s，实际: %s", plaintext, decrypted)
		}
		t.Logf("CBC解密成功: %s", decrypted)
	})
}
