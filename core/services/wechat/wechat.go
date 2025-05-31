package wechat

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// WechatConfig 微信公众号配置
type WechatConfig struct {
	AppID          string `json:"app_id"`           // 应用ID
	AppSecret      string `json:"app_secret"`       // 应用密钥
	Token          string `json:"token"`            // 消息校验Token
	EncodingAESKey string `json:"encoding_aes_key"` // 消息加解密密钥
}

// Wechat 微信公众号客户端
type Wechat struct {
	config      *WechatConfig
	client      *http.Client
	accessToken string
	expiresAt   time.Time
}

// AccessTokenResponse 获取访问令牌响应
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"` // 获取到的凭证
	ExpiresIn   int    `json:"expires_in"`   // 凭证有效时间，单位：秒
	ErrCode     int    `json:"errcode"`      // 错误码
	ErrMsg      string `json:"errmsg"`       // 错误信息
}

// UserInfo 用户信息
type UserInfo struct {
	Subscribe      int    `json:"subscribe"`       // 是否关注公众号
	OpenID         string `json:"openid"`          // 用户的标识，对当前公众号唯一
	Nickname       string `json:"nickname"`        // 用户的昵称
	Sex            int    `json:"sex"`             // 用户的性别，值为1时是男性，值为2时是女性，值为0时是未知
	City           string `json:"city"`            // 用户所在城市
	Country        string `json:"country"`         // 用户所在国家
	Province       string `json:"province"`        // 用户所在省份
	Language       string `json:"language"`        // 用户的语言，简体中文为zh_CN
	Headimgurl     string `json:"headimgurl"`      // 用户头像
	SubscribeTime  int64  `json:"subscribe_time"`  // 用户关注时间
	UnionID        string `json:"unionid"`         // 只有在用户将公众号绑定到微信开放平台帐号后，才会出现该字段
	Remark         string `json:"remark"`          // 公众号运营者对粉丝的备注，公众号运营者可在微信公众平台用户管理界面对粉丝添加备注
	GroupID        int    `json:"groupid"`         // 用户所在的分组ID
	TagidList      []int  `json:"tagid_list"`      // 用户被打上的标签ID列表
	SubscribeScene string `json:"subscribe_scene"` // 返回用户关注的渠道来源
	QrScene        int    `json:"qr_scene"`        // 二维码扫码场景（开发者自定义）
	QrSceneStr     string `json:"qr_scene_str"`    // 二维码扫码场景描述（开发者自定义）
	ErrCode        int    `json:"errcode"`         // 错误码
	ErrMsg         string `json:"errmsg"`          // 错误信息
}

// UserList 用户列表
type UserList struct {
	Total int `json:"total"` // 关注该公众账号的总用户数
	Count int `json:"count"` // 拉取的OPENID个数，最大值为10000
	Data  struct {
		OpenID []string `json:"openid"` // 列表数据，OPENID的列表
	} `json:"data"`
	NextOpenID string `json:"next_openid"` // 拉取列表的第一个用户的OPENID，不填默认从头开始拉取
	ErrCode    int    `json:"errcode"`     // 错误码
	ErrMsg     string `json:"errmsg"`      // 错误信息
}

// TemplateMessage 模板消息
type TemplateMessage struct {
	ToUser      string                 `json:"touser"`      // 接收者openid
	TemplateID  string                 `json:"template_id"` // 模板ID
	URL         string                 `json:"url"`         // 模板跳转链接（海外帐号没有跳转能力）
	MiniProgram *MiniProgram           `json:"miniprogram"` // 跳小程序所需数据，不需跳小程序可不用传该数据
	Data        map[string]interface{} `json:"data"`        // 模板数据
}

// MiniProgram 小程序信息
type MiniProgram struct {
	AppID    string `json:"appid"`    // 所需跳转到的小程序appid（该小程序appid必须与发模板消息的公众号是绑定关联关系，暂不支持小游戏）
	PagePath string `json:"pagepath"` // 所需跳转到小程序的具体页面路径，支持带参数,（示例index?foo=bar），要求该小程序已发布，暂不支持小游戏
}

// TemplateMessageResponse 模板消息发送响应
type TemplateMessageResponse struct {
	ErrCode int    `json:"errcode"` // 错误码
	ErrMsg  string `json:"errmsg"`  // 错误信息
	MsgID   int64  `json:"msgid"`   // 消息id
}

// CustomerServiceMessage 客服消息
type CustomerServiceMessage struct {
	ToUser    string     `json:"touser"`               // 普通用户openid
	MsgType   string     `json:"msgtype"`              // 消息类型
	Text      *TextMsg   `json:"text,omitempty"`       // 文本消息
	Image     *MediaMsg  `json:"image,omitempty"`      // 图片消息
	Voice     *MediaMsg  `json:"voice,omitempty"`      // 语音消息
	Video     *VideoMsg  `json:"video,omitempty"`      // 视频消息
	Music     *MusicMsg  `json:"music,omitempty"`      // 音乐消息
	News      *NewsMsg   `json:"news,omitempty"`       // 图文消息
	MpNews    *MpNewsMsg `json:"mpnews,omitempty"`     // 图文消息（点击跳转到图文消息页面）
	Card      *CardMsg   `json:"wxcard,omitempty"`     // 卡券
	KfAccount string     `json:"kf_account,omitempty"` // 指定特定的客服账号
}

// TextMsg 文本消息
type TextMsg struct {
	Content string `json:"content"` // 文本消息内容
}

// MediaMsg 媒体消息
type MediaMsg struct {
	MediaID string `json:"media_id"` // 媒体文件ID
}

// VideoMsg 视频消息
type VideoMsg struct {
	MediaID      string `json:"media_id"`       // 媒体文件ID
	ThumbMediaID string `json:"thumb_media_id"` // 缩略图的媒体ID
	Title        string `json:"title"`          // 视频消息的标题
	Description  string `json:"description"`    // 视频消息的描述
}

// MusicMsg 音乐消息
type MusicMsg struct {
	Title        string `json:"title"`          // 音乐标题
	Description  string `json:"description"`    // 音乐描述
	MusicURL     string `json:"musicurl"`       // 音乐链接
	HQMusicURL   string `json:"hqmusicurl"`     // 高质量音乐链接，WIFI环境优先使用该链接播放音乐
	ThumbMediaID string `json:"thumb_media_id"` // 缩略图的媒体ID
}

// NewsMsg 图文消息
type NewsMsg struct {
	Articles []Article `json:"articles"` // 图文消息，一个图文消息支持1到8条图文
}

// Article 图文消息文章
type Article struct {
	Title       string `json:"title"`       // 图文消息标题
	Description string `json:"description"` // 图文消息描述
	URL         string `json:"url"`         // 点击图文消息跳转链接
	PicURL      string `json:"picurl"`      // 图片链接，支持JPG、PNG格式，较好的效果为大图360*200，小图200*200
}

// MpNewsMsg 图文消息（点击跳转到图文消息页面）
type MpNewsMsg struct {
	MediaID string `json:"media_id"` // 媒体文件ID
}

// CardMsg 卡券消息
type CardMsg struct {
	CardID string `json:"card_id"` // 卡券ID
}

// QRCodeRequest 二维码生成请求
type QRCodeRequest struct {
	ExpireSeconds int         `json:"expire_seconds,omitempty"` // 该二维码有效时间，以秒为单位。 最大不超过2592000（即30天），此字段如果不填，则默认有效期是30秒
	ActionName    string      `json:"action_name"`              // 二维码类型，QR_SCENE为临时的整型参数值，QR_STR_SCENE为临时的字符串参数值，QR_LIMIT_SCENE为永久的整型参数值，QR_LIMIT_STR_SCENE为永久的字符串参数值
	ActionInfo    *ActionInfo `json:"action_info"`              // 二维码详细信息
}

// ActionInfo 二维码详细信息
type ActionInfo struct {
	Scene *Scene `json:"scene"` // 场景值ID
}

// Scene 场景值
type Scene struct {
	SceneID  int    `json:"scene_id,omitempty"`  // 场景值ID，临时二维码时为32位非0整型，永久二维码时最大值为100000（目前参数只支持1--100000）
	SceneStr string `json:"scene_str,omitempty"` // 场景值ID（字符串形式的ID），字符串类型，长度限制为1到64
}

// QRCodeResponse 二维码生成响应
type QRCodeResponse struct {
	Ticket        string `json:"ticket"`         // 获取的二维码ticket，凭借此ticket可以在有效时间内换取二维码
	ExpireSeconds int    `json:"expire_seconds"` // 该二维码有效时间，以秒为单位。 最大不超过2592000（即30天）
	URL           string `json:"url"`            // 二维码图片解析后的地址，开发者可根据该地址自行生成需要的二维码图片
	ErrCode       int    `json:"errcode"`        // 错误码
	ErrMsg        string `json:"errmsg"`         // 错误信息
}

// Menu 自定义菜单
type Menu struct {
	Button []Button `json:"button"` // 一级菜单数组，个数应为1~3个
}

// Button 菜单按钮
type Button struct {
	Type      string   `json:"type,omitempty"`       // 菜单的响应动作类型，view表示网页类型，click表示点击类型，miniprogram表示小程序类型
	Name      string   `json:"name"`                 // 菜单标题，不超过16个字节，子菜单不超过60个字节
	Key       string   `json:"key,omitempty"`        // 菜单KEY值，用于消息接口推送，不超过128字节
	URL       string   `json:"url,omitempty"`        // 网页链接，用户点击菜单可打开链接，不超过1024字节。type为miniprogram时，不支持小程序的老版本客户端等
	AppID     string   `json:"appid,omitempty"`      // 小程序的appid（仅认证公众号可配置）
	PagePath  string   `json:"pagepath,omitempty"`   // 小程序的页面路径
	SubButton []Button `json:"sub_button,omitempty"` // 二级菜单数组，个数应为1~5个
}

// MenuResponse 菜单响应
type MenuResponse struct {
	ErrCode int    `json:"errcode"` // 错误码
	ErrMsg  string `json:"errmsg"`  // 错误信息
}

// OAuthAccessTokenResponse 网页授权access_token响应
type OAuthAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`  // 网页授权接口调用凭证
	ExpiresIn    int    `json:"expires_in"`    // access_token接口调用凭证超时时间，单位（秒）
	RefreshToken string `json:"refresh_token"` // 用户刷新access_token
	OpenID       string `json:"openid"`        // 用户唯一标识
	Scope        string `json:"scope"`         // 用户授权的作用域，使用逗号（,）分隔
	ErrCode      int    `json:"errcode"`       // 错误码
	ErrMsg       string `json:"errmsg"`        // 错误信息
}

// OAuthUserInfo 网页授权用户信息
type OAuthUserInfo struct {
	OpenID     string   `json:"openid"`     // 用户的唯一标识
	Nickname   string   `json:"nickname"`   // 用户昵称
	Sex        int      `json:"sex"`        // 用户的性别，值为1时是男性，值为2时是女性，值为0时是未知
	Province   string   `json:"province"`   // 用户个人资料填写的省份
	City       string   `json:"city"`       // 普通用户个人资料填写的城市
	Country    string   `json:"country"`    // 国家，如中国为CN
	Headimgurl string   `json:"headimgurl"` // 用户头像，最后一个数值代表正方形头像大小（有0、46、64、96、132数值可选，0代表640*640正方形头像）
	Privilege  []string `json:"privilege"`  // 用户特权信息，json 数组，如微信沃卡用户为（chinaunicom）
	UnionID    string   `json:"unionid"`    // 只有在用户将公众号绑定到微信开放平台帐号后，才会出现该字段
	ErrCode    int      `json:"errcode"`    // 错误码
	ErrMsg     string   `json:"errmsg"`     // 错误信息
}

// NewWechat 创建微信公众号客户端
func NewWechat(config *WechatConfig) *Wechat {
	return &Wechat{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAccessToken 获取访问令牌
func (w *Wechat) GetAccessToken() (string, error) {
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

// GetUserInfo 获取用户信息
func (w *Wechat) GetUserInfo(openID string) (*UserInfo, error) {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN",
		accessToken, openID)

	resp, err := w.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo UserInfo
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return nil, err
	}

	if userInfo.ErrCode != 0 {
		return nil, fmt.Errorf("get user info error: %d %s", userInfo.ErrCode, userInfo.ErrMsg)
	}

	return &userInfo, nil
}

// GetUserList 获取用户列表
func (w *Wechat) GetUserList(nextOpenID string) (*UserList, error) {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/user/get?access_token=%s&next_openid=%s",
		accessToken, nextOpenID)

	resp, err := w.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userList UserList
	err = json.Unmarshal(body, &userList)
	if err != nil {
		return nil, err
	}

	if userList.ErrCode != 0 {
		return nil, fmt.Errorf("get user list error: %d %s", userList.ErrCode, userList.ErrMsg)
	}

	return &userList, nil
}

// SendTemplateMessage 发送模板消息
func (w *Wechat) SendTemplateMessage(msg *TemplateMessage) (*TemplateMessageResponse, error) {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s", accessToken)

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

// SendCustomerServiceMessage 发送客服消息
func (w *Wechat) SendCustomerServiceMessage(msg *CustomerServiceMessage) (*MenuResponse, error) {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s", accessToken)

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

	var response MenuResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// CreateQRCode 生成二维码
func (w *Wechat) CreateQRCode(req *QRCodeRequest) (*QRCodeResponse, error) {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=%s", accessToken)

	jsonData, err := json.Marshal(req)
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

	var response QRCodeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.ErrCode != 0 {
		return nil, fmt.Errorf("create qrcode error: %d %s", response.ErrCode, response.ErrMsg)
	}

	return &response, nil
}

// GetQRCodeImage 获取二维码图片
func (w *Wechat) GetQRCodeImage(ticket string) ([]byte, error) {
	url := fmt.Sprintf("https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=%s", url.QueryEscape(ticket))

	resp, err := w.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// CreateMenu 创建自定义菜单
func (w *Wechat) CreateMenu(menu *Menu) (*MenuResponse, error) {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/menu/create?access_token=%s", accessToken)

	jsonData, err := json.Marshal(menu)
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

	var response MenuResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// DeleteMenu 删除自定义菜单
func (w *Wechat) DeleteMenu() (*MenuResponse, error) {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/menu/delete?access_token=%s", accessToken)

	resp, err := w.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response MenuResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetOAuthURL 获取网页授权链接
func (w *Wechat) GetOAuthURL(redirectURI, state string, scope string) string {
	if scope == "" {
		scope = "snsapi_base"
	}
	return fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect",
		w.config.AppID, url.QueryEscape(redirectURI), scope, state)
}

// GetOAuthAccessToken 通过code换取网页授权access_token
func (w *Wechat) GetOAuthAccessToken(code string) (*OAuthAccessTokenResponse, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		w.config.AppID, w.config.AppSecret, code)

	resp, err := w.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response OAuthAccessTokenResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.ErrCode != 0 {
		return nil, fmt.Errorf("get oauth access token error: %d %s", response.ErrCode, response.ErrMsg)
	}

	return &response, nil
}

// GetOAuthUserInfo 拉取用户信息(需scope为 snsapi_userinfo)
func (w *Wechat) GetOAuthUserInfo(accessToken, openID string) (*OAuthUserInfo, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		accessToken, openID)

	resp, err := w.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo OAuthUserInfo
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return nil, err
	}

	if userInfo.ErrCode != 0 {
		return nil, fmt.Errorf("get oauth user info error: %d %s", userInfo.ErrCode, userInfo.ErrMsg)
	}

	return &userInfo, nil
}

// VerifySignature 验证微信服务器签名
func (w *Wechat) VerifySignature(signature, timestamp, nonce string) bool {
	// 1. 将token、timestamp、nonce三个参数进行字典序排序
	params := []string{w.config.Token, timestamp, nonce}
	sort.Strings(params)

	// 2. 将三个参数字符串拼接成一个字符串进行sha1加密
	str := strings.Join(params, "")
	h := sha1.New()
	h.Write([]byte(str))
	encrypted := fmt.Sprintf("%x", h.Sum(nil))

	// 3. 开发者获得加密后的字符串可与signature对比，标识该请求来源于微信
	return encrypted == signature
}

// SendTextMessage 发送文本客服消息
func (w *Wechat) SendTextMessage(openID, content string) (*MenuResponse, error) {
	msg := &CustomerServiceMessage{
		ToUser:  openID,
		MsgType: "text",
		Text: &TextMsg{
			Content: content,
		},
	}
	return w.SendCustomerServiceMessage(msg)
}

// SendImageMessage 发送图片客服消息
func (w *Wechat) SendImageMessage(openID, mediaID string) (*MenuResponse, error) {
	msg := &CustomerServiceMessage{
		ToUser:  openID,
		MsgType: "image",
		Image: &MediaMsg{
			MediaID: mediaID,
		},
	}
	return w.SendCustomerServiceMessage(msg)
}

// SendNewsMessage 发送图文客服消息
func (w *Wechat) SendNewsMessage(openID string, articles []Article) (*MenuResponse, error) {
	msg := &CustomerServiceMessage{
		ToUser:  openID,
		MsgType: "news",
		News: &NewsMsg{
			Articles: articles,
		},
	}
	return w.SendCustomerServiceMessage(msg)
}
