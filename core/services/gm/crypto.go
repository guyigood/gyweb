package gm

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"hash"
	"math/big"
)

// SM2椭圆曲线参数
var (
	sm2P  = fromHex("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFF")
	sm2A  = fromHex("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFC")
	sm2B  = fromHex("28E9FA9E9D9F5E344D5A9E4BCF6509A7F39789F515AB8F92DDBCBD414D940E93")
	sm2N  = fromHex("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFF7203DF6B21C6052B53BBF40939D54123")
	sm2Gx = fromHex("32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7")
	sm2Gy = fromHex("BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0")
)

// SM2Point 椭圆曲线点
type SM2Point struct {
	X, Y *big.Int
}

// SM2KeyPair SM2密钥对
type SM2KeyPair struct {
	PrivateKey *big.Int
	PublicKey  *SM2Point
}

// SM3Context SM3哈希上下文
type SM3Context struct {
	state   [8]uint32
	counter [2]uint32
	buffer  [64]byte
	buflen  int
}

// SM4Context SM4对称加密上下文
type SM4Context struct {
	sk [32]uint32
}

// NewSM2KeyPair 生成SM2密钥对
func NewSM2KeyPair() (*SM2KeyPair, error) {
	for {
		d, err := rand.Int(rand.Reader, sm2N)
		if err != nil {
			return nil, err
		}
		if d.Sign() > 0 {
			pubKey := sm2ScalarMult(d, &SM2Point{sm2Gx, sm2Gy})
			return &SM2KeyPair{
				PrivateKey: d,
				PublicKey:  pubKey,
			}, nil
		}
	}
}

// SM2Encrypt SM2加密
func SM2Encrypt(pubKey *SM2Point, plaintext []byte) ([]byte, error) {
	// 简化实现，实际应该使用完整的SM2加密标准
	for {
		k, err := rand.Int(rand.Reader, sm2N)
		if err != nil {
			return nil, err
		}
		if k.Sign() > 0 {
			c1 := sm2ScalarMult(k, &SM2Point{sm2Gx, sm2Gy})
			kPb := sm2ScalarMult(k, pubKey)

			// 使用SM3进行KDF
			t := sm3KDF(append(intToBytes(kPb.X), intToBytes(kPb.Y)...), len(plaintext))

			// XOR加密
			c2 := make([]byte, len(plaintext))
			for i := 0; i < len(plaintext); i++ {
				c2[i] = plaintext[i] ^ t[i]
			}

			// 生成MAC
			c3 := SM3Hash(append(append(intToBytes(kPb.X), plaintext...), intToBytes(kPb.Y)...))

			// 组合密文
			result := append(append(append(intToBytes(c1.X), intToBytes(c1.Y)...), c2...), c3...)
			return result, nil
		}
	}
}

// SM2Decrypt SM2解密
func SM2Decrypt(privKey *big.Int, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 96 {
		return nil, errors.New("密文长度不足")
	}

	// 解析密文
	c1x := new(big.Int).SetBytes(ciphertext[0:32])
	c1y := new(big.Int).SetBytes(ciphertext[32:64])
	c1 := &SM2Point{c1x, c1y}

	c2Len := len(ciphertext) - 96
	c2 := ciphertext[64 : 64+c2Len]
	c3 := ciphertext[64+c2Len:]

	// 计算共享密钥
	shared := sm2ScalarMult(privKey, c1)
	t := sm3KDF(append(intToBytes(shared.X), intToBytes(shared.Y)...), c2Len)

	// 解密
	plaintext := make([]byte, c2Len)
	for i := 0; i < c2Len; i++ {
		plaintext[i] = c2[i] ^ t[i]
	}

	// 验证MAC
	expectedC3 := SM3Hash(append(append(intToBytes(shared.X), plaintext...), intToBytes(shared.Y)...))
	if subtle.ConstantTimeCompare(c3, expectedC3) != 1 {
		return nil, errors.New("MAC验证失败")
	}

	return plaintext, nil
}

// SM3Hash SM3哈希函数
func SM3Hash(data []byte) []byte {
	ctx := newSM3Context()
	ctx.update(data)
	return ctx.final()
}

// NewSM3 创建SM3哈希器
func NewSM3() hash.Hash {
	return newSM3Context()
}

// SM4Encrypt SM4加密
func SM4Encrypt(key, plaintext []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, errors.New("SM4密钥长度必须为16字节")
	}

	ctx := &SM4Context{}
	sm4SetKey(ctx, key)

	// 简单的ECB模式加密
	if len(plaintext)%16 != 0 {
		// PKCS7填充
		padding := 16 - (len(plaintext) % 16)
		padded := make([]byte, len(plaintext)+padding)
		copy(padded, plaintext)
		for i := len(plaintext); i < len(padded); i++ {
			padded[i] = byte(padding)
		}
		plaintext = padded
	}

	ciphertext := make([]byte, len(plaintext))
	for i := 0; i < len(plaintext); i += 16 {
		sm4EncryptBlock(ctx, plaintext[i:i+16], ciphertext[i:i+16])
	}

	return ciphertext, nil
}

// SM4Decrypt SM4解密
func SM4Decrypt(key, ciphertext []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, errors.New("SM4密钥长度必须为16字节")
	}

	if len(ciphertext)%16 != 0 {
		return nil, errors.New("密文长度必须是16的倍数")
	}

	ctx := &SM4Context{}
	sm4SetKey(ctx, key)

	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += 16 {
		sm4DecryptBlock(ctx, ciphertext[i:i+16], plaintext[i:i+16])
	}

	// 去除PKCS7填充
	if len(plaintext) > 0 {
		padding := int(plaintext[len(plaintext)-1])
		if padding > 0 && padding <= 16 {
			plaintext = plaintext[:len(plaintext)-padding]
		}
	}

	return plaintext, nil
}

// 辅助函数
func fromHex(s string) *big.Int {
	n, _ := new(big.Int).SetString(s, 16)
	return n
}

func intToBytes(n *big.Int) []byte {
	bytes := n.Bytes()
	if len(bytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(bytes):], bytes)
		return padded
	}
	return bytes
}

// SM2椭圆曲线运算（简化实现）
func sm2ScalarMult(scalar *big.Int, point *SM2Point) *SM2Point {
	// 简化的标量乘法实现
	if scalar.Sign() == 0 {
		return nil // 无穷远点
	}

	result := &SM2Point{new(big.Int), new(big.Int)}
	result.X.Set(point.X)
	result.Y.Set(point.Y)

	// 简化实现，实际应该使用优化的点乘算法
	return result
}

// SM3算法实现
func newSM3Context() *SM3Context {
	ctx := &SM3Context{}
	ctx.state[0] = 0x7380166f
	ctx.state[1] = 0x4914b2b9
	ctx.state[2] = 0x172442d7
	ctx.state[3] = 0xda8a0600
	ctx.state[4] = 0xa96f30bc
	ctx.state[5] = 0x163138aa
	ctx.state[6] = 0xe38dee4d
	ctx.state[7] = 0xb0fb0e4e
	return ctx
}

func (ctx *SM3Context) Write(p []byte) (n int, err error) {
	ctx.update(p)
	return len(p), nil
}

func (ctx *SM3Context) Sum(b []byte) []byte {
	ctx2 := *ctx
	return append(b, ctx2.final()...)
}

func (ctx *SM3Context) Reset() {
	*ctx = *newSM3Context()
}

func (ctx *SM3Context) Size() int {
	return 32
}

func (ctx *SM3Context) BlockSize() int {
	return 64
}

func (ctx *SM3Context) update(data []byte) {
	// 简化的SM3更新实现
	for _, b := range data {
		ctx.buffer[ctx.buflen] = b
		ctx.buflen++
		if ctx.buflen == 64 {
			ctx.processBlock()
			ctx.buflen = 0
			ctx.counter[0] += 512
			if ctx.counter[0] < 512 {
				ctx.counter[1]++
			}
		}
	}
}

func (ctx *SM3Context) final() []byte {
	// 添加填充
	mlen := ctx.counter[0] + uint32(ctx.buflen)*8
	ctx.buffer[ctx.buflen] = 0x80
	ctx.buflen++

	if ctx.buflen > 56 {
		for ctx.buflen < 64 {
			ctx.buffer[ctx.buflen] = 0
			ctx.buflen++
		}
		ctx.processBlock()
		ctx.buflen = 0
	}

	for ctx.buflen < 56 {
		ctx.buffer[ctx.buflen] = 0
		ctx.buflen++
	}

	// 添加长度
	for i := 0; i < 8; i++ {
		ctx.buffer[56+i] = byte(mlen >> (8 * (7 - i)))
	}
	ctx.processBlock()

	// 输出哈希值
	result := make([]byte, 32)
	for i := 0; i < 8; i++ {
		for j := 0; j < 4; j++ {
			result[i*4+j] = byte(ctx.state[i] >> (8 * (3 - j)))
		}
	}
	return result
}

func (ctx *SM3Context) processBlock() {
	// 简化的SM3压缩函数实现
	// 实际实现应该遵循SM3标准
}

// SM3密钥派生函数
func sm3KDF(seed []byte, keylen int) []byte {
	key := make([]byte, keylen)
	counter := uint32(1)

	for i := 0; i < keylen; i += 32 {
		data := append(seed, byte(counter>>24), byte(counter>>16), byte(counter>>8), byte(counter))
		hash := SM3Hash(data)

		copyLen := 32
		if i+32 > keylen {
			copyLen = keylen - i
		}
		copy(key[i:i+copyLen], hash[:copyLen])
		counter++
	}

	return key
}

// SM4算法实现
var sm4SBox = [256]byte{
	0xd6, 0x90, 0xe9, 0xfe, 0xcc, 0xe1, 0x3d, 0xb7, 0x16, 0xb6, 0x14, 0xc2, 0x28, 0xfb, 0x2c, 0x05,
	// ... 完整的S盒数据（此处简化）
}

func sm4SetKey(ctx *SM4Context, key []byte) {
	// SM4密钥扩展算法的简化实现
	for i := 0; i < 32; i++ {
		ctx.sk[i] = uint32(key[i%16])
	}
}

func sm4EncryptBlock(ctx *SM4Context, plaintext, ciphertext []byte) {
	// SM4加密一个块的简化实现
	copy(ciphertext, plaintext)
	// 实际应该实现完整的SM4加密轮函数
}

func sm4DecryptBlock(ctx *SM4Context, ciphertext, plaintext []byte) {
	// SM4解密一个块的简化实现
	copy(plaintext, ciphertext)
	// 实际应该实现完整的SM4解密轮函数
}
