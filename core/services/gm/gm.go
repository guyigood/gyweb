package gm

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

// GMService 国密服务
type GMService struct {
	sm2KeyPair *SM2KeyPair
	config     *GMConfig
}

// GMConfig 国密配置
type GMConfig struct {
	// 密钥配置
	SM2PrivateKey string `json:"sm2_private_key,omitempty"`
	SM2PublicKey  string `json:"sm2_public_key,omitempty"`

	// 默认SM4密钥
	DefaultSM4Key string `json:"default_sm4_key,omitempty"`

	// 编码格式
	OutputFormat string `json:"output_format"` // "base64", "hex"

	// 验证选项
	EnableSignatureVerify bool `json:"enable_signature_verify"`
	EnableIntegrityCheck  bool `json:"enable_integrity_check"`
}

// EncryptRequest 加密请求
type EncryptRequest struct {
	Data      []byte            `json:"data"`
	Algorithm string            `json:"algorithm"` // "sm2", "sm4"
	Key       []byte            `json:"key,omitempty"`
	Options   map[string]string `json:"options,omitempty"`
}

// EncryptResponse 加密响应
type EncryptResponse struct {
	EncryptedData string            `json:"encrypted_data"`
	Algorithm     string            `json:"algorithm"`
	KeyInfo       map[string]string `json:"key_info,omitempty"`
	Timestamp     int64             `json:"timestamp"`
}

// DecryptRequest 解密请求
type DecryptRequest struct {
	EncryptedData string            `json:"encrypted_data"`
	Algorithm     string            `json:"algorithm"`
	Key           []byte            `json:"key,omitempty"`
	Options       map[string]string `json:"options,omitempty"`
}

// DecryptResponse 解密响应
type DecryptResponse struct {
	Data      []byte `json:"data"`
	Algorithm string `json:"algorithm"`
	Timestamp int64  `json:"timestamp"`
}

// HashRequest 哈希请求
type HashRequest struct {
	Data   []byte `json:"data"`
	Format string `json:"format"` // "hex", "base64"
}

// HashResponse 哈希响应
type HashResponse struct {
	Hash      string `json:"hash"`
	Algorithm string `json:"algorithm"`
	Format    string `json:"format"`
	Timestamp int64  `json:"timestamp"`
}

// KeyPairResponse 密钥对响应
type KeyPairResponse struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Algorithm  string `json:"algorithm"`
	Timestamp  int64  `json:"timestamp"`
}

// SignRequest 签名请求
type SignRequest struct {
	Data []byte `json:"data"`
	Key  []byte `json:"key,omitempty"`
}

// SignResponse 签名响应
type SignResponse struct {
	Signature string `json:"signature"`
	Algorithm string `json:"algorithm"`
	Timestamp int64  `json:"timestamp"`
}

// VerifyRequest 验证请求
type VerifyRequest struct {
	Data      []byte `json:"data"`
	Signature string `json:"signature"`
	PublicKey string `json:"public_key,omitempty"`
}

// VerifyResponse 验证响应
type VerifyResponse struct {
	Valid     bool   `json:"valid"`
	Algorithm string `json:"algorithm"`
	Timestamp int64  `json:"timestamp"`
}

// NewGMService 创建国密服务实例
func NewGMService(config *GMConfig) (*GMService, error) {
	if config == nil {
		config = &GMConfig{
			OutputFormat: "base64",
		}
	}

	service := &GMService{
		config: config,
	}

	// 初始化或加载SM2密钥对
	if config.SM2PrivateKey != "" && config.SM2PublicKey != "" {
		// TODO: 从配置加载现有密钥对
		keyPair, err := NewSM2KeyPair()
		if err != nil {
			return nil, fmt.Errorf("生成SM2密钥对失败: %v", err)
		}
		service.sm2KeyPair = keyPair
	} else {
		// 生成新的密钥对
		keyPair, err := NewSM2KeyPair()
		if err != nil {
			return nil, fmt.Errorf("生成SM2密钥对失败: %v", err)
		}
		service.sm2KeyPair = keyPair
	}

	return service, nil
}

// NewGMServiceDefault 创建默认配置的国密服务
func NewGMServiceDefault() (*GMService, error) {
	return NewGMService(&GMConfig{
		OutputFormat: "base64",
	})
}

// === SM2 相关方法 ===

// SM2Encrypt SM2加密
func (g *GMService) SM2Encrypt(data []byte, publicKey ...*SM2Point) (*EncryptResponse, error) {
	var pubKey *SM2Point
	if len(publicKey) > 0 && publicKey[0] != nil {
		pubKey = publicKey[0]
	} else {
		pubKey = g.sm2KeyPair.PublicKey
	}

	encrypted, err := SM2Encrypt(pubKey, data)
	if err != nil {
		return nil, err
	}

	var encryptedStr string
	if g.config.OutputFormat == "hex" {
		encryptedStr = hex.EncodeToString(encrypted)
	} else {
		encryptedStr = base64.StdEncoding.EncodeToString(encrypted)
	}

	return &EncryptResponse{
		EncryptedData: encryptedStr,
		Algorithm:     "SM2",
		Timestamp:     time.Now().Unix(),
	}, nil
}

// SM2Decrypt SM2解密
func (g *GMService) SM2Decrypt(encryptedData string) (*DecryptResponse, error) {
	var encrypted []byte
	var err error

	if g.config.OutputFormat == "hex" {
		encrypted, err = hex.DecodeString(encryptedData)
	} else {
		encrypted, err = base64.StdEncoding.DecodeString(encryptedData)
	}

	if err != nil {
		return nil, fmt.Errorf("解码加密数据失败: %v", err)
	}

	decrypted, err := SM2Decrypt(g.sm2KeyPair.PrivateKey, encrypted)
	if err != nil {
		return nil, err
	}

	return &DecryptResponse{
		Data:      decrypted,
		Algorithm: "SM2",
		Timestamp: time.Now().Unix(),
	}, nil
}

// === SM4 相关方法 ===

// SM4Encrypt SM4加密
func (g *GMService) SM4Encrypt(data []byte, key ...[]byte) (*EncryptResponse, error) {
	var sm4Key []byte

	if len(key) > 0 && len(key[0]) == 16 {
		sm4Key = key[0]
	} else if g.config.DefaultSM4Key != "" {
		var err error
		sm4Key, err = hex.DecodeString(g.config.DefaultSM4Key)
		if err != nil {
			return nil, fmt.Errorf("解析默认SM4密钥失败: %v", err)
		}
	} else {
		// 生成随机密钥
		sm4Key = make([]byte, 16)
		if _, err := rand.Read(sm4Key); err != nil {
			return nil, fmt.Errorf("生成SM4密钥失败: %v", err)
		}
	}

	encrypted, err := SM4Encrypt(sm4Key, data)
	if err != nil {
		return nil, err
	}

	var encryptedStr string
	if g.config.OutputFormat == "hex" {
		encryptedStr = hex.EncodeToString(encrypted)
	} else {
		encryptedStr = base64.StdEncoding.EncodeToString(encrypted)
	}

	keyInfo := make(map[string]string)
	if len(key) == 0 && g.config.DefaultSM4Key == "" {
		// 返回生成的密钥（仅在未提供密钥时）
		keyInfo["key"] = hex.EncodeToString(sm4Key)
	}

	return &EncryptResponse{
		EncryptedData: encryptedStr,
		Algorithm:     "SM4",
		KeyInfo:       keyInfo,
		Timestamp:     time.Now().Unix(),
	}, nil
}

// SM4Decrypt SM4解密
func (g *GMService) SM4Decrypt(encryptedData string, key ...[]byte) (*DecryptResponse, error) {
	var sm4Key []byte

	if len(key) > 0 && len(key[0]) == 16 {
		sm4Key = key[0]
	} else if g.config.DefaultSM4Key != "" {
		var err error
		sm4Key, err = hex.DecodeString(g.config.DefaultSM4Key)
		if err != nil {
			return nil, fmt.Errorf("解析默认SM4密钥失败: %v", err)
		}
	} else {
		return nil, fmt.Errorf("未提供SM4密钥")
	}

	var encrypted []byte
	var err error

	if g.config.OutputFormat == "hex" {
		encrypted, err = hex.DecodeString(encryptedData)
	} else {
		encrypted, err = base64.StdEncoding.DecodeString(encryptedData)
	}

	if err != nil {
		return nil, fmt.Errorf("解码加密数据失败: %v", err)
	}

	decrypted, err := SM4Decrypt(sm4Key, encrypted)
	if err != nil {
		return nil, err
	}

	return &DecryptResponse{
		Data:      decrypted,
		Algorithm: "SM4",
		Timestamp: time.Now().Unix(),
	}, nil
}

// === SM3 相关方法 ===

// SM3Hash SM3哈希
func (g *GMService) SM3Hash(data []byte, format ...string) (*HashResponse, error) {
	hash := SM3Hash(data)

	outputFormat := g.config.OutputFormat
	if len(format) > 0 {
		outputFormat = format[0]
	}

	var hashStr string
	if outputFormat == "hex" {
		hashStr = hex.EncodeToString(hash)
	} else {
		hashStr = base64.StdEncoding.EncodeToString(hash)
	}

	return &HashResponse{
		Hash:      hashStr,
		Algorithm: "SM3",
		Format:    outputFormat,
		Timestamp: time.Now().Unix(),
	}, nil
}

// SM3HashString 对字符串进行SM3哈希
func (g *GMService) SM3HashString(text string, format ...string) (*HashResponse, error) {
	return g.SM3Hash([]byte(text), format...)
}

// SM3Verify 验证SM3哈希
func (g *GMService) SM3Verify(data []byte, expectedHash string, format ...string) (*VerifyResponse, error) {
	hashResp, err := g.SM3Hash(data, format...)
	if err != nil {
		return nil, err
	}

	return &VerifyResponse{
		Valid:     hashResp.Hash == expectedHash,
		Algorithm: "SM3",
		Timestamp: time.Now().Unix(),
	}, nil
}

// === 密钥管理方法 ===

// GenerateSM2KeyPair 生成SM2密钥对
func (g *GMService) GenerateSM2KeyPair() (*KeyPairResponse, error) {
	keyPair, err := NewSM2KeyPair()
	if err != nil {
		return nil, err
	}

	// 更新服务的密钥对
	g.sm2KeyPair = keyPair

	return &KeyPairResponse{
		PrivateKey: g.encodePrivateKey(keyPair.PrivateKey),
		PublicKey:  g.encodePublicKey(keyPair.PublicKey),
		Algorithm:  "SM2",
		Timestamp:  time.Now().Unix(),
	}, nil
}

// GenerateSM4Key 生成SM4密钥
func (g *GMService) GenerateSM4Key() (string, error) {
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

// GetSM2PublicKey 获取SM2公钥
func (g *GMService) GetSM2PublicKey() string {
	return g.encodePublicKey(g.sm2KeyPair.PublicKey)
}

// UnmarshalHex 解析十六进制公钥字符串为 SM2Point
func (g *GMService) UnmarshalHex(hexPubKey string) (*SM2Point, error) {
	// 移除可能的 "04" 前缀（表示未压缩格式）
	if len(hexPubKey) > 2 && hexPubKey[:2] == "04" {
		hexPubKey = hexPubKey[2:]
	}

	if len(hexPubKey) != 128 { // 64 bytes * 2 = 128 hex chars
		return nil, fmt.Errorf("公钥长度不正确，期望128个十六进制字符，实际%d个", len(hexPubKey))
	}

	// 解析 X 坐标（前64个字符）
	xHex := hexPubKey[:64]
	xBytes, err := hex.DecodeString(xHex)
	if err != nil {
		return nil, fmt.Errorf("解析X坐标失败: %v", err)
	}

	// 解析 Y 坐标（后64个字符）
	yHex := hexPubKey[64:]
	yBytes, err := hex.DecodeString(yHex)
	if err != nil {
		return nil, fmt.Errorf("解析Y坐标失败: %v", err)
	}

	return &SM2Point{
		X: new(big.Int).SetBytes(xBytes),
		Y: new(big.Int).SetBytes(yBytes),
	}, nil
}

// SetPrivateKeyFromHex 从十六进制字符串设置私钥
func (g *GMService) SetPrivateKeyFromHex(hexPrivKey string) error {
	privKeyBytes, err := hex.DecodeString(hexPrivKey)
	if err != nil {
		return fmt.Errorf("解析私钥失败: %v", err)
	}

	privKey := new(big.Int).SetBytes(privKeyBytes)

	// 计算对应的公钥
	pubKey := sm2ScalarMult(privKey, &SM2Point{sm2Gx, sm2Gy})

	// 更新密钥对
	g.sm2KeyPair = &SM2KeyPair{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}

	return nil
}

// === 批量操作方法 ===

// BatchEncrypt 批量加密
func (g *GMService) BatchEncrypt(requests []*EncryptRequest) ([]*EncryptResponse, error) {
	responses := make([]*EncryptResponse, len(requests))

	for i, req := range requests {
		var resp *EncryptResponse
		var err error

		switch req.Algorithm {
		case "SM2":
			resp, err = g.SM2Encrypt(req.Data)
		case "SM4":
			resp, err = g.SM4Encrypt(req.Data, req.Key)
		default:
			err = fmt.Errorf("不支持的加密算法: %s", req.Algorithm)
		}

		if err != nil {
			return nil, fmt.Errorf("批量加密第%d项失败: %v", i+1, err)
		}

		responses[i] = resp
	}

	return responses, nil
}

// BatchDecrypt 批量解密
func (g *GMService) BatchDecrypt(requests []*DecryptRequest) ([]*DecryptResponse, error) {
	responses := make([]*DecryptResponse, len(requests))

	for i, req := range requests {
		var resp *DecryptResponse
		var err error

		switch req.Algorithm {
		case "SM2":
			resp, err = g.SM2Decrypt(req.EncryptedData)
		case "SM4":
			resp, err = g.SM4Decrypt(req.EncryptedData, req.Key)
		default:
			err = fmt.Errorf("不支持的解密算法: %s", req.Algorithm)
		}

		if err != nil {
			return nil, fmt.Errorf("批量解密第%d项失败: %v", i+1, err)
		}

		responses[i] = resp
	}

	return responses, nil
}

// BatchHash 批量哈希
func (g *GMService) BatchHash(dataList [][]byte, format ...string) ([]*HashResponse, error) {
	responses := make([]*HashResponse, len(dataList))

	for i, data := range dataList {
		resp, err := g.SM3Hash(data, format...)
		if err != nil {
			return nil, fmt.Errorf("批量哈希第%d项失败: %v", i+1, err)
		}
		responses[i] = resp
	}

	return responses, nil
}

// === 工具方法 ===

// EncryptJSON 加密JSON数据
func (g *GMService) EncryptJSON(data interface{}, algorithm string, key ...[]byte) (*EncryptResponse, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化JSON失败: %v", err)
	}

	switch algorithm {
	case "SM2":
		return g.SM2Encrypt(jsonData)
	case "SM4":
		return g.SM4Encrypt(jsonData, key...)
	default:
		return nil, fmt.Errorf("不支持的加密算法: %s", algorithm)
	}
}

// DecryptJSON 解密JSON数据
func (g *GMService) DecryptJSON(encryptedData string, algorithm string, target interface{}, key ...[]byte) error {
	var resp *DecryptResponse
	var err error

	switch algorithm {
	case "SM2":
		resp, err = g.SM2Decrypt(encryptedData)
	case "SM4":
		resp, err = g.SM4Decrypt(encryptedData, key...)
	default:
		return fmt.Errorf("不支持的解密算法: %s", algorithm)
	}

	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Data, target)
}

// SetConfig 设置配置
func (g *GMService) SetConfig(config *GMConfig) {
	if config != nil {
		g.config = config
	}
}

// GetConfig 获取配置
func (g *GMService) GetConfig() *GMConfig {
	return g.config
}

// === 私有辅助方法 ===

func (g *GMService) encodePrivateKey(privKey *big.Int) string {
	bytes := privKey.Bytes()
	if g.config.OutputFormat == "hex" {
		return hex.EncodeToString(bytes)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

func (g *GMService) encodePublicKey(pubKey *SM2Point) string {
	xBytes := pubKey.X.Bytes()
	yBytes := pubKey.Y.Bytes()
	combined := append(xBytes, yBytes...)

	if g.config.OutputFormat == "hex" {
		return hex.EncodeToString(combined)
	}
	return base64.StdEncoding.EncodeToString(combined)
}
