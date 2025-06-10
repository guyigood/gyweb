package scaffold

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Scaffold struct {
	TemplatePath string
	Replacements map[string]string
}

func NewScaffold(templatePath string) *Scaffold {
	return &Scaffold{
		TemplatePath: templatePath,
		Replacements: make(map[string]string),
	}
}

func (s *Scaffold) AddReplacement(placeholder, value string) {
	s.Replacements[placeholder] = value
}

func (s *Scaffold) CreateProject(projectName string) error {
	if _, err := os.Stat(projectName); !os.IsNotExist(err) {
		return fmt.Errorf("项目目录 '%s' 已存在", projectName)
	}

	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("创建项目目录失败: %v", err)
	}

	s.AddReplacement("{firstweb}", projectName)

	fmt.Printf("正在创建项目 '%s'...\n", projectName)
	if err := s.copyTemplate(s.TemplatePath, projectName); err != nil {
		os.RemoveAll(projectName)
		return fmt.Errorf("拷贝模板失败: %v", err)
	}

	fmt.Printf("项目 '%s' 创建成功！\n", projectName)
	return nil
}

func (s *Scaffold) copyTemplate(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, d.Type())
		}

		return s.copyFile(path, dstPath)
	})
}

func (s *Scaffold) copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if s.shouldReplaceContent(src) {
		return s.copyWithReplacement(srcFile, dstFile)
	} else {
		_, err = io.Copy(dstFile, srcFile)
		return err
	}
}

func (s *Scaffold) shouldReplaceContent(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	textExtensions := []string{
		".go", ".mod", ".sum", ".txt", ".md", ".json", ".yaml", ".yml",
		".html", ".css", ".js", ".sql", ".xml", ".toml", ".ini",
	}

	for _, textExt := range textExtensions {
		if ext == textExt {
			return true
		}
	}
	return false
}

func (s *Scaffold) copyWithReplacement(src io.Reader, dst io.Writer) error {
	content, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	contentStr := string(content)
	for placeholder, replacement := range s.Replacements {
		contentStr = strings.ReplaceAll(contentStr, placeholder, replacement)
	}

	_, err = dst.Write([]byte(contentStr))
	return err
}

func CreateProjectFromTemplate(templatePath, projectName string) error {
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("模板目录不存在: %s", templatePath)
	}

	scaffold := NewScaffold(templatePath)
	return scaffold.CreateProject(projectName)
}
