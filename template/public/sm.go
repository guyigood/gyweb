package public

import "github.com/guyigood/gyweb/core/services/smcrypto"

func Sm2Encrpt(plainText string) (string, error) {
	sm2 := smcrypto.NewSmCryptoService()
	return sm2.SM2Encrypt(plainText, PublicKey, 1)
}

func Sm2Decrypt(cipherText string) (string, error) {
	sm2 := smcrypto.NewSmCryptoService()
	return sm2.SM2Decrypt(cipherText, PrivateKey, 1)
}

func Sm3Hash(plainText string) string {
	sm3 := smcrypto.NewSmCryptoService()
	return sm3.GetSM3HashString(plainText)
}
