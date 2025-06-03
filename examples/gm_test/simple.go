package main

import (
	"fmt"

	"gyweb/core/services/smcrypto"
)

func main() {
	fmt.Println("=== 简单的国密兼容性测试 ===")

	// 测试SM3哈希
	testSM2()

}

func testSM2() {
	fmt.Println("\n--- SM3 哈希测试 ---")

	testData := "123456"
	//expectedHash := "207cf410532f92a47dee245ce9b11ff71f578ebd763eb3bbea44ebd043d018fb"

	PUBLIC_KEY := "04298364ec840088475eae92a591e01284d1abefcda348b47eb324bb521bb03b0b2a5bc393f6b71dabb8f15c99a0050818b56b23f31743b93df9cf8948f15ddb54"

	/** 私钥 */
	PRIVATE_KEY := "3037723d47292171677ec8bd7dc9af696c7472bc5f251b2cec07e65fdef22e25"
	sm_server := smcrypto.NewSmCryptoService()
	pass := sm_server.SM2Encrypt(testData, PUBLIC_KEY)

	fmt.Println(pass)
	fmt.Println(sm_server.SM2Decrypt(pass, PRIVATE_KEY))
}
