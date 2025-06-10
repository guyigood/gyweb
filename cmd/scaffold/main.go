package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/guyigood/gyweb/core/scaffold"
)

func main() {
	if len(os.Args) < 3 {
		showUsage()
		return
	}

	command := os.Args[1]
	projectName := os.Args[2]

	switch command {
	case "create":
		if err := createProject(projectName); err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("未知命令: %s\n", command)
		showUsage()
		os.Exit(1)
	}
}

func createProject(projectName string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取执行路径失败: %v", err)
	}

	baseDir := filepath.Dir(execPath)
	templatePath := filepath.Join(baseDir, "template")

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		templatePath = filepath.Join(baseDir, "..", "template")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			templatePath = "template"
		}
	}

	fmt.Printf("使用模板目录: %s\n", templatePath)
	return scaffold.CreateProjectFromTemplate(templatePath, projectName)
}

func showUsage() {
	fmt.Println("GyWeb 脚手架工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  scaffold create <项目名称>    创建新项目")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  scaffold create myproject    创建名为 'myproject' 的新项目")
	fmt.Println()
	fmt.Println("说明:")
	fmt.Println("  - 会从 template 目录拷贝所有文件到新项目目录")
	fmt.Println("  - 自动将代码中的 {firstweb} 替换为项目名称")
	fmt.Println("  - 支持 .go .mod .json .html 等文本文件的内容替换")
}
