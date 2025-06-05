package main

import (
	"encoding/hex"
	"fmt"

	"github.com/guyigood/gyweb/core/services/smcrypto"
)

func main() {
	fmt.Println("=== antherd/sm-crypto 专项兼容性测试 ===")

	testData := "123456"
	PUBLIC_KEY := "04298364ec840088475eae92a591e01284d1abefcda348b47eb324bb521bb03b0b2a5bc393f6b71dabb8f15c99a0050818b56b23f31743b93df9cf8948f15ddb54"
	PRIVATE_KEY := "3037723d47292171677ec8bd7dc9af696c7472bc5f251b2cec07e65fdef22e25"

	// Vue/Java antherd/sm-crypto生成的密文
	antherdEncrypted := "1006aaa2ac59c0286403f4d360efe11c139c64cb6717bfb0c37273e39c649ee5ed79ce8cc80ceb57c502a8fffa4ead2fad5b2a4b0e88753a022a5e683c92c2f3bd7f89eb16baab803470ea4e49aa8ac6c8f2c9e3f7a6a7e5e2433bfe8e0583ef31664b45c860"

	sm_server := smcrypto.NewSmCryptoService()

	fmt.Println("\n--- 1. 密文格式验证 ---")
	fmt.Printf("密文长度: %d字节\n", len(antherdEncrypted)/2)

	cipherBytes, _ := hex.DecodeString(antherdEncrypted)
	fmt.Printf("首字节: 0x%02X\n", cipherBytes[0])

	if sm_server.IsAntherdFormat(antherdEncrypted) {
		fmt.Println("✅ 确认为antherd格式")
	} else {
		fmt.Println("❌ 不是antherd格式")
	}

	fmt.Println("\n--- 2. 解密测试 ---")

	// 手动解析antherd密文结构
	fmt.Println("手动解析antherd密文结构:")
	fmt.Printf("完整密文: %s\n", antherdEncrypted)
	fmt.Printf("第1字节(标识): 0x%02X\n", cipherBytes[0])

	xBytes := cipherBytes[1:33]
	yBytes := cipherBytes[33:65]
	c3Bytes := cipherBytes[65:97]
	c2Bytes := cipherBytes[97:]

	fmt.Printf("X坐标(1-32): %s\n", hex.EncodeToString(xBytes))
	fmt.Printf("Y坐标(33-64): %s\n", hex.EncodeToString(yBytes))
	fmt.Printf("C3(65-96): %s\n", hex.EncodeToString(c3Bytes))
	fmt.Printf("C2(97-102): %s (长度: %d字节)\n", hex.EncodeToString(c2Bytes), len(c2Bytes))

	// 尝试不同的C2解码
	fmt.Printf("C2字节序列: ")
	for i, b := range c2Bytes {
		fmt.Printf("0x%02X ", b)
		if i < len(c2Bytes)-1 {
			fmt.Printf(", ")
		}
	}
	fmt.Println()

	// 检查是否是"123456"的某种编码
	testBytes := []byte("123456")
	fmt.Printf("期望的明文'123456'字节: %s\n", hex.EncodeToString(testBytes))

	// 这可能是加密后的数据，不是明文

	// 尝试使用新的antherd解密函数
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("❌ antherd解密panic: %v\n", r)
			}
		}()

		decrypted, err := sm_server.SM2DecryptAntherd(antherdEncrypted, PRIVATE_KEY)
		if err != nil {
			fmt.Printf("❌ antherd解密失败: %v\n", err)
		} else {
			fmt.Printf("✅ antherd解密成功: %s\n", decrypted)
		}
	}()

	fmt.Println("\n--- 3. 生成兼容密文测试 ---")

	// 生成antherd兼容格式密文
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("❌ antherd加密panic: %v\n", r)
			}
		}()

		goAntherdEncrypted, err := sm_server.SM2EncryptAntherd(testData, PUBLIC_KEY)
		if err != nil {
			fmt.Printf("❌ antherd加密失败: %v\n", err)
			return
		}

		fmt.Printf("✅ Go生成antherd格式密文成功\n")
		fmt.Printf("密文长度: %d字节\n", len(goAntherdEncrypted)/2)

		goBytes, _ := hex.DecodeString(goAntherdEncrypted)
		fmt.Printf("首字节: 0x%02X\n", goBytes[0])

		// 尝试解密自己生成的密文
		decrypted, err := sm_server.SM2DecryptAntherd(goAntherdEncrypted, PRIVATE_KEY)
		if err != nil {
			fmt.Printf("❌ 解密自生成密文失败: %v\n", err)
		} else {
			fmt.Printf("✅ 解密自生成密文成功: %s\n", decrypted)
		}
	}()

	fmt.Println("\n--- 4. 手动转换测试 ---")

	// 手动构造标准格式
	standardBytes := make([]byte, 0, 1+32+32+32+5)
	standardBytes = append(standardBytes, 0x04) // 标准前缀
	standardBytes = append(standardBytes, xBytes...)
	standardBytes = append(standardBytes, yBytes...)
	standardBytes = append(standardBytes, c3Bytes...)
	standardBytes = append(standardBytes, c2Bytes...)

	fmt.Printf("手动构造的标准格式密文长度: %d字节\n", len(standardBytes))
	fmt.Printf("标准格式密文: %s\n", hex.EncodeToString(standardBytes)[:64]+"...")

	// 尝试用标准方法解密
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("❌ 标准解密panic: %v\n", r)
			}
		}()

		decrypted, err := sm_server.SM2DecryptC1C3C2(hex.EncodeToString(standardBytes), PRIVATE_KEY)
		if err != nil {
			fmt.Printf("❌ 标准C1C3C2解密失败: %v\n", err)
		} else {
			fmt.Printf("✅ 标准C1C3C2解密成功: %s\n", decrypted)
		}
	}()

	fmt.Println("\n--- 测试完成 ---")
}
