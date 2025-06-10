# GyWeb 项目构建文件

.PHONY: build scaffold clean install help

# 默认目标
help:
	@echo "GyWeb 构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  build     构建主程序"
	@echo "  scaffold  构建脚手架工具"
	@echo "  install   安装脚手架工具到系统"
	@echo "  clean     清理构建文件"
	@echo "  test      运行测试"
	@echo ""

# 构建主程序
build:
	@echo "构建主程序..."
	@if not exist bin mkdir bin
	go build -o bin/gyweb main.go

# 构建脚手架工具
scaffold:
	@echo "构建脚手架工具..."
	@if not exist bin mkdir bin
	go build -o bin/gyweb-scaffold cmd/scaffold/main.go

# 构建脚手架工具 (Windows版本)
scaffold-win:
	@echo "构建脚手架工具 (Windows)..."
	@if not exist bin mkdir bin
	go build -o bin/gyweb-scaffold.exe cmd/scaffold/main.go

# 安装脚手架工具到系统 (需要管理员权限)
install: scaffold
	@echo "安装脚手架工具到系统..."
	@if exist bin/gyweb-scaffold.exe (copy bin\gyweb-scaffold.exe C:\Windows\System32\gyweb-scaffold.exe) else (cp bin/gyweb-scaffold /usr/local/bin/gyweb-scaffold)
	@echo "安装完成! 现在可以在任何地方使用 'gyweb-scaffold create projectname'"

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@if exist bin rmdir /s /q bin
	go clean

# 运行测试
test:
	@echo "运行测试..."
	go test ./...

# 创建示例项目 (用于测试)
demo:
	@echo "创建示例项目..."
	@if exist bin\gyweb-scaffold.exe (bin\gyweb-scaffold.exe create demo-project) else (./bin/gyweb-scaffold create demo-project)

# 构建所有
all: build scaffold 