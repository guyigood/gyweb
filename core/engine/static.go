package engine

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/guyigood/gyweb/core/gyarn"
)

// StaticConfig 静态文件服务配置
// StaticConfig 用于配置静态文件服务的各种参数
// 包括文件目录、URL前缀、是否允许目录浏览等选项
type StaticConfig struct {
	// Root 静态文件根目录路径
	Root string
	// Prefix URL路径前缀，默认为 "/static"
	Prefix string
	// IndexFile 默认索引文件名，如 "index.html"
	IndexFile string
	// Browse 是否允许目录浏览
	Browse bool
	// MaxAge 缓存时间（秒）
	MaxAge int
	// Compress 是否启用压缩
	Compress bool
}

// DefaultStaticConfig 返回默认的静态文件配置
// 提供了一套合理的默认配置，适用于大多数场景
func DefaultStaticConfig() *StaticConfig {
	return &StaticConfig{
		Root:      "./static",
		Prefix:    "/static",
		IndexFile: "index.html",
		Browse:    false,
		MaxAge:    86400, // 24小时
		Compress:  true,
	}
}

// Static 创建静态文件服务中间件
// 返回一个HandlerFunc，用于处理静态文件请求
// 支持自定义配置，包括文件路径、缓存策略等
func Static(config *StaticConfig) gyarn.HandlerFunc {
	if config == nil {
		config = DefaultStaticConfig()
	}

	// 规范化根目录路径
	absRoot, err := filepath.Abs(config.Root)
	if err != nil {
		absRoot = config.Root
	}

	return func(c *gyarn.Context) {
		// 检查是否为静态文件请求
		if !strings.HasPrefix(c.Path, config.Prefix) {
			c.Next()
			return
		}

		// 获取相对路径
		relativePath := strings.TrimPrefix(c.Path, config.Prefix)
		if relativePath == "" {
			relativePath = "/"
		}

		// 构建完整文件路径
		filePath := filepath.Join(absRoot, filepath.Clean(relativePath))

		// 安全检查：确保文件在根目录内
		if !strings.HasPrefix(filePath, absRoot) {
			c.String(http.StatusForbidden, "403 Forbidden")
			return
		}

		// 设置缓存头
		if config.MaxAge > 0 {
			c.SetHeader("Cache-Control", fmt.Sprintf("public, max-age=%d", config.MaxAge))
		}

		// 设置压缩头
		if config.Compress {
			c.SetHeader("Vary", "Accept-Encoding")
		}

		// 尝试打开文件
		fileInfo, err := http.Dir(absRoot).Open(relativePath)
		if err != nil {
			c.String(http.StatusNotFound, "404 Not Found")
			return
		}
		defer fileInfo.Close()

		// 获取文件信息
		stat, err := fileInfo.Stat()
		if err != nil {
			c.String(http.StatusNotFound, "404 Not Found")
			return
		}

		// 如果是目录
		if stat.IsDir() {
			// 检查是否有索引文件
			if config.IndexFile != "" {
				indexPath := filepath.Join(relativePath, config.IndexFile)
				indexFile, err := http.Dir(absRoot).Open(indexPath)
				if err == nil {
					indexFile.Close()
					// 重定向到索引文件
					if strings.HasSuffix(c.Path, "/") {
						c.Path = c.Path + config.IndexFile
					} else {
						c.Path = c.Path + "/" + config.IndexFile
					}
					// 递归调用处理索引文件
					Static(config)(c)
					return
				}
			}

			// 如果允许目录浏览
			if config.Browse {
				dirListHandler(absRoot, relativePath)(c)
				return
			}

			c.String(http.StatusForbidden, "403 Forbidden")
			return
		}

		// 设置内容类型
		contentType := getContentType(filepath.Ext(filePath))
		c.SetHeader("Content-Type", contentType)

		// 使用http.ServeFile提供文件服务
		http.ServeFile(c.Writer, c.Request, filePath)
		c.Abort()
	}
}

// StaticFS 创建基于http.FileSystem的静态文件服务
// 提供更大的灵活性，可以使用嵌入的文件系统或其他自定义文件系统
func StaticFS(fs http.FileSystem, prefix string) gyarn.HandlerFunc {
	return func(c *gyarn.Context) {
		if !strings.HasPrefix(c.Path, prefix) {
			c.Next()
			return
		}

		relativePath := strings.TrimPrefix(c.Path, prefix)
		if relativePath == "" {
			relativePath = "/"
		}

		file, err := fs.Open(relativePath)
		if err != nil {
			c.String(http.StatusNotFound, "404 Not Found")
			return
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			c.String(http.StatusNotFound, "404 Not Found")
			return
		}

		if stat.IsDir() {
			c.String(http.StatusForbidden, "403 Forbidden")
			return
		}

		http.ServeFile(c.Writer, c.Request, relativePath)
		c.Abort()
	}
}

// dirListHandler 目录浏览处理器
// 生成目录列表的HTML页面
func dirListHandler(root, dirPath string) gyarn.HandlerFunc {
	return func(c *gyarn.Context) {
		dir := http.Dir(root)
		file, err := dir.Open(dirPath)
		if err != nil {
			c.String(http.StatusNotFound, "404 Not Found")
			return
		}
		defer file.Close()

		files, err := file.Readdir(-1)
		if err != nil {
			c.String(http.StatusInternalServerError, "500 Internal Server Error")
			return
		}

		c.SetHeader("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusOK)

		// 生成目录列表HTML
		html := generateDirListHTML(dirPath, files)
		c.Writer.Write([]byte(html))
	}
}

// generateDirListHTML 生成目录列表HTML
// 创建美观的目录浏览页面
func generateDirListHTML(dirPath string, files []os.FileInfo) string {
	var html strings.Builder
	html.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>目录列表 - ` + dirPath + `</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 10px; border-radius: 5px; margin-bottom: 20px; }
        .file-list { border-collapse: collapse; width: 100%; }
        .file-list th, .file-list td { padding: 8px; text-align: left; border-bottom: 1px solid #ddd; }
        .file-list th { background-color: #f2f2f2; }
        .file-list tr:hover { background-color: #f5f5f5; }
        .icon { margin-right: 5px; }
        a { text-decoration: none; color: #0066cc; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="header">
        <h1>目录列表: ` + dirPath + `</h1>
        <p>文件数量: ` + fmt.Sprintf("%d", len(files)) + `</p>
    </div>
    <table class="file-list">
        <thead>
            <tr>
                <th>名称</th>
                <th>大小</th>
                <th>修改时间</th>
                <th>类型</th>
            </tr>
        </thead>
        <tbody>`)

	for _, file := range files {
		name := file.Name()
		size := formatFileSize(file.Size())
		modTime := file.ModTime().Format("2006-01-02 15:04:05")
		fileType := "文件"
		icon := "📄"
		link := name

		if file.IsDir() {
			fileType = "目录"
			icon = "📁"
			link = name + "/"
		}

		html.WriteString(fmt.Sprintf(`
            <tr>
                <td><a href="%s">%s %s</a></td>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
            </tr>`, link, icon, name, size, modTime, fileType))
	}

	html.WriteString(`
        </tbody>
    </table>
</body>
</html>`)

	return html.String()
}

// formatFileSize 格式化文件大小显示
// 将字节转换为人类可读格式
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// getContentType 根据文件扩展名获取内容类型
// 支持常见的文件类型映射
func getContentType(ext string) string {
	mimeTypes := map[string]string{
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".mp3":  "audio/mpeg",
		".mp4":  "video/mp4",
		".txt":  "text/plain",
		".ico":  "image/x-icon",
	}

	contentType, exists := mimeTypes[strings.ToLower(ext)]
	if !exists {
		contentType = "application/octet-stream"
	}
	return contentType
}
