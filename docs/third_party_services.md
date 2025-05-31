# 第三方服务集成指南

GyWeb 框架集成了腾讯、阿里的主要服务，提供完整的支付、小程序和企业应用开发支持。

## 概述

本框架提供以下第三方服务集成：

- **支付服务**: 微信支付、支付宝支付
- **微信小程序**: 登录、用户信息、模板消息、小程序码生成
- **钉钉集成**: 企业应用、消息推送、审批流程

## 支付集成

### 微信支付

#### 配置

```go
import "github.com/guyigood/gyweb/core/services/payment"

config := &payment.WechatPayConfig{
    AppID:     "your_wechat_app_id",     // 微信应用ID
    MchID:     "your_merchant_id",       // 商户号
    APIKey:    "your_api_key",           // API密钥
    NotifyURL: "https://your-domain.com/api/pay/wechat/notify", // 回调地址
    CertFile:  "/path/to/cert.pem",      // 证书文件路径（退款需要）
    KeyFile:   "/path/to/key.pem",       // 私钥文件路径（退款需要）
    IsSandbox: true,                     // 是否沙箱环境
}

wechatPay := payment.NewWechatPay(config)
```

#### 支付功能

**1. 统一下单**

```go
// JSAPI支付（小程序、公众号）
req := &payment.UnifiedOrderRequest{
    Body:           "商品描述",
    OutTradeNo:     "商户订单号",
    TotalFee:       100,  // 金额（分）
    SpbillCreateIP: "客户端IP",
    TradeType:      "JSAPI",
    OpenID:         "用户openid",
}

resp, err := wechatPay.UnifiedOrder(req)
if err != nil {
    // 处理错误
}

// 生成前端支付参数
if resp.PrepayID != "" {
    jsapiParams := wechatPay.GenerateJSAPIPayParams(resp.PrepayID)
    // 返回给前端
}
```

**2. Native支付（扫码支付）**

```go
codeURL, err := wechatPay.GenerateNativePayQR("商品描述", "订单号", 100, "客户端IP")
if err != nil {
    // 处理错误
}
// codeURL 用于生成二维码
```

**3. 查询订单**

```go
resp, err := wechatPay.QueryOrder("商户订单号")
if err != nil {
    // 处理错误
}

// 检查支付状态
if resp.TradeState == "SUCCESS" {
    // 支付成功
}
```

**4. 申请退款**

```go
refundReq := &payment.RefundRequest{
    OutTradeNo:  "原订单号",
    OutRefundNo: "退款单号",
    TotalFee:    100,  // 原订单金额
    RefundFee:   50,   // 退款金额
}

resp, err := wechatPay.Refund(refundReq)
```

**5. 支付回调处理**

```go
func handleWechatNotify(c *gyarn.Context) {
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

    // 处理支付成功业务逻辑
    if notify.ResultCode == "SUCCESS" {
        // 订单支付成功处理
        log.Printf("订单 %s 支付成功", notify.OutTradeNo)
    }

    // 返回微信确认
    c.XML(200, `<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>`)
}
```

### 支付宝支付

#### 配置

```go
config := &payment.AlipayConfig{
    AppID:      "your_alipay_app_id",
    PrivateKey: `-----BEGIN PRIVATE KEY-----
your_private_key_content
-----END PRIVATE KEY-----`,
    PublicKey:  `-----BEGIN PUBLIC KEY-----
alipay_public_key_content
-----END PUBLIC KEY-----`,
    NotifyURL:  "https://your-domain.com/api/pay/alipay/notify",
    ReturnURL:  "https://your-domain.com/pay/success",
    IsSandbox:  true,
}

alipay, err := payment.NewAlipay(config)
```

#### 支付功能

**1. 网页支付**

```go
req := &payment.AlipayTradePagePayRequest{
    OutTradeNo:  "订单号",
    TotalAmount: "1.00",  // 金额（元）
    Subject:     "商品标题",
    Body:        "商品描述",
    ProductCode: "FAST_INSTANT_TRADE_PAY",
}

payURL, err := alipay.TradePagePay(req)
// 重定向到 payURL 进行支付
```

**2. 查询订单**

```go
req := &payment.AlipayTradeQueryRequest{
    OutTradeNo: "商户订单号",
    // 或使用支付宝订单号
    // TradeNo: "支付宝订单号",
}

resp, err := alipay.TradeQuery(req)
```

**3. 申请退款**

```go
req := &payment.AlipayTradeRefundRequest{
    OutTradeNo:   "原订单号",
    RefundAmount: "0.50",  // 退款金额
    RefundReason: "退款原因",
    OutRequestNo: "退款请求号",
}

resp, err := alipay.TradeRefund(req)
```

**4. 验证异步通知**

```go
func handleAlipayNotify(c *gyarn.Context) {
    // 获取所有POST参数
    params := make(map[string]string)
    for k, v := range c.Request.Form {
        if len(v) > 0 {
            params[k] = v[0]
        }
    }

    // 验证签名
    if !alipay.VerifyNotify(params) {
        c.Error(400, "签名验证失败")
        return
    }

    // 处理业务逻辑
    if params["trade_status"] == "TRADE_SUCCESS" {
        log.Printf("订单 %s 支付成功", params["out_trade_no"])
    }

    c.String(200, "success")
}
```

## 微信小程序集成

### 配置

```go
import "github.com/guyigood/gyweb/core/services/miniprogram"

config := &miniprogram.WechatMiniConfig{
    AppID:     "your_miniprogram_app_id",
    AppSecret: "your_miniprogram_app_secret",
}

wechatMini := miniprogram.NewWechatMini(config)
```

### 功能说明

**1. 小程序登录**

```go
// 后端接口
func miniLogin(c *gyarn.Context) {
    var req struct {
        Code string `json:"code"`
    }
    c.BindJSON(&req)

    resp, err := wechatMini.Code2Session(req.Code)
    if err != nil {
        c.InternalServerError("登录失败")
        return
    }

    c.Success(gyarn.H{
        "openid":      resp.OpenID,
        "session_key": resp.SessionKey,
        "unionid":     resp.UnionID,
    })
}
```

前端调用：
```javascript
// 小程序端
wx.login({
    success: res => {
        if (res.code) {
            // 发送 code 到后端
            wx.request({
                url: 'https://your-api.com/api/miniprogram/login',
                method: 'POST',
                data: { code: res.code },
                success: data => {
                    // 处理登录结果
                    console.log('登录成功', data);
                }
            });
        }
    }
});
```

**2. 获取用户信息**

```go
func getUserInfo(c *gyarn.Context) {
    var req struct {
        SessionKey    string `json:"session_key"`
        EncryptedData string `json:"encrypted_data"`
        IV            string `json:"iv"`
    }
    c.BindJSON(&req)

    userInfo, err := wechatMini.GetUserInfo(req.SessionKey, req.EncryptedData, req.IV)
    if err != nil {
        c.InternalServerError("获取用户信息失败")
        return
    }

    c.Success(userInfo)
}
```

**3. 获取手机号**

```go
func getPhoneNumber(c *gyarn.Context) {
    var req struct {
        SessionKey    string `json:"session_key"`
        EncryptedData string `json:"encrypted_data"`
        IV            string `json:"iv"`
    }
    c.BindJSON(&req)

    phoneInfo, err := wechatMini.GetPhoneNumber(req.SessionKey, req.EncryptedData, req.IV)
    if err != nil {
        c.InternalServerError("获取手机号失败")
        return
    }

    c.Success(phoneInfo)
}
```

**4. 生成小程序码**

```go
func generateQRCode(c *gyarn.Context) {
    req := &miniprogram.QRCodeRequest{
        Scene: "user_id=123",  // 自定义参数
        Page:  "pages/index/index",  // 小程序页面路径
        Width: 430,  // 二维码宽度
    }

    qrCode, err := wechatMini.GenerateQRCode(req)
    if err != nil {
        c.InternalServerError("生成小程序码失败")
        return
    }

    // 返回图片数据
    c.Data(200, "image/png", qrCode)
}
```

**5. 发送订阅消息**

```go
func sendSubscribeMessage(c *gyarn.Context) {
    data := map[string]interface{}{
        "thing1": map[string]string{"value": "订单标题"},
        "time2":  map[string]string{"value": "2024-01-01 12:00:00"},
        "amount3": map[string]string{"value": "100.00元"},
    }

    resp, err := wechatMini.SendSubscribeMessage(
        "用户openid",
        "模板ID",
        data,
        "pages/order/detail?id=123",  // 跳转页面
        "formal",  // 小程序版本
    )

    if err != nil {
        c.InternalServerError("发送消息失败")
        return
    }

    c.Success(resp)
}
```

## 钉钉集成

### 配置

```go
import "github.com/guyigood/gyweb/core/services/dingtalk"

config := &dingtalk.DingTalkConfig{
    AppKey:       "your_dingtalk_app_key",
    AppSecret:    "your_dingtalk_app_secret",
    AgentID:      123456789,  // 应用AgentID
    RobotToken:   "robot_webhook_token",  // 机器人token
    RobotSecret:  "robot_secret",  // 机器人加签密钥
    IsOldVersion: true,  // 是否使用旧版API
}

dingTalk := dingtalk.NewDingTalk(config)
```

### 功能说明

**1. 获取用户信息**

```go
// 根据用户ID获取
userInfo, err := dingTalk.GetUserInfo("user123")

// 根据手机号获取
userInfo, err := dingTalk.GetUserByMobile("13800138000")
```

**2. 发送工作通知**

```go
// 发送文本消息
resp, err := dingTalk.SendTextMessage(
    "user1,user2",  // 用户ID列表
    "dept1,dept2",  // 部门ID列表
    "这是一条工作通知",  // 消息内容
    false,  // 是否发送给全员
)

// 发送Markdown消息
resp, err := dingTalk.SendMarkdownMessage(
    "user1,user2",
    "dept1,dept2",
    "通知标题",
    "## 这是Markdown消息\n**重要内容**",
    false,
)
```

**3. 发送机器人消息**

```go
// 发送文本消息并@指定人员
err := dingTalk.SendRobotTextMessage(
    "这是机器人消息 @13800138000",
    []string{"13800138000"},  // @的手机号
    []string{},  // @的用户ID
    false,  // 是否@所有人
)

// 发送Markdown消息
err := dingTalk.SendRobotMarkdownMessage(
    "报警通知",
    "## 系统报警\n> 服务器负载过高\n- 时间: 2024-01-01 12:00:00\n- 负载: 90%",
    []string{"13800138000"},
    []string{},
    false,
)
```

**4. 创建审批流程**

```go
formValues := []dingtalk.FormComponentValue{
    {
        Name:  "请假类型",
        Value: "年假",
    },
    {
        Name:  "请假天数",
        Value: "3",
    },
    {
        Name:  "请假原因",
        Value: "休息调整",
    },
}

req := &dingtalk.ApprovalProcessRequest{
    ProcessCode:         "PROC-123456",  // 审批流程代码
    OriginatorUserID:    "user123",      // 发起人ID
    DeptID:              1001,           // 部门ID
    FormComponentValues: formValues,     // 表单数据
}

resp, err := dingTalk.CreateApprovalProcess(req)
if err != nil {
    log.Printf("创建审批失败: %v", err)
} else {
    log.Printf("审批实例ID: %s", resp.ProcessInstanceID)
}
```

## API 路由示例

### 集成到路由中

```go
func setupAPI(r *engine.Engine) {
    // 支付路由
    payGroup := r.Group("/api/pay")
    payGroup.POST("/wechat/create", createWechatOrder)
    payGroup.POST("/wechat/query", queryWechatOrder)
    payGroup.POST("/wechat/notify", handleWechatNotify)
    payGroup.POST("/alipay/create", createAlipayOrder)
    payGroup.POST("/alipay/notify", handleAlipayNotify)

    // 小程序路由
    miniGroup := r.Group("/api/miniprogram")
    miniGroup.POST("/login", miniLogin)
    miniGroup.POST("/userinfo", getUserInfo)
    miniGroup.POST("/phone", getPhoneNumber)
    miniGroup.POST("/qrcode", generateQRCode)
    miniGroup.POST("/template", sendSubscribeMessage)

    // 钉钉路由
    dingGroup := r.Group("/api/dingtalk")
    dingGroup.GET("/user/:id", getDingUser)
    dingGroup.POST("/message/work", sendWorkMessage)
    dingGroup.POST("/message/robot", sendRobotMessage)
    dingGroup.POST("/approval", createApproval)
}
```

## 安全建议

### 支付安全
1. **生产环境必须使用HTTPS**
2. **妥善保管API密钥和证书文件**
3. **验证回调签名**
4. **实现幂等性处理**

### 小程序安全
1. **验证数据水印**
2. **不在前端存储敏感信息**
3. **使用session_key时注意过期时间**

### 钉钉安全
1. **使用机器人加签验证**
2. **限制API调用频率**
3. **验证来源IP**

## 常见问题

### Q: 微信支付回调验证失败
A: 检查以下几点：
- 确认API密钥配置正确
- 验证XML格式解析
- 检查签名算法实现

### Q: 支付宝RSA签名验证失败
A: 检查：
- 私钥和公钥格式是否正确
- 是否使用了正确的签名类型（RSA2推荐）
- 参数排序和编码是否正确

### Q: 小程序解密数据失败
A: 确认：
- session_key是否有效
- iv和encryptedData格式是否正确
- AES解密算法实现

### Q: 钉钉机器人发送失败
A: 检查：
- webhook地址是否正确
- 是否启用了加签验证
- 消息格式是否符合要求

## 示例项目

完整的示例代码请参考：`examples/third_party_services_example.go`

该示例包含了所有第三方服务的完整集成代码，可以直接运行测试。 