package payment

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// WechatPayConfig 微信支付配置
type WechatPayConfig struct {
	AppID     string `json:"app_id"`     // 应用ID
	MchID     string `json:"mch_id"`     // 商户号
	APIKey    string `json:"api_key"`    // API密钥
	NotifyURL string `json:"notify_url"` // 回调地址
	CertFile  string `json:"cert_file"`  // 证书文件路径
	KeyFile   string `json:"key_file"`   // 私钥文件路径
	IsSandbox bool   `json:"is_sandbox"` // 是否沙箱环境
}

// WechatPay 微信支付客户端
type WechatPay struct {
	config *WechatPayConfig
	client *http.Client
}

// UnifiedOrderRequest 统一下单请求
type UnifiedOrderRequest struct {
	AppID          string `xml:"appid"`            // 应用ID
	MchID          string `xml:"mch_id"`           // 商户号
	NonceStr       string `xml:"nonce_str"`        // 随机字符串
	Sign           string `xml:"sign"`             // 签名
	Body           string `xml:"body"`             // 商品描述
	OutTradeNo     string `xml:"out_trade_no"`     // 商户订单号
	TotalFee       int    `xml:"total_fee"`        // 总金额（分）
	SpbillCreateIP string `xml:"spbill_create_ip"` // 终端IP
	NotifyURL      string `xml:"notify_url"`       // 通知地址
	TradeType      string `xml:"trade_type"`       // 交易类型
	OpenID         string `xml:"openid,omitempty"` // 用户openid（JSAPI需要）
}

// UnifiedOrderResponse 统一下单响应
type UnifiedOrderResponse struct {
	ReturnCode string `xml:"return_code"`  // 返回状态码
	ReturnMsg  string `xml:"return_msg"`   // 返回信息
	AppID      string `xml:"appid"`        // 应用ID
	MchID      string `xml:"mch_id"`       // 商户号
	NonceStr   string `xml:"nonce_str"`    // 随机字符串
	Sign       string `xml:"sign"`         // 签名
	ResultCode string `xml:"result_code"`  // 业务结果
	PrepayID   string `xml:"prepay_id"`    // 预支付交易会话标识
	TradeType  string `xml:"trade_type"`   // 交易类型
	CodeURL    string `xml:"code_url"`     // 二维码链接
	ErrCode    string `xml:"err_code"`     // 错误代码
	ErrCodeDes string `xml:"err_code_des"` // 错误代码描述
}

// OrderQueryResponse 订单查询响应
type OrderQueryResponse struct {
	ReturnCode    string `xml:"return_code"`    // 返回状态码
	ReturnMsg     string `xml:"return_msg"`     // 返回信息
	ResultCode    string `xml:"result_code"`    // 业务结果
	OutTradeNo    string `xml:"out_trade_no"`   // 商户订单号
	TransactionID string `xml:"transaction_id"` // 微信支付订单号
	TradeState    string `xml:"trade_state"`    // 交易状态
	TotalFee      int    `xml:"total_fee"`      // 订单金额
	TimeEnd       string `xml:"time_end"`       // 支付完成时间
}

// RefundRequest 退款请求
type RefundRequest struct {
	AppID         string `xml:"appid"`          // 应用ID
	MchID         string `xml:"mch_id"`         // 商户号
	NonceStr      string `xml:"nonce_str"`      // 随机字符串
	Sign          string `xml:"sign"`           // 签名
	TransactionID string `xml:"transaction_id"` // 微信订单号
	OutTradeNo    string `xml:"out_trade_no"`   // 商户订单号
	OutRefundNo   string `xml:"out_refund_no"`  // 商户退款单号
	TotalFee      int    `xml:"total_fee"`      // 订单金额
	RefundFee     int    `xml:"refund_fee"`     // 退款金额
}

// RefundResponse 退款响应
type RefundResponse struct {
	ReturnCode  string `xml:"return_code"`   // 返回状态码
	ReturnMsg   string `xml:"return_msg"`    // 返回信息
	ResultCode  string `xml:"result_code"`   // 业务结果
	RefundID    string `xml:"refund_id"`     // 微信退款单号
	OutRefundNo string `xml:"out_refund_no"` // 商户退款单号
	RefundFee   int    `xml:"refund_fee"`    // 申请退款金额
	TotalFee    int    `xml:"total_fee"`     // 订单总金额
}

// NotifyData 支付通知数据
type NotifyData struct {
	ReturnCode    string `xml:"return_code"`    // 返回状态码
	ReturnMsg     string `xml:"return_msg"`     // 返回信息
	ResultCode    string `xml:"result_code"`    // 业务结果
	OpenID        string `xml:"openid"`         // 用户openid
	IsSubscribe   string `xml:"is_subscribe"`   // 是否关注公众账号
	TradeType     string `xml:"trade_type"`     // 交易类型
	BankType      string `xml:"bank_type"`      // 付款银行
	TotalFee      int    `xml:"total_fee"`      // 订单金额
	FeeType       string `xml:"fee_type"`       // 货币种类
	TransactionID string `xml:"transaction_id"` // 微信支付订单号
	OutTradeNo    string `xml:"out_trade_no"`   // 商户订单号
	Attach        string `xml:"attach"`         // 商家数据包
	TimeEnd       string `xml:"time_end"`       // 支付完成时间
}

// NewWechatPay 创建微信支付客户端
func NewWechatPay(config *WechatPayConfig) *WechatPay {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 如果有证书文件，配置SSL
	if config.CertFile != "" && config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err == nil {
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
			client.Transport = &http.Transport{
				TLSClientConfig: tlsConfig,
			}
		}
	}

	return &WechatPay{
		config: config,
		client: client,
	}
}

// generateNonceStr 生成随机字符串
func (w *WechatPay) generateNonceStr() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// signParams 生成签名
func (w *WechatPay) signParams(params map[string]interface{}) string {
	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signStr strings.Builder
	for i, key := range keys {
		if i > 0 {
			signStr.WriteString("&")
		}
		signStr.WriteString(fmt.Sprintf("%s=%v", key, params[key]))
	}
	signStr.WriteString("&key=" + w.config.APIKey)

	// MD5签名
	h := md5.New()
	h.Write([]byte(signStr.String()))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

// UnifiedOrder 统一下单
func (w *WechatPay) UnifiedOrder(req *UnifiedOrderRequest) (*UnifiedOrderResponse, error) {
	// 设置基本参数
	req.AppID = w.config.AppID
	req.MchID = w.config.MchID
	req.NonceStr = w.generateNonceStr()
	req.NotifyURL = w.config.NotifyURL

	// 生成签名
	params := map[string]interface{}{
		"appid":            req.AppID,
		"mch_id":           req.MchID,
		"nonce_str":        req.NonceStr,
		"body":             req.Body,
		"out_trade_no":     req.OutTradeNo,
		"total_fee":        req.TotalFee,
		"spbill_create_ip": req.SpbillCreateIP,
		"notify_url":       req.NotifyURL,
		"trade_type":       req.TradeType,
	}
	if req.OpenID != "" {
		params["openid"] = req.OpenID
	}
	req.Sign = w.signParams(params)

	// 构建XML
	xmlData, err := xml.Marshal(req)
	if err != nil {
		return nil, err
	}

	// 发送请求
	url := "https://api.mch.weixin.qq.com/pay/unifiedorder"
	if w.config.IsSandbox {
		url = "https://api.mch.weixin.qq.com/sandboxnew/pay/unifiedorder"
	}

	resp, err := w.client.Post(url, "application/xml", strings.NewReader(string(xmlData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response UnifiedOrderResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// QueryOrder 查询订单
func (w *WechatPay) QueryOrder(outTradeNo string) (*OrderQueryResponse, error) {
	params := map[string]interface{}{
		"appid":        w.config.AppID,
		"mch_id":       w.config.MchID,
		"out_trade_no": outTradeNo,
		"nonce_str":    w.generateNonceStr(),
	}
	sign := w.signParams(params)

	// 构建XML
	xmlData := fmt.Sprintf(`<xml>
		<appid>%s</appid>
		<mch_id>%s</mch_id>
		<out_trade_no>%s</out_trade_no>
		<nonce_str>%s</nonce_str>
		<sign>%s</sign>
	</xml>`, w.config.AppID, w.config.MchID, outTradeNo, params["nonce_str"], sign)

	// 发送请求
	url := "https://api.mch.weixin.qq.com/pay/orderquery"
	if w.config.IsSandbox {
		url = "https://api.mch.weixin.qq.com/sandboxnew/pay/orderquery"
	}

	resp, err := w.client.Post(url, "application/xml", strings.NewReader(xmlData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response OrderQueryResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Refund 申请退款
func (w *WechatPay) Refund(req *RefundRequest) (*RefundResponse, error) {
	// 设置基本参数
	req.AppID = w.config.AppID
	req.MchID = w.config.MchID
	req.NonceStr = w.generateNonceStr()

	// 生成签名
	params := map[string]interface{}{
		"appid":         req.AppID,
		"mch_id":        req.MchID,
		"nonce_str":     req.NonceStr,
		"out_trade_no":  req.OutTradeNo,
		"out_refund_no": req.OutRefundNo,
		"total_fee":     req.TotalFee,
		"refund_fee":    req.RefundFee,
	}
	if req.TransactionID != "" {
		params["transaction_id"] = req.TransactionID
	}
	req.Sign = w.signParams(params)

	// 构建XML
	xmlData, err := xml.Marshal(req)
	if err != nil {
		return nil, err
	}

	// 发送请求（需要证书）
	url := "https://api.mch.weixin.qq.com/secapi/pay/refund"
	if w.config.IsSandbox {
		url = "https://api.mch.weixin.qq.com/sandboxnew/secapi/pay/refund"
	}

	resp, err := w.client.Post(url, "application/xml", strings.NewReader(string(xmlData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response RefundResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// VerifyNotify 验证支付通知
func (w *WechatPay) VerifyNotify(xmlData []byte) (*NotifyData, error) {
	var notify NotifyData
	err := xml.Unmarshal(xmlData, &notify)
	if err != nil {
		return nil, err
	}

	// 验证签名
	// TODO: 实现签名验证逻辑

	return &notify, nil
}

// GenerateJSAPIPayParams 生成JSAPI支付参数
func (w *WechatPay) GenerateJSAPIPayParams(prepayID string) map[string]string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr := w.generateNonceStr()
	packageStr := "prepay_id=" + prepayID

	params := map[string]interface{}{
		"appId":     w.config.AppID,
		"timeStamp": timestamp,
		"nonceStr":  nonceStr,
		"package":   packageStr,
		"signType":  "MD5",
	}

	paySign := w.signParams(params)

	return map[string]string{
		"appId":     w.config.AppID,
		"timeStamp": timestamp,
		"nonceStr":  nonceStr,
		"package":   packageStr,
		"signType":  "MD5",
		"paySign":   paySign,
	}
}

// GenerateNativePayQR 生成Native支付二维码
func (w *WechatPay) GenerateNativePayQR(body string, outTradeNo string, totalFee int, clientIP string) (string, error) {
	req := &UnifiedOrderRequest{
		Body:           body,
		OutTradeNo:     outTradeNo,
		TotalFee:       totalFee,
		SpbillCreateIP: clientIP,
		TradeType:      "NATIVE",
	}

	resp, err := w.UnifiedOrder(req)
	if err != nil {
		return "", err
	}

	if resp.ReturnCode != "SUCCESS" || resp.ResultCode != "SUCCESS" {
		return "", fmt.Errorf("unified order failed: %s", resp.ReturnMsg)
	}

	return resp.CodeURL, nil
}
