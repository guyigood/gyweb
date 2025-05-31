package payment

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	AppID      string `json:"app_id"`      // 应用ID
	PrivateKey string `json:"private_key"` // 应用私钥
	PublicKey  string `json:"public_key"`  // 支付宝公钥
	NotifyURL  string `json:"notify_url"`  // 异步通知地址
	ReturnURL  string `json:"return_url"`  // 同步跳转地址
	IsSandbox  bool   `json:"is_sandbox"`  // 是否沙箱环境
	SignType   string `json:"sign_type"`   // 签名类型，默认RSA2
	Format     string `json:"format"`      // 数据格式，默认JSON
	Charset    string `json:"charset"`     // 编码格式，默认utf-8
	Version    string `json:"version"`     // 接口版本，默认1.0
}

// Alipay 支付宝客户端
type Alipay struct {
	config     *AlipayConfig
	client     *http.Client
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// AlipayTradeCreateRequest 统一收单交易创建接口请求参数
type AlipayTradeCreateRequest struct {
	OutTradeNo     string `json:"out_trade_no"`              // 商户订单号
	TotalAmount    string `json:"total_amount"`              // 订单总金额
	Subject        string `json:"subject"`                   // 订单标题
	Body           string `json:"body,omitempty"`            // 订单描述
	BuyerID        string `json:"buyer_id,omitempty"`        // 买家支付宝用户ID
	TimeoutExpress string `json:"timeout_express,omitempty"` // 该笔订单允许的最晚付款时间
	ProductCode    string `json:"product_code"`              // 销售产品码
}

// AlipayTradePagePayRequest 统一收单下单并支付页面接口请求参数
type AlipayTradePagePayRequest struct {
	OutTradeNo     string `json:"out_trade_no"`              // 商户订单号
	TotalAmount    string `json:"total_amount"`              // 订单总金额
	Subject        string `json:"subject"`                   // 订单标题
	Body           string `json:"body,omitempty"`            // 订单描述
	TimeoutExpress string `json:"timeout_express,omitempty"` // 该笔订单允许的最晚付款时间
	ProductCode    string `json:"product_code"`              // 销售产品码
}

// AlipayTradeQueryRequest 统一收单线上交易查询请求参数
type AlipayTradeQueryRequest struct {
	OutTradeNo string `json:"out_trade_no,omitempty"` // 商户订单号
	TradeNo    string `json:"trade_no,omitempty"`     // 支付宝交易号
}

// AlipayTradeRefundRequest 统一收单交易退款接口请求参数
type AlipayTradeRefundRequest struct {
	OutTradeNo   string `json:"out_trade_no,omitempty"`   // 商户订单号
	TradeNo      string `json:"trade_no,omitempty"`       // 支付宝交易号
	RefundAmount string `json:"refund_amount"`            // 退款金额
	RefundReason string `json:"refund_reason,omitempty"`  // 退款原因说明
	OutRequestNo string `json:"out_request_no,omitempty"` // 商户退款请求号
}

// AlipayResponse 支付宝响应基础结构
type AlipayResponse struct {
	Code    string `json:"code"`     // 网关返回码
	Msg     string `json:"msg"`      // 网关返回码描述
	SubCode string `json:"sub_code"` // 业务返回码
	SubMsg  string `json:"sub_msg"`  // 业务返回码描述
	Sign    string `json:"sign"`     // 签名
}

// AlipayTradeCreateResponse 统一收单交易创建接口响应
type AlipayTradeCreateResponse struct {
	AlipayResponse
	TradeNo    string `json:"trade_no"`     // 支付宝交易号
	OutTradeNo string `json:"out_trade_no"` // 商户订单号
}

// AlipayTradeQueryResponse 统一收单线上交易查询响应
type AlipayTradeQueryResponse struct {
	AlipayResponse
	TradeNo             string `json:"trade_no"`               // 支付宝交易号
	OutTradeNo          string `json:"out_trade_no"`           // 商户订单号
	BuyerLogonID        string `json:"buyer_logon_id"`         // 买家支付宝账号
	TradeStatus         string `json:"trade_status"`           // 交易状态
	TotalAmount         string `json:"total_amount"`           // 交易金额
	TransCurrency       string `json:"trans_currency"`         // 标价币种
	SettleCurrency      string `json:"settle_currency"`        // 订单结算币种
	SettleAmount        string `json:"settle_amount"`          // 结算币种订单金额
	PayCurrency         string `json:"pay_currency"`           // 订单支付币种
	PayAmount           string `json:"pay_amount"`             // 支付币种订单金额
	SettleTransRate     string `json:"settle_trans_rate"`      // 结算币种兑换标价币种汇率
	TransPayRate        string `json:"trans_pay_rate"`         // 标价币种兑换支付币种汇率
	BuyerPayAmount      string `json:"buyer_pay_amount"`       // 买家实付金额
	PointAmount         string `json:"point_amount"`           // 积分支付的金额
	InvoiceAmount       string `json:"invoice_amount"`         // 交易中用户支付的可开具发票的金额
	SendPayDate         string `json:"send_pay_date"`          // 本次交易打款给卖家的时间
	ReceiptAmount       string `json:"receipt_amount"`         // 实收金额
	StoreID             string `json:"store_id"`               // 商户门店编号
	TerminalID          string `json:"terminal_id"`            // 商户机具终端编号
	FundBillList        string `json:"fund_bill_list"`         // 交易支付使用的资金渠道
	StoreName           string `json:"store_name"`             // 请求交易支付中的商户店铺的名称
	BuyerUserID         string `json:"buyer_user_id"`          // 买家在支付宝的用户id
	ChargeAmount        string `json:"charge_amount"`          // 该笔交易针对收款方的收费金额
	ChargeFlags         string `json:"charge_flags"`           // 费率活动标识
	SettlementID        string `json:"settlement_id"`          // 支付清算编号
	TradeSettleInfo     string `json:"trade_settle_info"`      // 返回的交易结算信息
	AuthTradePayMode    string `json:"auth_trade_pay_mode"`    // 预授权支付模式
	BuyerUserType       string `json:"buyer_user_type"`        // 买家用户类型
	MdiscountAmount     string `json:"mdiscount_amount"`       // 商家优惠金额
	DiscountAmount      string `json:"discount_amount"`        // 平台优惠金额
	Subject             string `json:"subject"`                // 订单标题
	Body                string `json:"body"`                   // 订单描述
	AlipaySubMerchantID string `json:"alipay_sub_merchant_id"` // 间连受理商户信息体
	ExtInfos            string `json:"ext_infos"`              // 交易额外信息
}

// AlipayTradeRefundResponse 统一收单交易退款接口响应
type AlipayTradeRefundResponse struct {
	AlipayResponse
	TradeNo                 string `json:"trade_no"`                   // 支付宝交易号
	OutTradeNo              string `json:"out_trade_no"`               // 商户订单号
	BuyerLogonID            string `json:"buyer_logon_id"`             // 用户的登录id
	FundChange              string `json:"fund_change"`                // 本次退款是否发生了资金变化
	RefundFee               string `json:"refund_fee"`                 // 退款总金额
	RefundCurrency          string `json:"refund_currency"`            // 退款币种信息
	GMTRefundPay            string `json:"gmt_refund_pay"`             // 退款支付时间
	RefundDetailItemList    string `json:"refund_detail_item_list"`    // 退款使用的资金渠道
	StoreName               string `json:"store_name"`                 // 交易在支付时候的门店名称
	BuyerUserID             string `json:"buyer_user_id"`              // 买家在支付宝的用户id
	RefundPresetPaytoolList string `json:"refund_preset_paytool_list"` // 退回的前置资产列表
}

// NewAlipay 创建支付宝客户端
func NewAlipay(config *AlipayConfig) (*Alipay, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 设置默认值
	if config.SignType == "" {
		config.SignType = "RSA2"
	}
	if config.Format == "" {
		config.Format = "JSON"
	}
	if config.Charset == "" {
		config.Charset = "utf-8"
	}
	if config.Version == "" {
		config.Version = "1.0"
	}

	alipay := &Alipay{
		config: config,
		client: client,
	}

	// 解析私钥
	if err := alipay.loadPrivateKey(config.PrivateKey); err != nil {
		return nil, fmt.Errorf("加载私钥失败: %v", err)
	}

	// 解析公钥
	if err := alipay.loadPublicKey(config.PublicKey); err != nil {
		return nil, fmt.Errorf("加载公钥失败: %v", err)
	}

	return alipay, nil
}

// loadPrivateKey 加载私钥
func (a *Alipay) loadPrivateKey(privateKeyStr string) error {
	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return fmt.Errorf("私钥格式错误")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return err
		}
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("私钥不是RSA格式")
	}

	a.privateKey = rsaPrivateKey
	return nil
}

// loadPublicKey 加载公钥
func (a *Alipay) loadPublicKey(publicKeyStr string) error {
	block, _ := pem.Decode([]byte(publicKeyStr))
	if block == nil {
		return fmt.Errorf("公钥格式错误")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("公钥不是RSA格式")
	}

	a.publicKey = rsaPublicKey
	return nil
}

// getGatewayURL 获取网关地址
func (a *Alipay) getGatewayURL() string {
	if a.config.IsSandbox {
		return "https://openapi.alipaydev.com/gateway.do"
	}
	return "https://openapi.alipay.com/gateway.do"
}

// buildCommonParams 构建公共参数
func (a *Alipay) buildCommonParams(method string) map[string]string {
	params := map[string]string{
		"app_id":    a.config.AppID,
		"method":    method,
		"format":    a.config.Format,
		"charset":   a.config.Charset,
		"sign_type": a.config.SignType,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"version":   a.config.Version,
	}

	if a.config.NotifyURL != "" {
		params["notify_url"] = a.config.NotifyURL
	}
	if a.config.ReturnURL != "" {
		params["return_url"] = a.config.ReturnURL
	}

	return params
}

// sign 生成签名
func (a *Alipay) sign(params map[string]string) (string, error) {
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
		signStr.WriteString(fmt.Sprintf("%s=%s", key, params[key]))
	}

	// 生成签名
	var hash crypto.Hash
	if a.config.SignType == "RSA2" {
		hash = crypto.SHA256
	} else {
		hash = crypto.SHA1
	}

	h := hash.New()
	h.Write([]byte(signStr.String()))
	hashed := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, a.privateKey, hash, hashed)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// request 发送请求
func (a *Alipay) request(method string, bizContent interface{}) ([]byte, error) {
	params := a.buildCommonParams(method)

	// 序列化业务参数
	if bizContent != nil {
		bizContentJSON, err := json.Marshal(bizContent)
		if err != nil {
			return nil, err
		}
		params["biz_content"] = string(bizContentJSON)
	}

	// 生成签名
	sign, err := a.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	// 构建请求参数
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	// 发送请求
	resp, err := a.client.PostForm(a.getGatewayURL(), values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// TradeCreate 统一收单交易创建接口
func (a *Alipay) TradeCreate(req *AlipayTradeCreateRequest) (*AlipayTradeCreateResponse, error) {
	body, err := a.request("alipay.trade.create", req)
	if err != nil {
		return nil, err
	}

	var response struct {
		AlipayTradeCreateResponse `json:"alipay_trade_create_response"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.AlipayTradeCreateResponse, nil
}

// TradePagePay 统一收单下单并支付页面接口
func (a *Alipay) TradePagePay(req *AlipayTradePagePayRequest) (string, error) {
	params := a.buildCommonParams("alipay.trade.page.pay")

	// 序列化业务参数
	bizContentJSON, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	params["biz_content"] = string(bizContentJSON)

	// 生成签名
	sign, err := a.sign(params)
	if err != nil {
		return "", err
	}
	params["sign"] = sign

	// 构建支付链接
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	return a.getGatewayURL() + "?" + values.Encode(), nil
}

// TradeQuery 统一收单线上交易查询
func (a *Alipay) TradeQuery(req *AlipayTradeQueryRequest) (*AlipayTradeQueryResponse, error) {
	body, err := a.request("alipay.trade.query", req)
	if err != nil {
		return nil, err
	}

	var response struct {
		AlipayTradeQueryResponse `json:"alipay_trade_query_response"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.AlipayTradeQueryResponse, nil
}

// TradeRefund 统一收单交易退款接口
func (a *Alipay) TradeRefund(req *AlipayTradeRefundRequest) (*AlipayTradeRefundResponse, error) {
	body, err := a.request("alipay.trade.refund", req)
	if err != nil {
		return nil, err
	}

	var response struct {
		AlipayTradeRefundResponse `json:"alipay_trade_refund_response"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.AlipayTradeRefundResponse, nil
}

// VerifyNotify 验证异步通知
func (a *Alipay) VerifyNotify(params map[string]string) bool {
	sign := params["sign"]
	delete(params, "sign")
	delete(params, "sign_type")

	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		if params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 构建待验证字符串
	var signStr strings.Builder
	for i, key := range keys {
		if i > 0 {
			signStr.WriteString("&")
		}
		signStr.WriteString(fmt.Sprintf("%s=%s", key, params[key]))
	}

	// 验证签名
	var hash crypto.Hash
	if a.config.SignType == "RSA2" {
		hash = crypto.SHA256
	} else {
		hash = crypto.SHA1
	}

	h := hash.New()
	h.Write([]byte(signStr.String()))
	hashed := h.Sum(nil)

	signature, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}

	err = rsa.VerifyPKCS1v15(a.publicKey, hash, hashed, signature)
	return err == nil
}
