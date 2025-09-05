package common

import (
	"io"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/guyigood/gyweb/core/gyarn"
)

// UploadFile 上传文件,参数filename为上传的文件名,参数uploadPath为上传的文件路径
func UploadFile(c *gyarn.Context, filename string, uploadPath string) (string, error) {
	// 获取上传的文件
	file, err := c.FormFile(filename)
	if err != nil {
		return "", err
	}
	//默认/upload/  同时检查是否创建了目录
	if uploadPath == "" {
		uploadPath = "upload/"
	}
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		os.MkdirAll(uploadPath, 0755)
	}
	// TODO: 处理文件保存逻辑，按时间生成日期目录
	date := time.Now().Format("2006-01-02")
	uploadPath += date + "/"
	// 创建目录
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		os.MkdirAll(uploadPath, 0755)
	}
	// 保存文件
	return SaveFileSecure(file, uploadPath)

}

func SaveFileSecure(fh *multipart.FileHeader, uploadDir string) (string, error) {
	// 清洗文件名和路径，文件名使用时间戳+随机数+文件扩展名
	filename := time.Now().Format("20060102150405") + strconv.Itoa(rand.Intn(1000000)) + filepath.Ext(fh.Filename)
	dstPath := filepath.Join(uploadDir, filename)

	// 验证目录
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// 打开文件流
	srcFile, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	// 复制内容
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		os.Remove(dstPath) // 失败时清理残留
		return "", err
	}

	return dstPath, nil
}
