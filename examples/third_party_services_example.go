package main

import (
	"log"

	"github.com/guyigood/gyweb/core/engine"
	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
	"github.com/guyigood/gyweb/core/services/dingtalk"
	"github.com/guyigood/gyweb/core/services/miniprogram"
	"github.com/guyigood/gyweb/core/services/payment"
)

// 第三方服务集成示例
func mainThirdPartyServices() {
	r := engine.New()

	// 使用中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 初始化第三方服务客户端
	setupThirdPartyServices(r)

	log.Println("🚀 第三方服务集成示例启动...")
	log.Println("📋 API 端点:")
	log.Println("   支付相关:")
	log.Println("   - POST /api/pay/wechat/create          创建微信支付订单")
	log.Println("   - POST /api/pay/wechat/query           查询微信支付订单")
	log.Println("   - POST /api/pay/wechat/refund          微信支付退款")
	log.Println("   - POST /api/pay/wechat/notify          微信支付回调")
	log.Println("   - POST /api/pay/alipay/create          创建支付宝订单")
	log.Println("   - POST /api/pay/alipay/query           查询支付宝订单")
	log.Println("   - POST /api/pay/alipay/refund          支付宝退款")
	log.Println()
	log.Println("   微信小程序:")
	log.Println("   - POST /api/miniprogram/login          小程序登录")
	log.Println("   - POST /api/miniprogram/userinfo       获取用户信息")
	log.Println("   - POST /api/miniprogram/phone          获取手机号")
	log.Println("   - POST /api/miniprogram/qrcode         生成小程序码")
	log.Println("   - POST /api/miniprogram/template       发送模板消息")
	log.Println()
	log.Println("   钉钉集成:")
	log.Println("   - GET  /api/dingtalk/user/:id          获取用户信息")
	log.Println("   - POST /api/dingtalk/user/mobile       根据手机号获取用户")
	log.Println("   - POST /api/dingtalk/message/work      发送工作通知")
	log.Println("   - POST /api/dingtalk/message/robot     发送机器人消息")
	log.Println("   - POST /api/dingtalk/approval          创建审批流程")
	log.Println()

	if err := r.Run(":8080"); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}

// 设置第三方服务
func setupThirdPartyServices(r *engine.Engine) {
	// 微信支付配置
	wechatPayConfig := &payment.WechatPayConfig{
		AppID:     "your_wechat_app_id",
		MchID:     "your_merchant_id",
		APIKey:    "your_api_key",
		NotifyURL: "https://your-domain.com/api/pay/wechat/notify",
		IsSandbox: true, // 沙箱环境
	}
	wechatPay := payment.NewWechatPay(wechatPayConfig)

	// 支付宝配置
	alipayConfig := &payment.AlipayConfig{
		AppID:      "your_alipay_app_id",
		PrivateKey: "your_private_key",
		PublicKey:  "your_public_key",
		NotifyURL:  "https://your-domain.com/api/pay/alipay/notify",
		ReturnURL:  "https://your-domain.com/pay/success",
		IsSandbox:  true, // 沙箱环境
	}
	alipay, err := payment.NewAlipay(alipayConfig)
	if err != nil {
		log.Printf("支付宝配置错误: %v", err)
	}

	// 微信小程序配置
	miniConfig := &miniprogram.WechatMiniConfig{
		AppID:     "your_miniprogram_app_id",
		AppSecret: "your_miniprogram_app_secret",
	}
	wechatMini := miniprogram.NewWechatMini(miniConfig)

	// 钉钉配置
	dingConfig := &dingtalk.DingTalkConfig{
		AppKey:       "your_dingtalk_app_key",
		AppSecret:    "your_dingtalk_app_secret",
		AgentID:      123456789,
		RobotToken:   "your_robot_token",
		RobotSecret:  "your_robot_secret",
		IsOldVersion: true,
	}
	dingTalk := dingtalk.NewDingTalk(dingConfig)

	// 设置路由
	setupPaymentRoutes(r, wechatPay, alipay)
	setupMiniprogramRoutes(r, wechatMini)
	setupDingTalkRoutes(r, dingTalk)
}

// 设置支付相关路由
func setupPaymentRoutes(r *engine.Engine, wechatPay *payment.WechatPay, alipay *payment.Alipay) {
	payGroup := r.Group("/api/pay")

	// 微信支付
	wechatGroup := payGroup.Group("/wechat")
	{
		// 创建支付订单
		wechatGroup.POST("/create", func(c *gyarn.Context) {
			var req struct {
				Body       string `json:"body"`
				OutTradeNo string `json:"out_trade_no"`
				TotalFee   int    `json:"total_fee"`
				TradeType  string `json:"trade_type"`
				OpenID     string `json:"openid,omitempty"`
			}

			if err := c.BindJSON(&req); err != nil {
				c.BadRequest("无效的请求参数")
				return
			}

			orderReq := &payment.UnifiedOrderRequest{
				Body:           req.Body,
				OutTradeNo:     req.OutTradeNo,
				TotalFee:       req.TotalFee,
				SpbillCreateIP: c.ClientIP(),
				TradeType:      req.TradeType,
				OpenID:         req.OpenID,
			}

			resp, err := wechatPay.UnifiedOrder(orderReq)
			if err != nil {
				c.InternalServerError("创建订单失败: " + err.Error())
				return
			}

			result := gyarn.H{
				"return_code": resp.ReturnCode,
				"return_msg":  resp.ReturnMsg,
				"result_code": resp.ResultCode,
				"prepay_id":   resp.PrepayID,
				"code_url":    resp.CodeURL,
			}

			// 如果是JSAPI支付，生成支付参数
			if req.TradeType == "JSAPI" && resp.PrepayID != "" {
				jsapiParams := wechatPay.GenerateJSAPIPayParams(resp.PrepayID)
				result["jsapi_params"] = jsapiParams
			}

			c.Success(result)
		})

		// 查询订单
		wechatGroup.POST("/query", func(c *gyarn.Context) {
			var req struct {
				OutTradeNo string `json:"out_trade_no"`
			}

			if err := c.BindJSON(&req); err != nil {
				c.BadRequest("无效的请求参数")
				return
			}

			resp, err := wechatPay.QueryOrder(req.OutTradeNo)
			if err != nil {
				c.InternalServerError("查询订单失败: " + err.Error())
				return
			}

			c.Success(gyarn.H{
				"return_code":    resp.ReturnCode,
				"result_code":    resp.ResultCode,
				"out_trade_no":   resp.OutTradeNo,
				"transaction_id": resp.TransactionID,
				"trade_state":    resp.TradeState,
				"total_fee":      resp.TotalFee,
				"time_end":       resp.TimeEnd,
			})
		})

		// 退款
		wechatGroup.POST("/refund", func(c *gyarn.Context) {
			var req struct {
				OutTradeNo  string `json:"out_trade_no"`
				OutRefundNo string `json:"out_refund_no"`
				TotalFee    int    `json:"total_fee"`
				RefundFee   int    `json:"refund_fee"`
			}

			if err := c.BindJSON(&req); err != nil {
				c.BadRequest("无效的请求参数")
				return
			}

			refundReq := &payment.RefundRequest{
				OutTradeNo:  req.OutTradeNo,
				OutRefundNo: req.OutRefundNo,
				TotalFee:    req.TotalFee,
				RefundFee:   req.RefundFee,
			}

			resp, err := wechatPay.Refund(refundReq)
			if err != nil {
				c.InternalServerError("退款失败: " + err.Error())
				return
			}

			c.Success(gyarn.H{
				"return_code":   resp.ReturnCode,
				"result_code":   resp.ResultCode,
				"refund_id":     resp.RefundID,
				"out_refund_no": resp.OutRefundNo,
				"refund_fee":    resp.RefundFee,
			})
		})

		// 支付回调
		wechatGroup.POST("/notify", func(c *gyarn.Context) {
			body, err := c.GetRawData()
			if err != nil {
				c.Error(400, "读取数据失败")
				return
			}

			notify, err := wechatPay.VerifyNotify(body)
			if err != nil {
				c.Error(400, "验证失败")
				return
			}

			// 处理支付成功逻辑
			log.Printf("支付成功: %s", notify.OutTradeNo)

			c.XML(200, `<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>`)
		})
	}

	// 支付宝
	if alipay != nil {
		alipayGroup := payGroup.Group("/alipay")
		{
			// 创建支付订单
			alipayGroup.POST("/create", func(c *gyarn.Context) {
				var req struct {
					OutTradeNo  string `json:"out_trade_no"`
					TotalAmount string `json:"total_amount"`
					Subject     string `json:"subject"`
					Body        string `json:"body"`
					ProductCode string `json:"product_code"`
				}

				if err := c.BindJSON(&req); err != nil {
					c.BadRequest("无效的请求参数")
					return
				}

				// 网页支付
				if req.ProductCode == "FAST_INSTANT_TRADE_PAY" {
					pageReq := &payment.AlipayTradePagePayRequest{
						OutTradeNo:  req.OutTradeNo,
						TotalAmount: req.TotalAmount,
						Subject:     req.Subject,
						Body:        req.Body,
						ProductCode: req.ProductCode,
					}

					payURL, err := alipay.TradePagePay(pageReq)
					if err != nil {
						c.InternalServerError("创建订单失败: " + err.Error())
						return
					}

					c.Success(gyarn.H{
						"pay_url": payURL,
					})
				} else {
					// 其他支付方式
					createReq := &payment.AlipayTradeCreateRequest{
						OutTradeNo:  req.OutTradeNo,
						TotalAmount: req.TotalAmount,
						Subject:     req.Subject,
						Body:        req.Body,
						ProductCode: req.ProductCode,
					}

					resp, err := alipay.TradeCreate(createReq)
					if err != nil {
						c.InternalServerError("创建订单失败: " + err.Error())
						return
					}

					c.Success(gyarn.H{
						"code":         resp.Code,
						"msg":          resp.Msg,
						"trade_no":     resp.TradeNo,
						"out_trade_no": resp.OutTradeNo,
					})
				}
			})

			// 查询订单
			alipayGroup.POST("/query", func(c *gyarn.Context) {
				var req struct {
					OutTradeNo string `json:"out_trade_no"`
					TradeNo    string `json:"trade_no"`
				}

				if err := c.BindJSON(&req); err != nil {
					c.BadRequest("无效的请求参数")
					return
				}

				queryReq := &payment.AlipayTradeQueryRequest{
					OutTradeNo: req.OutTradeNo,
					TradeNo:    req.TradeNo,
				}

				resp, err := alipay.TradeQuery(queryReq)
				if err != nil {
					c.InternalServerError("查询订单失败: " + err.Error())
					return
				}

				c.Success(gyarn.H{
					"code":         resp.Code,
					"msg":          resp.Msg,
					"trade_no":     resp.TradeNo,
					"out_trade_no": resp.OutTradeNo,
					"trade_status": resp.TradeStatus,
					"total_amount": resp.TotalAmount,
				})
			})

			// 退款
			alipayGroup.POST("/refund", func(c *gyarn.Context) {
				var req struct {
					OutTradeNo   string `json:"out_trade_no"`
					TradeNo      string `json:"trade_no"`
					RefundAmount string `json:"refund_amount"`
					RefundReason string `json:"refund_reason"`
					OutRequestNo string `json:"out_request_no"`
				}

				if err := c.BindJSON(&req); err != nil {
					c.BadRequest("无效的请求参数")
					return
				}

				refundReq := &payment.AlipayTradeRefundRequest{
					OutTradeNo:   req.OutTradeNo,
					TradeNo:      req.TradeNo,
					RefundAmount: req.RefundAmount,
					RefundReason: req.RefundReason,
					OutRequestNo: req.OutRequestNo,
				}

				resp, err := alipay.TradeRefund(refundReq)
				if err != nil {
					c.InternalServerError("退款失败: " + err.Error())
					return
				}

				c.Success(gyarn.H{
					"code":         resp.Code,
					"msg":          resp.Msg,
					"trade_no":     resp.TradeNo,
					"out_trade_no": resp.OutTradeNo,
					"refund_fee":   resp.RefundFee,
				})
			})
		}
	}
}

// 设置微信小程序路由
func setupMiniprogramRoutes(r *engine.Engine, wechatMini *miniprogram.WechatMini) {
	miniGroup := r.Group("/api/miniprogram")

	// 小程序登录
	miniGroup.POST("/login", func(c *gyarn.Context) {
		var req struct {
			Code string `json:"code"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		resp, err := wechatMini.Code2Session(req.Code)
		if err != nil {
			c.InternalServerError("登录失败: " + err.Error())
			return
		}

		c.Success(gyarn.H{
			"openid":      resp.OpenID,
			"session_key": resp.SessionKey,
			"unionid":     resp.UnionID,
		})
	})

	// 获取用户信息
	miniGroup.POST("/userinfo", func(c *gyarn.Context) {
		var req struct {
			SessionKey    string `json:"session_key"`
			EncryptedData string `json:"encrypted_data"`
			IV            string `json:"iv"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		userInfo, err := wechatMini.GetUserInfo(req.SessionKey, req.EncryptedData, req.IV)
		if err != nil {
			c.InternalServerError("获取用户信息失败: " + err.Error())
			return
		}

		c.Success(userInfo)
	})

	// 获取手机号
	miniGroup.POST("/phone", func(c *gyarn.Context) {
		var req struct {
			SessionKey    string `json:"session_key"`
			EncryptedData string `json:"encrypted_data"`
			IV            string `json:"iv"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		phoneInfo, err := wechatMini.GetPhoneNumber(req.SessionKey, req.EncryptedData, req.IV)
		if err != nil {
			c.InternalServerError("获取手机号失败: " + err.Error())
			return
		}

		c.Success(phoneInfo)
	})

	// 生成小程序码
	miniGroup.POST("/qrcode", func(c *gyarn.Context) {
		var req struct {
			Scene string `json:"scene"`
			Page  string `json:"page"`
			Width int    `json:"width"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		qrReq := &miniprogram.QRCodeRequest{
			Scene: req.Scene,
			Page:  req.Page,
			Width: req.Width,
		}

		qrCode, err := wechatMini.GenerateQRCode(qrReq)
		if err != nil {
			c.InternalServerError("生成小程序码失败: " + err.Error())
			return
		}

		// 返回二进制数据
		c.Data(200, "image/png", qrCode)
	})

	// 发送模板消息
	miniGroup.POST("/template", func(c *gyarn.Context) {
		var req struct {
			ToUser     string                 `json:"to_user"`
			TemplateID string                 `json:"template_id"`
			Data       map[string]interface{} `json:"data"`
			Page       string                 `json:"page"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		resp, err := wechatMini.SendSubscribeMessage(req.ToUser, req.TemplateID, req.Data, req.Page, "formal")
		if err != nil {
			c.InternalServerError("发送模板消息失败: " + err.Error())
			return
		}

		c.Success(gyarn.H{
			"errcode": resp.ErrCode,
			"errmsg":  resp.ErrMsg,
		})
	})
}

// 设置钉钉路由
func setupDingTalkRoutes(r *engine.Engine, dingTalk *dingtalk.DingTalk) {
	dingGroup := r.Group("/api/dingtalk")

	// 获取用户信息
	dingGroup.GET("/user/:id", func(c *gyarn.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.BadRequest("用户ID不能为空")
			return
		}

		userInfo, err := dingTalk.GetUserInfo(userID)
		if err != nil {
			c.InternalServerError("获取用户信息失败: " + err.Error())
			return
		}

		c.Success(userInfo)
	})

	// 根据手机号获取用户
	dingGroup.POST("/user/mobile", func(c *gyarn.Context) {
		var req struct {
			Mobile string `json:"mobile"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		userInfo, err := dingTalk.GetUserByMobile(req.Mobile)
		if err != nil {
			c.InternalServerError("获取用户信息失败: " + err.Error())
			return
		}

		c.Success(userInfo)
	})

	// 发送工作通知
	dingGroup.POST("/message/work", func(c *gyarn.Context) {
		var req struct {
			UserList  string `json:"user_list"`
			DeptList  string `json:"dept_list"`
			Content   string `json:"content"`
			ToAllUser bool   `json:"to_all_user"`
			Type      string `json:"type"` // text, markdown
			Title     string `json:"title,omitempty"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		var resp *dingtalk.SendMessageResponse
		var err error

		if req.Type == "markdown" {
			resp, err = dingTalk.SendMarkdownMessage(req.UserList, req.DeptList, req.Title, req.Content, req.ToAllUser)
		} else {
			resp, err = dingTalk.SendTextMessage(req.UserList, req.DeptList, req.Content, req.ToAllUser)
		}

		if err != nil {
			c.InternalServerError("发送消息失败: " + err.Error())
			return
		}

		c.Success(gyarn.H{
			"errcode": resp.ErrCode,
			"errmsg":  resp.ErrMsg,
			"task_id": resp.TaskID,
		})
	})

	// 发送机器人消息
	dingGroup.POST("/message/robot", func(c *gyarn.Context) {
		var req struct {
			Content   string   `json:"content"`
			Type      string   `json:"type"` // text, markdown
			Title     string   `json:"title,omitempty"`
			AtMobiles []string `json:"at_mobiles,omitempty"`
			AtUserIds []string `json:"at_user_ids,omitempty"`
			IsAtAll   bool     `json:"is_at_all,omitempty"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		var err error
		if req.Type == "markdown" {
			err = dingTalk.SendRobotMarkdownMessage(req.Title, req.Content, req.AtMobiles, req.AtUserIds, req.IsAtAll)
		} else {
			err = dingTalk.SendRobotTextMessage(req.Content, req.AtMobiles, req.AtUserIds, req.IsAtAll)
		}

		if err != nil {
			c.InternalServerError("发送机器人消息失败: " + err.Error())
			return
		}

		c.Success(gyarn.H{
			"message": "发送成功",
		})
	})

	// 创建审批流程
	dingGroup.POST("/approval", func(c *gyarn.Context) {
		var req struct {
			ProcessCode      string                        `json:"process_code"`
			OriginatorUserID string                        `json:"originator_user_id"`
			DeptID           int                           `json:"dept_id"`
			FormValues       []dingtalk.FormComponentValue `json:"form_values"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.BadRequest("无效的请求参数")
			return
		}

		approvalReq := &dingtalk.ApprovalProcessRequest{
			ProcessCode:         req.ProcessCode,
			OriginatorUserID:    req.OriginatorUserID,
			DeptID:              req.DeptID,
			FormComponentValues: req.FormValues,
		}

		resp, err := dingTalk.CreateApprovalProcess(approvalReq)
		if err != nil {
			c.InternalServerError("创建审批流程失败: " + err.Error())
			return
		}

		c.Success(gyarn.H{
			"errcode":             resp.ErrCode,
			"errmsg":              resp.ErrMsg,
			"process_instance_id": resp.ProcessInstanceID,
		})
	})
}
