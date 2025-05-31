package main

import (
	"fmt"
	"log"
	"time"

	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
	"github.com/guyigood/gyweb/core/services/excel"
	"github.com/guyigood/gyweb/core/services/wechat"
)

// 用户数据结构示例
type User struct {
	ID       int       `json:"id" excel:"ID"`
	Name     string    `json:"name" excel:"姓名"`
	Email    string    `json:"email" excel:"邮箱"`
	Age      int       `json:"age" excel:"年龄"`
	IsActive bool      `json:"is_active" excel:"是否激活"`
	CreateAt time.Time `json:"create_at" excel:"创建时间"`
}

// 附加服务示例
func mainAdditionalServices() {
	r := engine.New()

	// 使用中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 初始化附加服务
	setupAdditionalServices(r)

	log.Println("🚀 附加服务集成示例启动...")
	log.Println("📋 API 端点:")
	log.Println("   微信公众号:")
	log.Println("   - GET  /api/wechat/verify              验证微信服务器")
	log.Println("   - POST /api/wechat/verify              处理微信消息")
	log.Println("   - GET  /api/wechat/user/:openid        获取用户信息")
	log.Println("   - POST /api/wechat/message/template    发送模板消息")
	log.Println("   - POST /api/wechat/message/text        发送文本消息")
	log.Println("   - POST /api/wechat/qrcode              生成二维码")
	log.Println("   - POST /api/wechat/menu                创建菜单")
	log.Println("   - GET  /api/wechat/oauth               网页授权")
	log.Println()
	log.Println("   Excel 操作:")
	log.Println("   - POST /api/excel/import               导入Excel数据")
	log.Println("   - GET  /api/excel/export               导出Excel数据")
	log.Println("   - POST /api/excel/template             生成Excel模板")
	log.Println()

	if err := r.Run(":8080"); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}

// 设置附加服务
func setupAdditionalServices(r *engine.Engine) {
	// 微信公众号配置
	wechatConfig := &wechat.WechatConfig{
		AppID:          "your_wechat_app_id",
		AppSecret:      "your_wechat_app_secret",
		Token:          "your_wechat_token",
		EncodingAESKey: "your_encoding_aes_key",
	}
	wechatClient := wechat.NewWechat(wechatConfig)

	// 设置路由
	setupWechatRoutes(r, wechatClient)
	setupExcelRoutes(r)
}

// 设置微信公众号路由
func setupWechatRoutes(r *engine.Engine, wechatClient *wechat.Wechat) {
	wechatGroup := r.Group("/api/wechat")

	// 验证微信服务器
	wechatGroup.GET("/verify", func(c *gyarn.Context) {
		signature := c.Query("signature")
		timestamp := c.Query("timestamp")
		nonce := c.Query("nonce")
		echostr := c.Query("echostr")

		if wechatClient.VerifySignature(signature, timestamp, nonce) {
			c.String(200, echostr)
		} else {
			c.String(403, "验证失败")
		}
	})

	// 处理微信消息
	wechatGroup.POST("/verify", func(c *gyarn.Context) {
		// 这里处理微信推送的消息
		// 实际项目中需要解析XML消息并处理不同类型的事件
		c.String(200, "success")
	})

	// 获取用户信息
	wechatGroup.GET("/user/:openid", func(c *gyarn.Context) {
		openID := c.Param("openid")
		if openID == "" {
			c.BadRequest("OpenID不能为空")
			return
		}

		userInfo, err := wechatClient.GetUserInfo(openID)
		if err != nil {
			c.InternalServerError("获取用户信息失败: " + err.Error())
			return
		}

		c.Success(userInfo)
	})

	// 发送模板消息
	wechatGroup.POST("/message/template", func(c *gyarn.Context) {
		var req struct {
			ToUser     string                 `json:"to_user"`
			TemplateID string                 `json:"template_id"`
			URL        string                 `json:"url"`
			Data       map[string]interface{} `json:"data"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		msg := &wechat.TemplateMessage{
			ToUser:     req.ToUser,
			TemplateID: req.TemplateID,
			URL:        req.URL,
			Data:       req.Data,
		}

		resp, err := wechatClient.SendTemplateMessage(msg)
		if err != nil {
			c.InternalServerError("发送模板消息失败: " + err.Error())
			return
		}

		c.Success(gyarn.H{
			"errcode": resp.ErrCode,
			"errmsg":  resp.ErrMsg,
			"msgid":   resp.MsgID,
		})
	})

	// 发送文本消息
	wechatGroup.POST("/message/text", func(c *gyarn.Context) {
		var req struct {
			OpenID  string `json:"openid"`
			Content string `json:"content"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		resp, err := wechatClient.SendTextMessage(req.OpenID, req.Content)
		if err != nil {
			c.InternalServerError("发送文本消息失败: " + err.Error())
			return
		}

		c.Success(gyarn.H{
			"errcode": resp.ErrCode,
			"errmsg":  resp.ErrMsg,
		})
	})

	// 生成二维码
	wechatGroup.POST("/qrcode", func(c *gyarn.Context) {
		var req struct {
			ActionName    string `json:"action_name"`    // QR_SCENE, QR_STR_SCENE, QR_LIMIT_SCENE, QR_LIMIT_STR_SCENE
			SceneID       int    `json:"scene_id"`       // 场景值ID
			SceneStr      string `json:"scene_str"`      // 场景值字符串
			ExpireSeconds int    `json:"expire_seconds"` // 过期时间（秒）
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		qrReq := &wechat.QRCodeRequest{
			ActionName:    req.ActionName,
			ExpireSeconds: req.ExpireSeconds,
			ActionInfo: &wechat.ActionInfo{
				Scene: &wechat.Scene{
					SceneID:  req.SceneID,
					SceneStr: req.SceneStr,
				},
			},
		}

		qrResp, err := wechatClient.CreateQRCode(qrReq)
		if err != nil {
			c.InternalServerError("生成二维码失败: " + err.Error())
			return
		}

		// 获取二维码图片
		qrImage, err := wechatClient.GetQRCodeImage(qrResp.Ticket)
		if err != nil {
			c.InternalServerError("获取二维码图片失败: " + err.Error())
			return
		}

		c.Success(gyarn.H{
			"ticket":         qrResp.Ticket,
			"url":            qrResp.URL,
			"expire_seconds": qrResp.ExpireSeconds,
			"qr_image":       qrImage,
		})
	})

	// 创建菜单
	wechatGroup.POST("/menu", func(c *gyarn.Context) {
		var req struct {
			Buttons []wechat.Button `json:"buttons"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		menu := &wechat.Menu{
			Button: req.Buttons,
		}

		resp, err := wechatClient.CreateMenu(menu)
		if err != nil {
			c.InternalServerError("创建菜单失败: " + err.Error())
			return
		}

		c.Success(gyarn.H{
			"errcode": resp.ErrCode,
			"errmsg":  resp.ErrMsg,
		})
	})

	// 网页授权
	wechatGroup.GET("/oauth", func(c *gyarn.Context) {
		code := c.Query("code")
		state := c.Query("state")

		if code == "" {
			// 第一步：用户同意授权，获取code
			redirectURI := "https://your-domain.com/api/wechat/oauth"
			scope := "snsapi_userinfo"
			oauthURL := wechatClient.GetOAuthURL(redirectURI, state, scope)

			c.Success(gyarn.H{
				"oauth_url": oauthURL,
			})
			return
		}

		// 第二步：通过code换取网页授权access_token
		oauthToken, err := wechatClient.GetOAuthAccessToken(code)
		if err != nil {
			c.InternalServerError("获取授权token失败: " + err.Error())
			return
		}

		// 第三步：拉取用户信息
		oauthUser, err := wechatClient.GetOAuthUserInfo(oauthToken.AccessToken, oauthToken.OpenID)
		if err != nil {
			c.InternalServerError("获取用户信息失败: " + err.Error())
			return
		}

		c.Success(gyarn.H{
			"access_token": oauthToken.AccessToken,
			"openid":       oauthToken.OpenID,
			"user_info":    oauthUser,
		})
	})
}

// 设置Excel路由
func setupExcelRoutes(r *engine.Engine) {
	excelGroup := r.Group("/api/excel")

	// 导入Excel数据
	excelGroup.POST("/import", func(c *gyarn.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.BadRequest("请选择Excel文件")
			return
		}

		// 打开上传的文件
		src, err := file.Open()
		if err != nil {
			c.InternalServerError("打开文件失败")
			return
		}
		defer src.Close()

		// 读取文件内容
		fileData := make([]byte, file.Size)
		if _, err := src.Read(fileData); err != nil {
			c.InternalServerError("读取文件失败")
			return
		}

		// 创建Excel服务
		excelService, err := excel.NewExcelServiceWithReader(fileData)
		if err != nil {
			c.InternalServerError("解析Excel文件失败: " + err.Error())
			return
		}
		defer excelService.Close()

		// 配置导入选项
		importOptions := &excel.ImportOptions{
			StartRow:  2,
			HeaderRow: 1,
			ColumnMaps: []excel.ColumnMap{
				{Name: "ID", Field: "ID", Required: true, DataType: "int"},
				{Name: "姓名", Field: "Name", Required: true, DataType: "string"},
				{Name: "邮箱", Field: "Email", Required: false, DataType: "string"},
				{Name: "年龄", Field: "Age", Required: false, DataType: "int"},
				{Name: "是否激活", Field: "IsActive", Required: false, DataType: "bool"},
				{Name: "创建时间", Field: "CreateAt", Required: false, DataType: "time", Format: "2006-01-02"},
			},
			ValidateFunc: func(data interface{}) error {
				user := data.(*User)
				if user.Age < 0 || user.Age > 150 {
					return fmt.Errorf("年龄必须在0-150之间")
				}
				return nil
			},
		}

		// 导入数据
		var users []User
		result, err := excelService.ImportData(importOptions, &users)
		if err != nil {
			c.InternalServerError("导入数据失败: " + err.Error())
			return
		}

		c.Success(gyarn.H{
			"success_count": result.SuccessCount,
			"error_count":   result.ErrorCount,
			"errors":        result.Errors,
			"users":         users,
		})
	})

	// 导出Excel数据
	excelGroup.GET("/export", func(c *gyarn.Context) {
		// 模拟用户数据
		users := []User{
			{
				ID:       1,
				Name:     "张三",
				Email:    "zhangsan@example.com",
				Age:      25,
				IsActive: true,
				CreateAt: time.Now(),
			},
			{
				ID:       2,
				Name:     "李四",
				Email:    "lisi@example.com",
				Age:      30,
				IsActive: false,
				CreateAt: time.Now().AddDate(0, -1, 0),
			},
		}

		// 创建Excel服务
		excelService := excel.NewExcelService()
		defer excelService.Close()

		// 配置导出选项
		exportOptions := &excel.ExportOptions{
			SheetName: "用户列表",
			Headers:   []string{"ID", "姓名", "邮箱", "年龄", "是否激活", "创建时间"},
			ColumnMaps: []excel.ColumnMap{
				{Name: "ID", Field: "ID", DataType: "int"},
				{Name: "姓名", Field: "Name", DataType: "string"},
				{Name: "邮箱", Field: "Email", DataType: "string"},
				{Name: "年龄", Field: "Age", DataType: "int"},
				{Name: "是否激活", Field: "IsActive", DataType: "bool"},
				{Name: "创建时间", Field: "CreateAt", DataType: "time", Format: "2006-01-02 15:04:05"},
			},
			StyleConfig: &excel.StyleConfig{
				HeaderStyle: &excel.CellStyle{
					Font: &excel.FontStyle{
						Bold:  true,
						Size:  12,
						Color: "#FFFFFF",
					},
					Fill: &excel.FillStyle{
						Type:    "pattern",
						Pattern: 1,
						Color:   "#4472C4",
					},
					Alignment: &excel.AlignmentStyle{
						Horizontal: "center",
						Vertical:   "center",
					},
				},
				DataStyle: &excel.CellStyle{
					Alignment: &excel.AlignmentStyle{
						Horizontal: "left",
						Vertical:   "center",
					},
					Border: &excel.BorderStyle{
						Type:  "thin",
						Color: "#000000",
					},
				},
				ColumnWidths: []float64{10, 15, 25, 10, 12, 20},
			},
		}

		// 导出数据
		if err := excelService.ExportData(users, exportOptions); err != nil {
			c.InternalServerError("导出数据失败: " + err.Error())
			return
		}

		// 获取文件字节数据
		fileData, err := excelService.GetBytes()
		if err != nil {
			c.InternalServerError("生成Excel文件失败: " + err.Error())
			return
		}

		// 设置响应头
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=users.xlsx")
		c.Header("Content-Length", fmt.Sprintf("%d", len(fileData)))

		// 返回文件数据
		c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)
	})

	// 生成Excel模板
	excelGroup.POST("/template", func(c *gyarn.Context) {
		var req struct {
			SheetName string   `json:"sheet_name"`
			Headers   []string `json:"headers"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		// 创建Excel服务
		excelService := excel.NewExcelService()
		defer excelService.Close()

		sheetName := req.SheetName
		if sheetName == "" {
			sheetName = "模板"
		}

		// 添加工作表
		if err := excelService.AddSheet(sheetName); err != nil {
			c.InternalServerError("创建工作表失败: " + err.Error())
			return
		}

		// 设置表头
		for i, header := range req.Headers {
			cell := fmt.Sprintf("%s1", getColumnName(i+1))
			if err := excelService.SetCellValue(sheetName, cell, header); err != nil {
				c.InternalServerError("设置表头失败: " + err.Error())
				return
			}
		}

		// 获取文件字节数据
		fileData, err := excelService.GetBytes()
		if err != nil {
			c.InternalServerError("生成Excel模板失败: " + err.Error())
			return
		}

		// 设置响应头
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=template.xlsx")
		c.Header("Content-Length", fmt.Sprintf("%d", len(fileData)))

		// 返回文件数据
		c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)
	})
}

// getColumnName 获取列名（A, B, C, ..., AA, AB, ...）
func getColumnName(column int) string {
	name := ""
	for column > 0 {
		column--
		name = string(rune('A'+(column%26))) + name
		column /= 26
	}
	return name
}
