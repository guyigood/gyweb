package main

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/guyigood/gyweb/core/utils/captcha"
)

func main() {
	fmt.Println("=== gyweb 验证码生成工具示例 ===\n")

	// 示例1: 使用自定义文本生成验证码
	fmt.Println("1. 使用自定义文本生成验证码:")
	imageData1, err := captcha.GenerateCaptcha("AB12", nil)
	if err != nil {
		log.Fatal("生成验证码失败:", err)
	}
	fmt.Printf("   文本: AB12\n")
	fmt.Printf("   图片大小: %d 字符\n", len(imageData1))
	fmt.Printf("   格式: %s\n\n", imageData1[:30]+"...")

	// 示例2: 快速生成随机验证码
	fmt.Println("2. 快速生成随机验证码:")
	text2, imageData2, err := captcha.QuickGenerate()
	if err != nil {
		log.Fatal("快速生成验证码失败:", err)
	}
	fmt.Printf("   随机文本: %s\n", text2)
	fmt.Printf("   图片大小: %d 字符\n\n", len(imageData2))

	// 示例3: 生成不同类型的随机文本
	fmt.Println("3. 生成不同类型的随机文本:")
	numberCode := captcha.GenerateRandomText(4, "number")
	letterCode := captcha.GenerateRandomText(6, "letter")
	mixedCode := captcha.GenerateRandomText(5, "mixed")
	fmt.Printf("   4位数字: %s\n", numberCode)
	fmt.Printf("   6位字母: %s\n", letterCode)
	fmt.Printf("   5位混合: %s\n\n", mixedCode)

	// 示例4: 使用自定义配置
	fmt.Println("4. 使用自定义配置生成验证码:")
	config := &captcha.CaptchaConfig{
		Width:      150,
		Height:     60,
		NoiseCount: 80,
		NoiseLevel: 0.5,
		FontSize:   28,
		BgColor:    color.RGBA{R: 255, G: 255, B: 255, A: 255}, // 白色背景
		TextColor:  color.RGBA{R: 255, G: 0, B: 0, A: 255},     // 红色文字
		NoiseColor: color.RGBA{R: 200, G: 200, B: 200, A: 100}, // 浅灰噪点
	}
	imageData4, err := captcha.GenerateCaptcha("HELLO", config)
	if err != nil {
		log.Fatal("自定义配置生成验证码失败:", err)
	}
	fmt.Printf("   文本: HELLO\n")
	fmt.Printf("   自定义尺寸: 150x60\n")
	fmt.Printf("   图片大小: %d 字符\n\n", len(imageData4))

	// 示例5: 保存验证码到文件 (可选)
	fmt.Println("5. 保存验证码到文件:")
	if saveToFile(imageData1, "captcha_AB12.png") {
		fmt.Println("   ✓ 已保存 captcha_AB12.png")
	}
	if saveToFile(imageData2, fmt.Sprintf("captcha_%s.png", text2)) {
		fmt.Printf("   ✓ 已保存 captcha_%s.png\n", text2)
	}
	if saveToFile(imageData4, "captcha_HELLO.png") {
		fmt.Println("   ✓ 已保存 captcha_HELLO.png")
	}

	fmt.Println("\n=== 使用提示 ===")
	fmt.Println("1. 生成的图片数据是 data URL 格式，可直接在 HTML 中使用")
	fmt.Println("2. 建议将验证码文本存储在 session 中")
	fmt.Println("3. 验证时忽略大小写，使用 strings.ToUpper() 进行比较")
	fmt.Println("4. 为安全考虑，验证码应设置过期时间")
	fmt.Println("\n示例运行完成!")
}

// saveToFile 将 base64 图片数据保存到文件
func saveToFile(dataURL, filename string) bool {
	// 解析 data URL
	if !strings.HasPrefix(dataURL, "data:image/png;base64,") {
		fmt.Printf("   ✗ 无效的数据格式: %s\n", filename)
		return false
	}

	// 提取 base64 数据
	base64Data := strings.TrimPrefix(dataURL, "data:image/png;base64,")

	// 解码 base64
	imageBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		fmt.Printf("   ✗ 解码失败 %s: %v\n", filename, err)
		return false
	}

	// 确保目录存在
	dir := filepath.Dir(filename)
	if dir != "." {
		os.MkdirAll(dir, 0755)
	}

	// 写入文件
	err = os.WriteFile(filename, imageBytes, 0644)
	if err != nil {
		fmt.Printf("   ✗ 保存失败 %s: %v\n", filename, err)
		return false
	}

	return true
}
