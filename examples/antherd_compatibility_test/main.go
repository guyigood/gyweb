package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/guyigood/gyweb/core/services/smcrypto"
)

func main() {
	fmt.Println("=== antherd/sm-crypto 兼容性测试 ===")

	// 创建SmCrypto服务实例
	smService := smcrypto.NewSmCryptoService()

	// 使用已知的密钥对 - 去掉04前缀，使用64字节格式
	privateKey := "59276e27d506861a16680f3ad9c02dccef3cc1fa3cdbe4ce6d54b80deac1bc21"
	publicKey := "09f9df311e5421a150dd7d161e4bc5c672179fad1833fc076bb08ff356f35020ccea490ce26775a52dc6ea718cc1aa600aea1a8c8d2bc8d0c5d0fd16c6b25c6d5e"

	// 测试明文
	plaintext := "123456"

	fmt.Printf("私钥: %s\n", privateKey)
	fmt.Printf("公钥: %s\n", publicKey)
	fmt.Printf("公钥字符串长度: %d (预期字节数: %d)\n", len(publicKey), len(publicKey)/2)
	fmt.Printf("明文: %s\n", plaintext)
	fmt.Println()

	// 测试1: 先使用标准SM2加密
	fmt.Println("=== 测试1: 使用标准SM2加密 ===")
	standardCiphertext, err := smService.SM2Encrypt(plaintext, publicKey)
	if err != nil {
		log.Printf("标准SM2加密失败: %v", err)
	} else {
		fmt.Printf("标准密文: %s\n", standardCiphertext)
		fmt.Printf("标准密文长度: %d字节\n", len(standardCiphertext)/2)

		// 解密验证
		decrypted, err := smService.SM2Decrypt(standardCiphertext, privateKey)
		if err != nil {
			log.Printf("标准解密失败: %v", err)
		} else {
			fmt.Printf("解密结果: %s\n", decrypted)
		}
	}

	// 测试2: 尝试解密已知的Vue密文
	fmt.Println("\n=== 测试2: 分析Vue前端生成的密文 ===")
	vueCiphertext := "1006aaa2ac59c0286403f4d360efe11c139c64cb6717bfb0c37273e39c649ee5ed79ce8cc80ceb57c502a8fffa4ead2fad5b2a4b0e88753a022a5e683c92c2f3bd7f89eb16baab803470ea4e49aa8ac6c8f2c9e3f7a6a7e5e2433bfe8e0583ef31664b45c860"

	fmt.Printf("Vue密文: %s\n", vueCiphertext)
	fmt.Printf("Vue密文长度: %d字节 (%d个十六进制字符)\n", len(vueCiphertext)/2, len(vueCiphertext))

	// 分析密文结构
	analyzeCiphertext(vueCiphertext)

	// 测试3: 尝试使用JavaScript兼容KDF解密Vue密文
	fmt.Println("\n=== 测试3: 使用JavaScript兼容KDF解密Vue密文 ===")
	vueDecrypted, err := smcrypto.SM2DecryptWithJSKDF(vueCiphertext, privateKey)
	if err != nil {
		log.Printf("Vue密文解密失败: %v", err)
	} else {
		fmt.Printf("Vue密文解密结果: '%s'\n", vueDecrypted)
		if vueDecrypted == plaintext {
			fmt.Println("✓ Vue密文解密成功!")
		} else {
			fmt.Printf("✗ Vue密文解密结果不匹配，期望: '%s'\n", plaintext)
		}
	}

	// 测试4: 使用JavaScript兼容KDF加密测试
	fmt.Println("\n=== 测试4: 使用JavaScript兼容KDF加密测试 ===")
	for i := 0; i < 3; i++ {
		jsCiphertext, err := smcrypto.SM2EncryptWithJSKDF(plaintext, publicKey)
		if err != nil {
			log.Printf("第%d次JS兼容加密失败: %v", i+1, err)
			continue
		}

		fmt.Printf("第%d次JS兼容加密: %s (长度: %d字节)\n", i+1, jsCiphertext, len(jsCiphertext)/2)

		// 解密验证
		decrypted, err := smcrypto.SM2DecryptWithJSKDF(jsCiphertext, privateKey)
		if err != nil {
			log.Printf("第%d次JS兼容解密失败: %v", i+1, err)
		} else if decrypted == plaintext {
			fmt.Printf("✓ 第%d次JS兼容验证成功\n", i+1)
		} else {
			fmt.Printf("✗ 第%d次JS兼容验证失败\n", i+1)
		}
	}

	fmt.Println("\n=== 测试完成 ===")
}

// analyzeCiphertext 分析密文结构
func analyzeCiphertext(ciphertextHex string) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		fmt.Printf("解析密文失败: %v\n", err)
		return
	}

	fmt.Printf("密文结构分析:\n")
	fmt.Printf("  总长度: %d字节\n", len(ciphertext))

	if len(ciphertext) >= 1 {
		fmt.Printf("  前缀: 0x%02X", ciphertext[0])
		if ciphertext[0] == 0x10 {
			fmt.Printf(" (antherd格式)")
		} else if ciphertext[0] == 0x04 {
			fmt.Printf(" (标准格式)")
		}
		fmt.Println()
	}

	if len(ciphertext) >= 97 && ciphertext[0] == 0x10 {
		// antherd格式分析
		fmt.Printf("  X坐标: %x (32字节)\n", ciphertext[1:33])
		fmt.Printf("  Y坐标: %x (32字节)\n", ciphertext[33:65])
		fmt.Printf("  C3哈希: %x (32字节)\n", ciphertext[65:97])
		c2Len := len(ciphertext) - 97
		fmt.Printf("  C2数据: %x (%d字节)\n", ciphertext[97:], c2Len)
		fmt.Printf("  C2文本: '%s'\n", string(ciphertext[97:]))
	}
}
