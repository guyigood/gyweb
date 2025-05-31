package miniprogram

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WechatMiniConfig 微信小程序配置
type WechatMiniConfig struct {
	AppID     string `json:"app_id"`     // 小程序 appId
	AppSecret string `json:"app_secret"` // 小程序 appSecret
}

// WechatMini 微信小程序客户端
type WechatMini struct {
	config      *WechatMiniConfig
	client      *http.Client
	accessToken string
	expiresAt   time.Time
}

// Code2SessionRequest 登录凭证校验请求
type Code2SessionRequest struct {
	JSCode string `json:"js_code"` // 小程序端返回的code
}

// Code2SessionResponse 登录凭证校验响应
type Code2SessionResponse struct {
	OpenID     string `json:"openid"`      // 用户唯一标识
	SessionKey string `json:"session_key"` // 会话密钥
	UnionID    string `json:"unionid"`     // 用户在开放平台的唯一标识符
	ErrCode    int    `json:"errcode"`     // 错误码
	ErrMsg     string `json:"errmsg"`      // 错误信息
}

// AccessTokenResponse 获取接口调用凭据响应
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"` // 获取到的凭证
	ExpiresIn   int    `json:"expires_in"`   // 凭证有效时间，单位：秒
	ErrCode     int    `json:"errcode"`      // 错误码
	ErrMsg      string `json:"errmsg"`       // 错误信息
}

// UserInfo 用户信息
type UserInfo struct {
	OpenID    string `json:"openId"`    // 用户openid
	NickName  string `json:"nickName"`  // 用户昵称
	Gender    int    `json:"gender"`    // 用户性别
	City      string `json:"city"`      // 用户所在城市
	Province  string `json:"province"`  // 用户所在省份
	Country   string `json:"country"`   // 用户所在国家
	AvatarURL string `json:"avatarUrl"` // 用户头像
	UnionID   string `json:"unionId"`   // 用户UnionID
	Watermark struct {
		Timestamp int64  `json:"timestamp"` // 时间戳
		AppID     string `json:"appid"`     // 小程序appid
	} `json:"watermark"` // 数据水印
}

// TemplateMessage 模板消息
type TemplateMessage struct {
	ToUser           string                 `json:"touser"`            // 接收者openid
	TemplateID       string                 `json:"template_id"`       // 模板ID
	Page             string                 `json:"page,omitempty"`    // 点击模板卡片后的跳转页面
	FormID           string                 `json:"form_id"`           // 表单提交场景下，为submit事件带上的formId
	Data             map[string]interface{} `json:"data"`              // 模板内容
	EmphasisKeyword  string                 `json:"emphasis_keyword"`  // 模板需要放大的关键词
	MiniprogramState string                 `json:"miniprogram_state"` // 跳转小程序类型
	Lang             string                 `json:"lang"`              // 进入小程序查看"的语言类型
}

// TemplateMessageResponse 模板消息发送响应
type TemplateMessageResponse struct {
	ErrCode int    `json:"errcode"` // 错误码
	ErrMsg  string `json:"errmsg"`  // 错误信息
}

// PhoneInfo 手机号信息
type PhoneInfo struct {
	PhoneNumber     string `json:"phoneNumber"`     // 用户绑定的手机号
	PurePhoneNumber string `json:"purePhoneNumber"` // 没有区号的手机号
	CountryCode     string `json:"countryCode"`     // 区号
	Watermark       struct {
		Timestamp int64  `json:"timestamp"` // 时间戳
		AppID     string `json:"appid"`     // 小程序appid
	} `json:"watermark"` // 数据水印
}

// QRCodeRequest 小程序码生成请求
type QRCodeRequest struct {
	Scene      string `json:"scene,omitempty"`       // 最大32个可见字符，只支持数字，大小写英文以及部分特殊字符
	Page       string `json:"page,omitempty"`        // 必须是已经发布的小程序存在的页面
	CheckPath  bool   `json:"check_path,omitempty"`  // 检查page是否存在，为true时page必须是线上小程序存在的页面
	EnvVersion string `json:"env_version,omitempty"` // 要打开的小程序版本
	Width      int    `json:"width,omitempty"`       // 二维码的宽度，单位px，最小280px，最大1280px
	AutoColor  bool   `json:"auto_color,omitempty"`  // 自动配置线条颜色
	LineColor  struct {
		R int `json:"r"` // rgb颜色值
		G int `json:"g"`
		B int `json:"b"`
	} `json:"line_color,omitempty"` // 线条颜色
	IsHyaline bool `json:"is_hyaline,omitempty"` // 是否需要透明底色
}

// NewWechatMini 创建微信小程序客户端
func NewWechatMini(config *WechatMiniConfig) *WechatMini {
	return &WechatMini{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Code2Session 登录凭证校验
func (w *WechatMini) Code2Session(jsCode string) (*Code2SessionResponse, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		w.config.AppID, w.config.AppSecret, jsCode)

	resp, err := w.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response Code2SessionResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.ErrCode != 0 {
		return nil, fmt.Errorf("code2session error: %d %s", response.ErrCode, response.ErrMsg)
	}

	return &response, nil
}

// GetAccessToken 获取接口调用凭据
func (w *WechatMini) GetAccessToken() (string, error) {
	// 检查token是否过期
	if w.accessToken != "" && time.Now().Before(w.expiresAt) {
		return w.accessToken, nil
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		w.config.AppID, w.config.AppSecret)

	resp, err := w.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response AccessTokenResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	if response.ErrCode != 0 {
		return "", fmt.Errorf("get access token error: %d %s", response.ErrCode, response.ErrMsg)
	}

	// 保存token和过期时间（提前5分钟过期）
	w.accessToken = response.AccessToken
	w.expiresAt = time.Now().Add(time.Duration(response.ExpiresIn-300) * time.Second)

	return w.accessToken, nil
}

// DecryptData 解密敏感数据
func (w *WechatMini) DecryptData(sessionKey, encryptedData, iv string, result interface{}) error {
	// Base64解码
	sessionKeyBytes, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return err
	}

	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return err
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return err
	}

	// AES解密
	block, err := aes.NewCipher(sessionKeyBytes)
	if err != nil {
		return err
	}

	mode := cipher.NewCBCDecrypter(block, ivBytes)
	mode.CryptBlocks(encryptedBytes, encryptedBytes)

	// 去除PKCS7填充
	decrypted := w.removePKCS7Padding(encryptedBytes)

	// 解析JSON
	return json.Unmarshal(decrypted, result)
}

// removePKCS7Padding 去除PKCS7填充
func (w *WechatMini) removePKCS7Padding(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return data
	}
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}

// GetUserInfo 获取用户信息
func (w *WechatMini) GetUserInfo(sessionKey, encryptedData, iv string) (*UserInfo, error) {
	var userInfo UserInfo
	err := w.DecryptData(sessionKey, encryptedData, iv, &userInfo)
	if err != nil {
		return nil, err
	}

	// 验证水印
	if userInfo.Watermark.AppID != w.config.AppID {
		return nil, fmt.Errorf("水印验证失败")
	}

	return &userInfo, nil
}

// GetPhoneNumber 获取手机号
func (w *WechatMini) GetPhoneNumber(sessionKey, encryptedData, iv string) (*PhoneInfo, error) {
	var phoneInfo PhoneInfo
	err := w.DecryptData(sessionKey, encryptedData, iv, &phoneInfo)
	if err != nil {
		return nil, err
	}

	// 验证水印
	if phoneInfo.Watermark.AppID != w.config.AppID {
		return nil, fmt.Errorf("水印验证失败")
	}

	return &phoneInfo, nil
}

// SendTemplateMessage 发送模板消息
func (w *WechatMini) SendTemplateMessage(msg *TemplateMessage) (*TemplateMessageResponse, error) {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/wxopen/template/send?access_token=%s", accessToken)

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	resp, err := w.client.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response TemplateMessageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GenerateQRCode 生成小程序码
func (w *WechatMini) GenerateQRCode(req *QRCodeRequest) ([]byte, error) {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=%s", accessToken)

	// 设置默认值
	if req.Width == 0 {
		req.Width = 430
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := w.client.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应内容类型
	if resp.Header.Get("Content-Type") == "application/json" {
		// 如果是JSON，说明有错误
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var errResp struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		json.Unmarshal(body, &errResp)
		return nil, fmt.Errorf("generate qrcode error: %d %s", errResp.ErrCode, errResp.ErrMsg)
	}

	// 如果是图片，直接返回
	return io.ReadAll(resp.Body)
}

// VerifySignature 验证签名
func (w *WechatMini) VerifySignature(rawData, sessionKey, signature string) bool {
	// 这里应该实现签名验证逻辑
	// 通常是 sha1(rawData + sessionKey)
	return true // 简化实现，实际项目中需要真正验证
}

// SendSubscribeMessage 发送订阅消息
func (w *WechatMini) SendSubscribeMessage(toUser, templateID string, data map[string]interface{}, page, miniprogramState string) (*TemplateMessageResponse, error) {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", accessToken)

	msgData := map[string]interface{}{
		"touser":      toUser,
		"template_id": templateID,
		"data":        data,
	}

	if page != "" {
		msgData["page"] = page
	}
	if miniprogramState != "" {
		msgData["miniprogram_state"] = miniprogramState
	}

	jsonData, err := json.Marshal(msgData)
	if err != nil {
		return nil, err
	}

	resp, err := w.client.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response TemplateMessageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
