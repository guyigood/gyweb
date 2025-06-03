package smcrypto

import (
	"fmt"
	"log"
)

// Example 展示国密算法的使用示例
func Example() {
	service := NewSmCryptoService()

	fmt.Println("=== 国密算法使用示例 ===")

	// SM2 示例
	fmt.Println("\n1. SM2 椭圆曲线加密算法示例:")

	// 生成密钥对
	keyPair, err := service.GenerateSM2KeyPair()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("私钥: %s\n", keyPair.PrivateKey)
	fmt.Printf("公钥: %s\n", keyPair.PublicKey)

	// 加密解密
	plaintext := "Hello, 国密SM2!"
	ciphertext, err := service.SM2Encrypt(plaintext, keyPair.PublicKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("加密: %s -> %s\n", plaintext, ciphertext)

	decrypted, err := service.SM2Decrypt(ciphertext, keyPair.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("解密: %s -> %s\n", ciphertext, decrypted)

	// 签名验签
	message := "Hello, 国密SM2签名!"
	signature, err := service.SM2Sign(message, keyPair.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("签名: %s -> %s\n", message, signature)

	valid, err := service.SM2Verify(message, signature, keyPair.PublicKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("验签: %v\n", valid)

	// SM3 示例
	fmt.Println("\n2. SM3 密码杂凑算法示例:")
	data := "Hello, 国密SM3!"
	hash := service.SM3Hash(data)
	fmt.Printf("哈希: %s -> %s\n", data, hash)

	// SM4 示例
	fmt.Println("\n3. SM4 对称加密算法示例:")

	// ECB模式
	fmt.Println("ECB模式:")
	sm4Key, err := service.GenerateSM4Key()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("密钥: %s\n", sm4Key)

	sm4Plaintext := "Hello, 国密SM4!"
	sm4Options := &SM4Options{
		Mode:    "ECB",
		Padding: "PKCS7",
	}

	sm4Ciphertext, err := service.SM4Encrypt(sm4Plaintext, sm4Key, sm4Options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("加密: %s -> %s\n", sm4Plaintext, sm4Ciphertext)

	sm4Decrypted, err := service.SM4Decrypt(sm4Ciphertext, sm4Key, sm4Options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("解密: %s -> %s\n", sm4Ciphertext, sm4Decrypted)

	// CBC模式
	fmt.Println("CBC模式:")
	iv, err := service.GenerateIV()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("IV: %s\n", iv)

	sm4CBCOptions := &SM4Options{
		Mode:    "CBC",
		Padding: "PKCS7",
		IV:      iv,
	}

	sm4CBCCiphertext, err := service.SM4Encrypt(sm4Plaintext, sm4Key, sm4CBCOptions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("加密: %s -> %s\n", sm4Plaintext, sm4CBCCiphertext)

	sm4CBCDecrypted, err := service.SM4Decrypt(sm4CBCCiphertext, sm4Key, sm4CBCOptions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("解密: %s -> %s\n", sm4CBCCiphertext, sm4CBCDecrypted)

	fmt.Println("\n=== 示例完成 ===")
}
