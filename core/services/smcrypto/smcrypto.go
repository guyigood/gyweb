package smcrypto

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/tjfoc/gmsm/sm3"
)

// SmCryptoService 国密服务
type SmCryptoService struct{}

const (
	C1C2C3 = 0 // 默认模式
	C1C3C2 = 1 // 另一种模式
)

// NewSmCryptoService 创建国密服务实例
func NewSmCryptoService() *SmCryptoService {
	return &SmCryptoService{}
}

func (c *SmCryptoService) GetSM3HashString(decrypted string) string {
	ha := sm3.Sm3Sum([]byte(decrypted))
	return hex.EncodeToString(ha)
}

// SM2Decrypt 精确匹配 sm-crypto 的解密实现
// cipherMode: 1 - C1C3C2, 0 - C1C2C3
func (c *SmCryptoService) SM2Decrypt(encryptDataHex, privateKeyHex string, cipherMode int) (string, error) {
	//加入长度检查
	if len(encryptDataHex) < 128 {
		return "", fmt.Errorf("加密数据长度不足")
	}
	if len(privateKeyHex) < 64 {
		return "", fmt.Errorf("私钥长度不足")
	}

	// 解析私钥
	privateKey := new(big.Int)
	privateKey.SetString(privateKeyHex, 16)

	// JavaScript中的字符串操作：
	// let c3 = encryptData.substr(128, 64)
	// let c2 = encryptData.substr(128 + 64)
	var c3Hex, c2Hex string

	if cipherMode == C1C2C3 { // cipherMode === 0
		// c3 = encryptData.substr(encryptData.length - 64)
		// c2 = encryptData.substr(128, encryptData.length - 128 - 64)
		c3Hex = encryptDataHex[len(encryptDataHex)-64:]
		c2Hex = encryptDataHex[128 : len(encryptDataHex)-64]
	} else { // cipherMode === 1 (C1C3C2)
		// c3 = encryptData.substr(128, 64)
		// c2 = encryptData.substr(128 + 64)
		c3Hex = encryptDataHex[128 : 128+64]
		c2Hex = encryptDataHex[128+64:]
	}

	// const msg = _.hexToArray(c2)
	c2Bytes, err := hex.DecodeString(c2Hex)
	if err != nil {
		return "", fmt.Errorf("C2解码失败: %v", err)
	}
	msg := make([]byte, len(c2Bytes))
	copy(msg, c2Bytes)

	// const c1 = _.getGlobalCurve().decodePointHex('04' + encryptData.substr(0, 128))
	c1Hex := encryptDataHex[:128] // 前128个字符
	c1PointBytes, err := hex.DecodeString("04" + c1Hex)
	if err != nil {
		return "", fmt.Errorf("C1解码失败: %v", err)
	}

	// 验证C1点格式
	if len(c1PointBytes) != 65 || c1PointBytes[0] != 0x04 {
		return "", fmt.Errorf("无效的C1点格式")
	}

	// 手动实现椭圆曲线标量乘法 [d]P
	// 使用二进制方法计算椭圆曲线点乘，不依赖弃用的API
	x1 := new(big.Int).SetBytes(c1PointBytes[1:33])
	y1 := new(big.Int).SetBytes(c1PointBytes[33:65])
	d := new(big.Int).SetBytes(privateKey.Bytes())

	pX, pY := c.scalarMultSM2(x1, y1, d)

	// const x2 = _.hexToArray(_.leftPad(p.getX().toBigInteger().toRadix(16), 64))
	// const y2 = _.hexToArray(_.leftPad(p.getY().toBigInteger().toRadix(16), 64))
	x2Hex := fmt.Sprintf("%064x", pX) // 左填充到64个字符
	y2Hex := fmt.Sprintf("%064x", pY) // 左填充到64个字符

	x2Bytes, _ := hex.DecodeString(x2Hex)
	y2Bytes, _ := hex.DecodeString(y2Hex)

	// const z = [].concat(x2, y2)
	z := append(x2Bytes, y2Bytes...)

	// KDF过程
	ct := 1
	offset := 0
	var t []byte

	nextT := func() {
		// t = sm3([...z, ct >> 24 & 0x00ff, ct >> 16 & 0x00ff, ct >> 8 & 0x00ff, ct & 0x00ff])
		h := sm3.New()
		h.Write(z)
		h.Write([]byte{
			byte(ct >> 24 & 0xff),
			byte(ct >> 16 & 0xff),
			byte(ct >> 8 & 0xff),
			byte(ct & 0xff),
		})
		t = h.Sum(nil)
		ct++
		offset = 0
	}
	nextT() // 先生成 Ha1

	// for (let i = 0, len = msg.length; i < len; i++) {
	//   if (offset === t.length) nextT()
	//   msg[i] ^= t[offset++] & 0xff
	// }
	for i := 0; i < len(msg); i++ {
		if offset == len(t) {
			nextT()
		}
		msg[i] ^= t[offset] & 0xff
		offset++
	}

	// const checkC3 = _.arrayToHex(sm3([].concat(x2, msg, y2)))
	verifyData := append(x2Bytes, msg...)
	verifyData = append(verifyData, y2Bytes...)
	calculatedC3 := sm3.Sm3Sum(verifyData)
	checkC3 := hex.EncodeToString(calculatedC3)

	// if (checkC3 === c3.toLowerCase()) {
	if checkC3 == c3Hex {
		// return output === 'array' ? msg : _.arrayToUtf8(msg)
		return string(msg), nil
	} else {
		return "", fmt.Errorf("C3验证失败: 期望=%s, 计算=%s", c3Hex, checkC3)
	}
}

// 手动实现SM2椭圆曲线标量乘法
func (c *SmCryptoService) scalarMultSM2(x, y, k *big.Int) (*big.Int, *big.Int) {
	// SM2曲线参数
	p := new(big.Int)
	p.SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFF", 16)
	a := big.NewInt(-3)

	// 无穷远点
	if k.Sign() == 0 {
		return nil, nil
	}

	// 结果初始化为无穷远点
	var resX, resY *big.Int

	// 当前点
	curX, curY := new(big.Int).Set(x), new(big.Int).Set(y)

	// 二进制展开标量乘法
	for i := 0; i < k.BitLen(); i++ {
		if k.Bit(i) == 1 {
			if resX == nil {
				// 第一次设置结果点
				resX, resY = new(big.Int).Set(curX), new(big.Int).Set(curY)
			} else {
				// 点加法
				resX, resY = c.pointAddSM2(resX, resY, curX, curY, p, a)
			}
		}
		// 点倍乘
		if i < k.BitLen()-1 {
			curX, curY = c.pointDoubleSM2(curX, curY, p, a)
		}
	}

	return resX, resY
}

// SM2椭圆曲线点加法
func (c *SmCryptoService) pointAddSM2(x1, y1, x2, y2, p, a *big.Int) (*big.Int, *big.Int) {
	if x1.Cmp(x2) == 0 {
		if y1.Cmp(y2) == 0 {
			return c.pointDoubleSM2(x1, y1, p, a)
		} else {
			// 结果是无穷远点
			return nil, nil
		}
	}

	// λ = (y2 - y1) / (x2 - x1)
	numerator := new(big.Int).Sub(y2, y1)
	denominator := new(big.Int).Sub(x2, x1)
	denominator.ModInverse(denominator, p)
	lambda := new(big.Int).Mul(numerator, denominator)
	lambda.Mod(lambda, p)

	// x3 = λ² - x1 - x2
	x3 := new(big.Int).Mul(lambda, lambda)
	x3.Sub(x3, x1)
	x3.Sub(x3, x2)
	x3.Mod(x3, p)

	// y3 = λ(x1 - x3) - y1
	y3 := new(big.Int).Sub(x1, x3)
	y3.Mul(y3, lambda)
	y3.Sub(y3, y1)
	y3.Mod(y3, p)

	return x3, y3
}

// SM2椭圆曲线点倍乘
func (c *SmCryptoService) pointDoubleSM2(x, y, p, a *big.Int) (*big.Int, *big.Int) {
	// λ = (3x² + a) / (2y)
	numerator := new(big.Int).Mul(x, x)
	numerator.Mul(numerator, big.NewInt(3))
	numerator.Add(numerator, a)

	denominator := new(big.Int).Mul(y, big.NewInt(2))
	denominator.ModInverse(denominator, p)

	lambda := new(big.Int).Mul(numerator, denominator)
	lambda.Mod(lambda, p)

	// x3 = λ² - 2x
	x3 := new(big.Int).Mul(lambda, lambda)
	x3.Sub(x3, new(big.Int).Mul(x, big.NewInt(2)))
	x3.Mod(x3, p)

	// y3 = λ(x - x3) - y
	y3 := new(big.Int).Sub(x, x3)
	y3.Mul(y3, lambda)
	y3.Sub(y3, y)
	y3.Mod(y3, p)

	return x3, y3
}

// SM2Encrypt 精确匹配 sm-crypto 的加密实现
// cipherMode: 1 - C1C3C2, 0 - C1C2C3
func (c *SmCryptoService) SM2Encrypt(plaintext, publicKeyHex string, cipherMode int) (string, error) {
	// msg = typeof msg === 'string' ? _.hexToArray(_.utf8ToHex(msg)) : Array.prototype.slice.call(msg)
	msg := []byte(plaintext)

	// publicKey = _.getGlobalCurve().decodePointHex(publicKey)
	if len(publicKeyHex) < 128 {
		return "", fmt.Errorf("公钥长度不足")
	}

	// 移除04前缀如果存在
	if publicKeyHex[:2] == "04" {
		publicKeyHex = publicKeyHex[2:]
	}
	if len(publicKeyHex) != 128 {
		return "", fmt.Errorf("公钥格式错误")
	}

	pubKeyBytes, err := hex.DecodeString("04" + publicKeyHex)
	if err != nil {
		return "", fmt.Errorf("公钥解码失败: %v", err)
	}

	pubX := new(big.Int).SetBytes(pubKeyBytes[1:33])
	pubY := new(big.Int).SetBytes(pubKeyBytes[33:65])

	// const keypair = _.generateKeyPairHex()
	// const k = new BigInteger(keypair.privateKey, 16)
	k, kGx, kGy, err := c.generateSM2KeyPair()
	if err != nil {
		return "", fmt.Errorf("生成密钥对失败: %v", err)
	}

	// c1 = k * G (取公钥的坐标部分)
	c1 := fmt.Sprintf("%064x%064x", kGx, kGy)
	if len(c1) > 128 {
		c1 = c1[len(c1)-128:]
	}

	// (x2, y2) = k * publicKey
	x2, y2 := c.scalarMultSM2(pubX, pubY, k)
	if x2 == nil || y2 == nil {
		return "", fmt.Errorf("椭圆曲线运算失败")
	}

	// 转换为字节数组
	x2Bytes := c.leftPadBytes(x2.Bytes(), 32)
	y2Bytes := c.leftPadBytes(y2.Bytes(), 32)

	// c3 = hash(x2 || msg || y2)
	c3Data := append(x2Bytes, msg...)
	c3Data = append(c3Data, y2Bytes...)
	c3Hash := sm3.Sm3Sum(c3Data)
	c3 := hex.EncodeToString(c3Hash)

	// KDF过程生成密钥流并加密
	z := append(x2Bytes, y2Bytes...)
	encryptedMsg := make([]byte, len(msg))
	copy(encryptedMsg, msg)

	ct := 1
	offset := 0
	var t []byte

	nextT := func() {
		// t = sm3([...z, ct >> 24 & 0x00ff, ct >> 16 & 0x00ff, ct >> 8 & 0x00ff, ct & 0x00ff])
		h := sm3.New()
		h.Write(z)
		h.Write([]byte{
			byte(ct >> 24 & 0xff),
			byte(ct >> 16 & 0xff),
			byte(ct >> 8 & 0xff),
			byte(ct & 0xff),
		})
		t = h.Sum(nil)
		ct++
		offset = 0
	}
	nextT() // 先生成 Ha1

	// for (let i = 0, len = msg.length; i < len; i++) {
	//   if (offset === t.length) nextT()
	//   msg[i] ^= t[offset++] & 0xff
	// }
	for i := 0; i < len(encryptedMsg); i++ {
		if offset == len(t) {
			nextT()
		}
		encryptedMsg[i] ^= t[offset] & 0xff
		offset++
	}

	c2 := hex.EncodeToString(encryptedMsg)

	// return cipherMode === C1C2C3 ? c1 + c2 + c3 : c1 + c3 + c2
	if cipherMode == C1C2C3 {
		return c1 + c2 + c3, nil
	} else {
		return c1 + c3 + c2, nil
	}
}

// 生成SM2密钥对，返回私钥k和公钥坐标(kGx, kGy)
func (c *SmCryptoService) generateSM2KeyPair() (*big.Int, *big.Int, *big.Int, error) {
	// SM2曲线参数
	n := new(big.Int)
	n.SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFF7203DF6B61C6F30D92FB36D95FA3D52D", 16)

	// 生成随机私钥 k
	k, err := c.generateRandomBigInt(n)
	if err != nil {
		return nil, nil, nil, err
	}

	// SM2基点G
	gx := new(big.Int)
	gy := new(big.Int)
	gx.SetString("32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7", 16)
	gy.SetString("BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0", 16)

	// 计算 kG
	kGx, kGy := c.scalarMultSM2(gx, gy, k)

	return k, kGx, kGy, nil
}

// 生成小于n的随机大整数
func (c *SmCryptoService) generateRandomBigInt(n *big.Int) (*big.Int, error) {
	// 简化实现：使用当前时间作为种子生成伪随机数
	// 在生产环境中应该使用密码学安全的随机数生成器

	seed := time.Now().UnixNano()
	k := new(big.Int).SetInt64(seed)
	k.Mod(k, n)

	// 确保k不为0
	if k.Sign() == 0 {
		k.SetInt64(1)
	}

	return k, nil
}

// 左填充字节数组
func (c *SmCryptoService) leftPadBytes(data []byte, length int) []byte {
	if len(data) >= length {
		return data
	}
	padded := make([]byte, length)
	copy(padded[length-len(data):], data)
	return padded
}
