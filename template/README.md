# {project_name}    

基于 GyWeb 框架创建的Web应用程序

## 项目简介

{project_name} 是使用 GyWeb 脚手架工具创建的Go Web应用程序。

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 运行应用

```bash
go run main.go
```

应用将在 `http://localhost:8080` 启动

### API 端点

- `GET /` - 首页
- `GET /api/hello` - Hello API

## 项目结构

```
{project_name}/
├── main.go        # 主程序入口
├── go.mod         # Go模块文件  
├── README.md      # 项目说明
└── ...
```

## 开发指南

### 添加新路由

在 `main.go` 中添加新的路由处理器：

```go
r.GET("/api/new", func(c *engine.Context) {
    c.JSON(200, map[string]string{
        "message": "New API endpoint",
    })
})
```

### 使用路由组

```go
api := r.Group("/api/v1")
{
    api.GET("/users", getUsersHandler)
    api.POST("/users", createUserHandler)
}
```

## 部署

构建生产版本：

```bash
go build -o {project_name} main.go
```

运行：

```bash
./{project_name}
```

## 许可证

MIT License 