package gm

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// GMUtil 国密工具类
type GMUtil struct {
	service *GMService
}

// NewGMUtil 创建国密工具实例
func NewGMUtil(service *GMService) *GMUtil {
	return &GMUtil{service: service}
}

// === 快速加解密函数 ===

// QuickSM2Encrypt 快速SM2加密
func QuickSM2Encrypt(data []byte) (string, error) {
	service, err := NewGMServiceDefault()
	if err != nil {
		return "", err
	}

	resp, err := service.SM2Encrypt(data)
	if err != nil {
		return "", err
	}

	return resp.EncryptedData, nil
}

// QuickSM2EncryptString 快速SM2加密字符串
func QuickSM2EncryptString(text string) (string, error) {
	return QuickSM2Encrypt([]byte(text))
}

// QuickSM4Encrypt 快速SM4加密
func QuickSM4Encrypt(data []byte, key []byte) (string, error) {
	service, err := NewGMServiceDefault()
	if err != nil {
		return "", err
	}

	resp, err := service.SM4Encrypt(data, key)
	if err != nil {
		return "", err
	}

	return resp.EncryptedData, nil
}

// QuickSM4EncryptString 快速SM4加密字符串
func QuickSM4EncryptString(text string, key []byte) (string, error) {
	return QuickSM4Encrypt([]byte(text), key)
}

// QuickSM3Hash 快速SM3哈希
func QuickSM3Hash(data []byte, format ...string) (string, error) {
	service, err := NewGMServiceDefault()
	if err != nil {
		return "", err
	}

	resp, err := service.SM3Hash(data, format...)
	if err != nil {
		return "", err
	}

	return resp.Hash, nil
}

// QuickSM3HashString 快速SM3哈希字符串
func QuickSM3HashString(text string, format ...string) (string, error) {
	return QuickSM3Hash([]byte(text), format...)
}

// === 配置预设 ===

// GetGMConfigDefault 获取默认配置
func GetGMConfigDefault() *GMConfig {
	return &GMConfig{
		OutputFormat:          "base64",
		EnableSignatureVerify: false,
		EnableIntegrityCheck:  false,
	}
}

// GetGMConfigSecure 获取安全配置
func GetGMConfigSecure() *GMConfig {
	return &GMConfig{
		OutputFormat:          "hex",
		EnableSignatureVerify: true,
		EnableIntegrityCheck:  true,
	}
}

// GetGMConfigPerformance 获取性能优化配置
func GetGMConfigPerformance() *GMConfig {
	return &GMConfig{
		OutputFormat:          "base64",
		EnableSignatureVerify: false,
		EnableIntegrityCheck:  false,
	}
}

// GetGMConfigCustom 获取自定义配置
func GetGMConfigCustom(outputFormat string, sm4Key string) *GMConfig {
	return &GMConfig{
		OutputFormat:  outputFormat,
		DefaultSM4Key: sm4Key,
	}
}

// === 辅助工具函数 ===

// GenerateRandomKey 生成随机密钥
func GenerateRandomKey(length int) ([]byte, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	return key, err
}

// GenerateRandomKeyHex 生成十六进制随机密钥
func GenerateRandomKeyHex(length int) (string, error) {
	key, err := GenerateRandomKey(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

// GenerateRandomKeyBase64 生成Base64随机密钥
func GenerateRandomKeyBase64(length int) (string, error) {
	key, err := GenerateRandomKey(length)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// ValidateKeyLength 验证密钥长度
func ValidateKeyLength(key []byte, algorithm string) error {
	switch strings.ToUpper(algorithm) {
	case "SM4":
		if len(key) != 16 {
			return fmt.Errorf("SM4密钥长度必须为16字节，当前为%d字节", len(key))
		}
	case "AES128":
		if len(key) != 16 {
			return fmt.Errorf("AES128密钥长度必须为16字节，当前为%d字节", len(key))
		}
	case "AES256":
		if len(key) != 32 {
			return fmt.Errorf("AES256密钥长度必须为32字节，当前为%d字节", len(key))
		}
	default:
		return fmt.Errorf("不支持的算法: %s", algorithm)
	}
	return nil
}

// EncodeHex 编码为十六进制
func EncodeHex(data []byte) string {
	return hex.EncodeToString(data)
}

// DecodeHex 解码十六进制
func DecodeHex(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

// EncodeBase64 编码为Base64
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 解码Base64
func DecodeBase64(base64Str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64Str)
}

// === 数据格式转换 ===

// ConvertFormat 转换数据格式
func ConvertFormat(data string, fromFormat, toFormat string) (string, error) {
	var bytes []byte
	var err error

	// 解码原格式
	switch strings.ToLower(fromFormat) {
	case "hex":
		bytes, err = hex.DecodeString(data)
	case "base64":
		bytes, err = base64.StdEncoding.DecodeString(data)
	default:
		return "", fmt.Errorf("不支持的源格式: %s", fromFormat)
	}

	if err != nil {
		return "", fmt.Errorf("解码失败: %v", err)
	}

	// 编码为目标格式
	switch strings.ToLower(toFormat) {
	case "hex":
		return hex.EncodeToString(bytes), nil
	case "base64":
		return base64.StdEncoding.EncodeToString(bytes), nil
	default:
		return "", fmt.Errorf("不支持的目标格式: %s", toFormat)
	}
}

// === 密钥工具 ===

// ValidateSM4Key 验证SM4密钥
func ValidateSM4Key(key []byte) error {
	if len(key) != 16 {
		return fmt.Errorf("SM4密钥长度必须为16字节")
	}
	return nil
}

// ValidateSM4KeyString 验证SM4密钥字符串
func ValidateSM4KeyString(keyStr, format string) error {
	var key []byte
	var err error

	switch strings.ToLower(format) {
	case "hex":
		key, err = hex.DecodeString(keyStr)
	case "base64":
		key, err = base64.StdEncoding.DecodeString(keyStr)
	default:
		return fmt.Errorf("不支持的密钥格式: %s", format)
	}

	if err != nil {
		return fmt.Errorf("解析密钥失败: %v", err)
	}

	return ValidateSM4Key(key)
}

// ParseSM4Key 解析SM4密钥
func ParseSM4Key(keyStr, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "hex":
		return hex.DecodeString(keyStr)
	case "base64":
		return base64.StdEncoding.DecodeString(keyStr)
	default:
		return nil, fmt.Errorf("不支持的密钥格式: %s", format)
	}
}

// === 工具类方法 ===

// (u *GMUtil) QuickEncrypt 工具类快速加密
func (u *GMUtil) QuickEncrypt(data []byte, algorithm string, key ...[]byte) (string, error) {
	switch strings.ToUpper(algorithm) {
	case "SM2":
		resp, err := u.service.SM2Encrypt(data)
		if err != nil {
			return "", err
		}
		return resp.EncryptedData, nil
	case "SM4":
		resp, err := u.service.SM4Encrypt(data, key...)
		if err != nil {
			return "", err
		}
		return resp.EncryptedData, nil
	default:
		return "", fmt.Errorf("不支持的加密算法: %s", algorithm)
	}
}

// (u *GMUtil) QuickDecrypt 工具类快速解密
func (u *GMUtil) QuickDecrypt(encryptedData, algorithm string, key ...[]byte) ([]byte, error) {
	switch strings.ToUpper(algorithm) {
	case "SM2":
		resp, err := u.service.SM2Decrypt(encryptedData)
		if err != nil {
			return nil, err
		}
		return resp.Data, nil
	case "SM4":
		resp, err := u.service.SM4Decrypt(encryptedData, key...)
		if err != nil {
			return nil, err
		}
		return resp.Data, nil
	default:
		return nil, fmt.Errorf("不支持的解密算法: %s", algorithm)
	}
}

// (u *GMUtil) QuickHash 工具类快速哈希
func (u *GMUtil) QuickHash(data []byte, format ...string) (string, error) {
	resp, err := u.service.SM3Hash(data, format...)
	if err != nil {
		return "", err
	}
	return resp.Hash, nil
}

// === 批处理工具 ===

// BatchProcessFiles 批量处理文件（概念性实现）
type FileProcessor struct {
	Algorithm string
	Key       []byte
	Operation string // "encrypt", "decrypt", "hash"
}

// ProcessData 处理数据
func (fp *FileProcessor) ProcessData(data []byte) ([]byte, error) {
	service, err := NewGMServiceDefault()
	if err != nil {
		return nil, err
	}

	switch fp.Operation {
	case "encrypt":
		if fp.Algorithm == "SM2" {
			resp, err := service.SM2Encrypt(data)
			if err != nil {
				return nil, err
			}
			return []byte(resp.EncryptedData), nil
		} else if fp.Algorithm == "SM4" {
			resp, err := service.SM4Encrypt(data, fp.Key)
			if err != nil {
				return nil, err
			}
			return []byte(resp.EncryptedData), nil
		}
	case "decrypt":
		if fp.Algorithm == "SM2" {
			resp, err := service.SM2Decrypt(string(data))
			if err != nil {
				return nil, err
			}
			return resp.Data, nil
		} else if fp.Algorithm == "SM4" {
			resp, err := service.SM4Decrypt(string(data), fp.Key)
			if err != nil {
				return nil, err
			}
			return resp.Data, nil
		}
	case "hash":
		resp, err := service.SM3Hash(data)
		if err != nil {
			return nil, err
		}
		return []byte(resp.Hash), nil
	}

	return nil, fmt.Errorf("不支持的操作: %s", fp.Operation)
}

// === 性能测试工具 ===

// BenchmarkConfig 性能测试配置
type BenchmarkConfig struct {
	DataSize   int    // 数据大小（字节）
	Iterations int    // 迭代次数
	Algorithm  string // 算法类型
}

// RunBenchmark 运行性能测试（简化实现）
func RunBenchmark(config *BenchmarkConfig) (*BenchmarkResult, error) {
	// 这里只是一个概念性实现，实际应该使用Go的testing包
	service, err := NewGMServiceDefault()
	if err != nil {
		return nil, err
	}

	testData := make([]byte, config.DataSize)
	rand.Read(testData)

	start := time.Now()

	for i := 0; i < config.Iterations; i++ {
		switch config.Algorithm {
		case "SM2":
			_, err := service.SM2Encrypt(testData)
			if err != nil {
				return nil, err
			}
		case "SM4":
			key, _ := GenerateRandomKey(16)
			_, err := service.SM4Encrypt(testData, key)
			if err != nil {
				return nil, err
			}
		case "SM3":
			_, err := service.SM3Hash(testData)
			if err != nil {
				return nil, err
			}
		}
	}

	duration := time.Since(start)

	return &BenchmarkResult{
		Algorithm:      config.Algorithm,
		DataSize:       config.DataSize,
		Iterations:     config.Iterations,
		TotalDuration:  duration,
		AvgDuration:    duration / time.Duration(config.Iterations),
		ThroughputMBps: float64(config.DataSize*config.Iterations) / 1024 / 1024 / duration.Seconds(),
	}, nil
}

// BenchmarkResult 性能测试结果
type BenchmarkResult struct {
	Algorithm      string        `json:"algorithm"`
	DataSize       int           `json:"data_size"`
	Iterations     int           `json:"iterations"`
	TotalDuration  time.Duration `json:"total_duration"`
	AvgDuration    time.Duration `json:"avg_duration"`
	ThroughputMBps float64       `json:"throughput_mbps"`
}
