package captcha

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	Width      int        // 图片宽度
	Height     int        // 图片高度
	NoiseCount int        // 噪点数量
	NoiseLevel float64    // 噪声强度
	FontSize   int        // 字体大小
	BgColor    color.RGBA // 背景颜色
	TextColor  color.RGBA // 文字颜色
	NoiseColor color.RGBA // 噪点颜色
}

// 创建一个全局的随机源，避免重复调用已弃用的rand.Seed
var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))

// DefaultConfig 默认配置
func DefaultConfig() *CaptchaConfig {
	return &CaptchaConfig{
		Width:      120,
		Height:     40,
		NoiseCount: 50,
		NoiseLevel: 0.3,
		FontSize:   24,
		BgColor:    color.RGBA{R: 240, G: 240, B: 240, A: 255}, // 浅灰色背景
		TextColor:  color.RGBA{R: 0, G: 0, B: 0, A: 255},       // 黑色文字
		NoiseColor: color.RGBA{R: 128, G: 128, B: 128, A: 128}, // 半透明灰色噪点
	}
}

// GenerateCaptcha 生成验证码图片
// text: 要显示的验证码文本
// config: 可选配置，传nil使用默认配置
// 返回: base64编码的PNG图片数据
func GenerateCaptcha(text string, config *CaptchaConfig) (string, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 创建图片
	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))

	// 填充背景色
	draw.Draw(img, img.Bounds(), &image.Uniform{config.BgColor}, image.Point{}, draw.Src)

	// 添加噪点
	addNoise(img, config)

	// 添加干扰线
	addLines(img, config)

	// 绘制文字
	err := drawText(img, text, config)
	if err != nil {
		return "", fmt.Errorf("绘制文字失败: %v", err)
	}

	// 转换为base64
	return imageToBase64(img)
}

// GenerateRandomText 生成随机验证码文本
// length: 验证码长度
// charType: 字符类型 "number"(数字) "letter"(字母) "mixed"(数字+字母)
func GenerateRandomText(length int, charType string) string {
	var chars string
	switch strings.ToLower(charType) {
	case "number":
		chars = "0123456789"
	case "letter":
		chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case "mixed":
		chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	default:
		chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[randSource.Intn(len(chars))]
	}

	return string(result)
}

// QuickGenerate 快速生成验证码
// 使用默认配置生成4位数字验证码
func QuickGenerate() (text string, imageData string, err error) {
	text = GenerateRandomText(4, "mixed")
	imageData, err = GenerateCaptcha(text, nil)
	return
}

// addNoise 添加噪点
func addNoise(img *image.RGBA, config *CaptchaConfig) {
	bounds := img.Bounds()

	for i := 0; i < config.NoiseCount; i++ {
		x := randSource.Intn(bounds.Max.X)
		y := randSource.Intn(bounds.Max.Y)

		// 随机噪点颜色变化
		r := config.NoiseColor.R + uint8(randSource.Intn(50)-25)
		g := config.NoiseColor.G + uint8(randSource.Intn(50)-25)
		b := config.NoiseColor.B + uint8(randSource.Intn(50)-25)

		img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: config.NoiseColor.A})
	}
}

// addLines 添加干扰线
func addLines(img *image.RGBA, config *CaptchaConfig) {
	bounds := img.Bounds()

	// 添加3-5条随机线条
	lineCount := 3 + randSource.Intn(3)
	for i := 0; i < lineCount; i++ {
		x1 := randSource.Intn(bounds.Max.X)
		y1 := randSource.Intn(bounds.Max.Y)
		x2 := randSource.Intn(bounds.Max.X)
		y2 := randSource.Intn(bounds.Max.Y)

		drawLine(img, x1, y1, x2, y2, config.NoiseColor)
	}
}

// drawLine 绘制线条
func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	sy := 1

	if x1 >= x2 {
		sx = -1
	}
	if y1 >= y2 {
		sy = -1
	}

	err := dx - dy
	x, y := x1, y1

	for {
		img.Set(x, y, c)

		if x == x2 && y == y2 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// abs 求绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// drawText 绘制文字
func drawText(img *image.RGBA, text string, config *CaptchaConfig) error {
	bounds := img.Bounds()

	// 计算字符间距
	charWidth := bounds.Max.X / len(text)
	startX := charWidth / 4

	// 使用基础字体
	face := basicfont.Face7x13

	for i, char := range text {
		// 计算字符位置，添加一些随机偏移
		x := startX + i*charWidth + randSource.Intn(5)
		y := bounds.Max.Y/2 + randSource.Intn(10) - 5

		// 绘制字符
		point := fixed.Point26_6{
			X: fixed.Int26_6(x * 64),
			Y: fixed.Int26_6(y * 64),
		}

		d := &font.Drawer{
			Dst:  img,
			Src:  &image.Uniform{config.TextColor},
			Face: face,
			Dot:  point,
		}

		d.DrawString(string(char))
	}

	return nil
}

// imageToBase64 将图片转换为base64编码
func imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer

	err := png.Encode(&buf, img)
	if err != nil {
		return "", fmt.Errorf("PNG编码失败: %v", err)
	}

	// 返回完整的data URL格式
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/png;base64," + encoded, nil
}
