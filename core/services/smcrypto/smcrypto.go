package smcrypto

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ZZMarquis/gm/sm2"
	"github.com/ZZMarquis/gm/sm3"
	"github.com/ZZMarquis/gm/sm4"
)

// SmCryptoService 国密服务
type SmCryptoService struct{}

// SM2KeyPair SM2密钥对
type SM2KeyPair struct {
	PrivateKey string `json:"private_key"` // 私钥（十六进制字符串）
	PublicKey  string `json:"public_key"`  // 公钥（十六进制字符串）
}

// SM4Options SM4加密选项
type SM4Options struct {
	Mode    string `json:"mode"`    // 加密模式：ECB、CBC
	Padding string `json:"padding"` // 填充模式：PKCS7、ZERO
	IV      string `json:"iv"`      // 初始向量（CBC模式需要）
}

// NewSmCryptoService 创建国密服务实例
func NewSmCryptoService() *SmCryptoService {
	return &SmCryptoService{}
}

// ==================== SM2 椭圆曲线加密算法 ====================

// GenerateSM2KeyPair 生成SM2密钥对
func (s *SmCryptoService) GenerateSM2KeyPair() (*SM2KeyPair, error) {
	priv, pub, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成SM2密钥对失败: %v", err)
	}

	// 获取私钥字节
	privBytes := priv.D.Bytes()
	privHex := hex.EncodeToString(privBytes)

	// 获取公钥字节
	pubBytes := append(pub.X.Bytes(), pub.Y.Bytes()...)
	pubHex := hex.EncodeToString(pubBytes)

	return &SM2KeyPair{
		PrivateKey: privHex,
		PublicKey:  pubHex,
	}, nil
}

// SM2Encrypt SM2加密
func (s *SmCryptoService) SM2Encrypt(plaintext, publicKeyHex string) (string, error) {
	if plaintext == "" {
		return "", errors.New("明文不能为空")
	}
	if publicKeyHex == "" {
		return "", errors.New("公钥不能为空")
	}

	// 解析公钥
	pubKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return "", fmt.Errorf("公钥格式错误: %v", err)
	}

	pub, err := sm2.RawBytesToPublicKey(pubKeyBytes)
	if err != nil {
		return "", fmt.Errorf("解析公钥失败: %v", err)
	}

	// 加密
	ciphertext, err := sm2.Encrypt(pub, []byte(plaintext), sm2.C1C2C3)
	if err != nil {
		return "", fmt.Errorf("SM2加密失败: %v", err)
	}

	return hex.EncodeToString(ciphertext), nil
}

// SM2Decrypt SM2解密
func (s *SmCryptoService) SM2Decrypt(ciphertextHex, privateKeyHex string) (string, error) {
	if ciphertextHex == "" {
		return "", errors.New("密文不能为空")
	}
	if privateKeyHex == "" {
		return "", errors.New("私钥不能为空")
	}

	// 解析密文
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("密文格式错误: %v", err)
	}

	// 解析私钥
	privKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("私钥格式错误: %v", err)
	}

	priv, err := sm2.RawBytesToPrivateKey(privKeyBytes)
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %v", err)
	}

	// 解密
	plaintext, err := sm2.Decrypt(priv, ciphertext, sm2.C1C2C3)
	if err != nil {
		return "", fmt.Errorf("SM2解密失败: %v", err)
	}

	return string(plaintext), nil
}

// SM2Sign SM2签名
func (s *SmCryptoService) SM2Sign(message, privateKeyHex string) (string, error) {
	if message == "" {
		return "", errors.New("待签名消息不能为空")
	}
	if privateKeyHex == "" {
		return "", errors.New("私钥不能为空")
	}

	// 解析私钥
	privKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("私钥格式错误: %v", err)
	}

	priv, err := sm2.RawBytesToPrivateKey(privKeyBytes)
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %v", err)
	}

	// 签名（使用默认的用户ID）
	signature, err := sm2.Sign(priv, nil, []byte(message))
	if err != nil {
		return "", fmt.Errorf("SM2签名失败: %v", err)
	}

	return hex.EncodeToString(signature), nil
}

// SM2Verify SM2验签
func (s *SmCryptoService) SM2Verify(message, signatureHex, publicKeyHex string) (bool, error) {
	if message == "" {
		return false, errors.New("待验证消息不能为空")
	}
	if signatureHex == "" {
		return false, errors.New("签名不能为空")
	}
	if publicKeyHex == "" {
		return false, errors.New("公钥不能为空")
	}

	// 解析签名
	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false, fmt.Errorf("签名格式错误: %v", err)
	}

	// 解析公钥
	pubKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return false, fmt.Errorf("公钥格式错误: %v", err)
	}

	pub, err := sm2.RawBytesToPublicKey(pubKeyBytes)
	if err != nil {
		return false, fmt.Errorf("解析公钥失败: %v", err)
	}

	// 验签（使用默认的用户ID）
	return sm2.Verify(pub, nil, []byte(message), signature), nil
}

// ==================== SM3 密码杂凑算法 ====================

// SM3Hash SM3哈希
func (s *SmCryptoService) SM3Hash(data string) string {
	h := sm3.New()
	h.Write([]byte(data))
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}

// SM3HashBytes SM3哈希（字节数组）
func (s *SmCryptoService) SM3HashBytes(data []byte) string {
	h := sm3.New()
	h.Write(data)
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}

// ==================== SM4 对称加密算法 ====================

// SM4Encrypt SM4加密
func (s *SmCryptoService) SM4Encrypt(plaintext, keyHex string, options *SM4Options) (string, error) {
	if plaintext == "" {
		return "", errors.New("明文不能为空")
	}
	if keyHex == "" {
		return "", errors.New("密钥不能为空")
	}

	// 解析密钥
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("密钥格式错误: %v", err)
	}

	if len(key) != 16 {
		return "", errors.New("SM4密钥长度必须为16字节")
	}

	// 设置默认选项
	if options == nil {
		options = &SM4Options{
			Mode:    "ECB",
			Padding: "PKCS7",
		}
	}

	var ciphertext []byte
	plaintextBytes := []byte(plaintext)

	switch options.Mode {
	case "ECB":
		ciphertext, err = s.sm4EncryptECB(plaintextBytes, key, options.Padding)
	case "CBC":
		if options.IV == "" {
			return "", errors.New("CBC模式需要提供初始向量IV")
		}
		iv, ivErr := hex.DecodeString(options.IV)
		if ivErr != nil {
			return "", fmt.Errorf("IV格式错误: %v", ivErr)
		}
		if len(iv) != 16 {
			return "", errors.New("IV长度必须为16字节")
		}
		ciphertext, err = s.sm4EncryptCBC(plaintextBytes, key, iv, options.Padding)
	default:
		return "", fmt.Errorf("不支持的加密模式: %s", options.Mode)
	}

	if err != nil {
		return "", fmt.Errorf("SM4加密失败: %v", err)
	}

	return hex.EncodeToString(ciphertext), nil
}

// SM4Decrypt SM4解密
func (s *SmCryptoService) SM4Decrypt(ciphertextHex, keyHex string, options *SM4Options) (string, error) {
	if ciphertextHex == "" {
		return "", errors.New("密文不能为空")
	}
	if keyHex == "" {
		return "", errors.New("密钥不能为空")
	}

	// 解析密文
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("密文格式错误: %v", err)
	}

	// 解析密钥
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("密钥格式错误: %v", err)
	}

	if len(key) != 16 {
		return "", errors.New("SM4密钥长度必须为16字节")
	}

	// 设置默认选项
	if options == nil {
		options = &SM4Options{
			Mode:    "ECB",
			Padding: "PKCS7",
		}
	}

	var plaintext []byte

	switch options.Mode {
	case "ECB":
		plaintext, err = s.sm4DecryptECB(ciphertext, key, options.Padding)
	case "CBC":
		if options.IV == "" {
			return "", errors.New("CBC模式需要提供初始向量IV")
		}
		iv, ivErr := hex.DecodeString(options.IV)
		if ivErr != nil {
			return "", fmt.Errorf("IV格式错误: %v", ivErr)
		}
		if len(iv) != 16 {
			return "", errors.New("IV长度必须为16字节")
		}
		plaintext, err = s.sm4DecryptCBC(ciphertext, key, iv, options.Padding)
	default:
		return "", fmt.Errorf("不支持的解密模式: %s", options.Mode)
	}

	if err != nil {
		return "", fmt.Errorf("SM4解密失败: %v", err)
	}

	return string(plaintext), nil
}

// GenerateSM4Key 生成SM4密钥
func (s *SmCryptoService) GenerateSM4Key() (string, error) {
	key := make([]byte, 16)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("生成SM4密钥失败: %v", err)
	}
	return hex.EncodeToString(key), nil
}

// GenerateIV 生成初始向量
func (s *SmCryptoService) GenerateIV() (string, error) {
	iv := make([]byte, 16)
	_, err := rand.Read(iv)
	if err != nil {
		return "", fmt.Errorf("生成IV失败: %v", err)
	}
	return hex.EncodeToString(iv), nil
}

// ==================== SM4 私有方法 ====================

func (s *SmCryptoService) sm4EncryptECB(plaintext, key []byte, padding string) ([]byte, error) {
	// PKCS7填充
	if padding == "PKCS7" {
		plaintext = s.pkcs7Padding(plaintext, 16)
	}

	ciphertext := make([]byte, len(plaintext))
	cipher, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 分块加密
	for i := 0; i < len(plaintext); i += 16 {
		cipher.Encrypt(ciphertext[i:i+16], plaintext[i:i+16])
	}

	return ciphertext, nil
}

func (s *SmCryptoService) sm4DecryptECB(ciphertext, key []byte, padding string) ([]byte, error) {
	if len(ciphertext)%16 != 0 {
		return nil, errors.New("密文长度必须是16的倍数")
	}

	plaintext := make([]byte, len(ciphertext))
	cipher, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 分块解密
	for i := 0; i < len(ciphertext); i += 16 {
		cipher.Decrypt(plaintext[i:i+16], ciphertext[i:i+16])
	}

	// 去除填充
	if padding == "PKCS7" {
		plaintext = s.pkcs7UnPadding(plaintext)
	}

	return plaintext, nil
}

func (s *SmCryptoService) sm4EncryptCBC(plaintext, key, iv []byte, padding string) ([]byte, error) {
	// PKCS7填充
	if padding == "PKCS7" {
		plaintext = s.pkcs7Padding(plaintext, 16)
	}

	ciphertext := make([]byte, len(plaintext))
	cipher, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// CBC加密
	prevBlock := iv
	for i := 0; i < len(plaintext); i += 16 {
		// XOR with previous block
		for j := 0; j < 16; j++ {
			plaintext[i+j] ^= prevBlock[j]
		}
		cipher.Encrypt(ciphertext[i:i+16], plaintext[i:i+16])
		prevBlock = ciphertext[i : i+16]
	}

	return ciphertext, nil
}

func (s *SmCryptoService) sm4DecryptCBC(ciphertext, key, iv []byte, padding string) ([]byte, error) {
	if len(ciphertext)%16 != 0 {
		return nil, errors.New("密文长度必须是16的倍数")
	}

	plaintext := make([]byte, len(ciphertext))
	cipher, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// CBC解密
	prevBlock := iv
	for i := 0; i < len(ciphertext); i += 16 {
		cipher.Decrypt(plaintext[i:i+16], ciphertext[i:i+16])
		// XOR with previous block
		for j := 0; j < 16; j++ {
			plaintext[i+j] ^= prevBlock[j]
		}
		prevBlock = ciphertext[i : i+16]
	}

	// 去除填充
	if padding == "PKCS7" {
		plaintext = s.pkcs7UnPadding(plaintext)
	}

	return plaintext, nil
}

// PKCS7填充
func (s *SmCryptoService) pkcs7Padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(plaintext, padtext...)
}

// PKCS7去填充
func (s *SmCryptoService) pkcs7UnPadding(plaintext []byte) []byte {
	length := len(plaintext)
	if length == 0 {
		return plaintext
	}
	unpadding := int(plaintext[length-1])
	if unpadding > length {
		return plaintext
	}
	return plaintext[:length-unpadding]
}
