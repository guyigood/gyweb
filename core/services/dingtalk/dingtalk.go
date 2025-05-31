package dingtalk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	AppKey       string `json:"app_key"`        // 应用的AppKey
	AppSecret    string `json:"app_secret"`     // 应用的AppSecret
	AgentID      int64  `json:"agent_id"`       // 应用的AgentId
	CorpID       string `json:"corp_id"`        // 企业的CorpId
	CorpSecret   string `json:"corp_secret"`    // 企业应用的CorpSecret
	RobotToken   string `json:"robot_token"`    // 机器人webhook token
	RobotSecret  string `json:"robot_secret"`   // 机器人加签秘钥
	IsOldVersion bool   `json:"is_old_version"` // 是否旧版API
}

// DingTalk 钉钉客户端
type DingTalk struct {
	config      *DingTalkConfig
	client      *http.Client
	accessToken string
	expiresAt   time.Time
}

// AccessTokenResponse 获取访问令牌响应
type AccessTokenResponse struct {
	ErrCode     int    `json:"errcode"`      // 返回码
	ErrMsg      string `json:"errmsg"`       // 对返回码的文本描述内容
	AccessToken string `json:"access_token"` // 获取到的凭证
	ExpiresIn   int    `json:"expires_in"`   // 凭证有效时间，单位：秒
}

// UserInfo 用户信息
type UserInfo struct {
	ErrCode      int    `json:"errcode"`      // 返回码
	ErrMsg       string `json:"errmsg"`       // 错误信息
	UserID       string `json:"userid"`       // 员工的userid
	Name         string `json:"name"`         // 员工名称
	Avatar       string `json:"avatar"`       // 头像url
	Mobile       string `json:"mobile"`       // 手机号
	Email        string `json:"email"`        // 邮箱
	UnionID      string `json:"unionid"`      // 员工在当前开发者企业账号范围内的唯一标识
	OpenID       string `json:"openid"`       // 员工在当前应用内的唯一标识
	StateCode    string `json:"state_code"`   // 手机号对应的国家号
	JobNumber    string `json:"job_number"`   // 员工工号
	Title        string `json:"title"`        // 职位
	WorkPlace    string `json:"work_place"`   // 办公地点
	Remark       string `json:"remark"`       // 备注
	DeptIDList   []int  `json:"dept_id_list"` // 员工部门id列表
	Extension    string `json:"extension"`    // 扩展属性
	HiredDate    int64  `json:"hired_date"`   // 入职时间
	Active       bool   `json:"active"`       // 表示该用户是否激活了钉钉
	Admin        bool   `json:"admin"`        // 是否为企业的管理员
	Boss         bool   `json:"boss"`         // 是否为企业的老板
	LeaderInDept []struct {
		DeptID   int  `json:"dept_id"`   // 部门id
		IsLeader bool `json:"is_leader"` // 是否为部门主管
	} `json:"leader_in_dept"` // 在对应的部门中是否为主管
	RoleList []struct {
		GroupName string `json:"group_name"` // 角色组名称
		Name      string `json:"name"`       // 角色名称
		ID        int    `json:"id"`         // 角色id
	} `json:"role_list"` // 角色列表
}

// Message 消息结构
type Message struct {
	AgentID   int64       `json:"agentid"`      // 应用agentId
	UserList  string      `json:"userid_list"`  // 接收者的userid列表
	DeptList  string      `json:"dept_id_list"` // 接收者的部门id列表
	ToAllUser bool        `json:"to_all_user"`  // 是否发送给企业全部用户
	Msg       interface{} `json:"msg"`          // 消息内容
}

// TextMessage 文本消息
type TextMessage struct {
	MsgType string `json:"msgtype"` // 消息类型，此时固定为：text
	Text    struct {
		Content string `json:"content"` // 消息内容，最长不超过2048个字节
	} `json:"text"`
}

// MarkdownMessage markdown消息
type MarkdownMessage struct {
	MsgType  string `json:"msgtype"` // 消息类型，此时固定为：markdown
	Markdown struct {
		Title string `json:"title"` // 首屏会话透出的展示内容
		Text  string `json:"text"`  // markdown格式的消息
	} `json:"markdown"`
}

// LinkMessage 链接消息
type LinkMessage struct {
	MsgType string `json:"msgtype"` // 消息类型，此时固定为：link
	Link    struct {
		MessageURL string `json:"messageUrl"` // 点击消息跳转的URL
		PicURL     string `json:"picUrl"`     // 图片URL
		Title      string `json:"title"`      // 消息标题
		Text       string `json:"text"`       // 消息内容。如果太长只会部分展示
	} `json:"link"`
}

// ActionCardMessage 卡片消息
type ActionCardMessage struct {
	MsgType    string `json:"msgtype"` // 消息类型，此时固定为：actionCard
	ActionCard struct {
		Title          string `json:"title"`          // 首屏会话透出的展示内容
		Text           string `json:"text"`           // markdown格式的消息
		SingleTitle    string `json:"singleTitle"`    // 单个按钮的方案
		SingleURL      string `json:"singleURL"`      // 点击singleTitle按钮触发的URL
		BtnOrientation string `json:"btnOrientation"` // 0-按钮竖直排列，1-按钮横向排列
		Btns           []struct {
			Title     string `json:"title"`     // 按钮方案
			ActionURL string `json:"actionURL"` // 点击按钮触发的URL
		} `json:"btns"` // 按钮
	} `json:"actionCard"`
}

// SendMessageResponse 发送消息响应
type SendMessageResponse struct {
	ErrCode int    `json:"errcode"` // 返回码
	ErrMsg  string `json:"errmsg"`  // 对返回码的文本描述内容
	TaskID  int64  `json:"task_id"` // 创建的异步发送任务id
}

// RobotMessage 机器人消息
type RobotMessage struct {
	MsgType    string         `json:"msgtype"` // 消息类型
	Text       *TextMsg       `json:"text,omitempty"`
	Link       *LinkMsg       `json:"link,omitempty"`
	Markdown   *MarkdownMsg   `json:"markdown,omitempty"`
	ActionCard *ActionCardMsg `json:"actionCard,omitempty"`
	At         *AtMsg         `json:"at,omitempty"`
}

// TextMsg 文本消息内容
type TextMsg struct {
	Content string `json:"content"`
}

// LinkMsg 链接消息内容
type LinkMsg struct {
	Text       string `json:"text"`
	Title      string `json:"title"`
	PicURL     string `json:"picUrl"`
	MessageURL string `json:"messageUrl"`
}

// MarkdownMsg markdown消息内容
type MarkdownMsg struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// ActionCardMsg 卡片消息内容
type ActionCardMsg struct {
	Title          string `json:"title"`
	Text           string `json:"text"`
	SingleTitle    string `json:"singleTitle,omitempty"`
	SingleURL      string `json:"singleURL,omitempty"`
	BtnOrientation string `json:"btnOrientation,omitempty"`
	Btns           []struct {
		Title     string `json:"title"`
		ActionURL string `json:"actionURL"`
	} `json:"btns,omitempty"`
}

// AtMsg @成员信息
type AtMsg struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	AtUserIds []string `json:"atUserIds,omitempty"`
	IsAtAll   bool     `json:"isAtAll,omitempty"`
}

// ApprovalProcessRequest 审批流程请求
type ApprovalProcessRequest struct {
	ProcessCode           string                   `json:"process_code"`            // 审批流的唯一码
	OriginatorUserID      string                   `json:"originator_user_id"`      // 审批实例发起人的userid
	DeptID                int                      `json:"dept_id"`                 // 发起人所在的部门
	Approvers             string                   `json:"approvers"`               // 审批人userid列表
	ApproversV2           []map[string]interface{} `json:"approvers_v2"`            // 审批人列表
	CcList                string                   `json:"cc_list"`                 // 抄送人userid列表
	CcPosition            string                   `json:"cc_position"`             // 抄送时间
	FormComponentValues   []FormComponentValue     `json:"form_component_values"`   // 审批流表单参数
	TargetSelectActioners string                   `json:"target_select_actioners"` // 指定审批人
	MicroappAgentID       int64                    `json:"microapp_agent_id"`       // 应用id
}

// FormComponentValue 表单组件值
type FormComponentValue struct {
	Name     string      `json:"name"`      // 表单每一栏的名称
	Value    interface{} `json:"value"`     // 表单每一栏的值
	ExtValue string      `json:"ext_value"` // 扩展值
}

// ApprovalProcessResponse 审批流程响应
type ApprovalProcessResponse struct {
	ErrCode           int    `json:"errcode"`             // 返回码
	ErrMsg            string `json:"errmsg"`              // 对返回码的文本描述内容
	ProcessInstanceID string `json:"process_instance_id"` // 审批实例id
}

// NewDingTalk 创建钉钉客户端
func NewDingTalk(config *DingTalkConfig) *DingTalk {
	return &DingTalk{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAccessToken 获取访问令牌
func (d *DingTalk) GetAccessToken() (string, error) {
	// 检查token是否过期
	if d.accessToken != "" && time.Now().Before(d.expiresAt) {
		return d.accessToken, nil
	}

	var apiURL string
	if d.config.IsOldVersion {
		// 旧版API
		apiURL = fmt.Sprintf("https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s",
			d.config.AppKey, d.config.AppSecret)
	} else {
		// 新版API
		apiURL = "https://oapi.dingtalk.com/gettoken"

		// 构建请求参数
		data := map[string]string{
			"appkey":    d.config.AppKey,
			"appsecret": d.config.AppSecret,
		}

		jsonData, _ := json.Marshal(data)
		resp, err := d.client.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
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
		d.accessToken = response.AccessToken
		d.expiresAt = time.Now().Add(time.Duration(response.ExpiresIn-300) * time.Second)

		return d.accessToken, nil
	}

	// 旧版API GET请求
	resp, err := d.client.Get(apiURL)
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
	d.accessToken = response.AccessToken
	d.expiresAt = time.Now().Add(time.Duration(response.ExpiresIn-300) * time.Second)

	return d.accessToken, nil
}

// GetUserInfo 获取用户详情
func (d *DingTalk) GetUserInfo(userID string) (*UserInfo, error) {
	accessToken, err := d.GetAccessToken()
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://oapi.dingtalk.com/topapi/v2/user/get?access_token=%s", accessToken)

	data := map[string]string{
		"userid": userID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		ErrCode int      `json:"errcode"`
		ErrMsg  string   `json:"errmsg"`
		Result  UserInfo `json:"result"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.ErrCode != 0 {
		return nil, fmt.Errorf("get user info error: %d %s", response.ErrCode, response.ErrMsg)
	}

	return &response.Result, nil
}

// SendWorkMessage 发送工作通知消息
func (d *DingTalk) SendWorkMessage(msg *Message) (*SendMessageResponse, error) {
	accessToken, err := d.GetAccessToken()
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://oapi.dingtalk.com/topapi/message/corpconversation/asyncsend_v2?access_token=%s", accessToken)

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response SendMessageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// SendRobotMessage 发送机器人消息
func (d *DingTalk) SendRobotMessage(msg *RobotMessage) error {
	if d.config.RobotToken == "" {
		return fmt.Errorf("robot token not configured")
	}

	// 计算签名
	timestamp := time.Now().UnixNano() / 1e6
	sign := d.calculateSign(timestamp, d.config.RobotSecret)

	// 构建webhook URL
	webhookURL := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s&timestamp=%d&sign=%s",
		d.config.RobotToken, timestamp, url.QueryEscape(sign))

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := d.client.Post(webhookURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	if response.ErrCode != 0 {
		return fmt.Errorf("send robot message error: %d %s", response.ErrCode, response.ErrMsg)
	}

	return nil
}

// calculateSign 计算签名
func (d *DingTalk) calculateSign(timestamp int64, secret string) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// CreateApprovalProcess 创建审批流程
func (d *DingTalk) CreateApprovalProcess(req *ApprovalProcessRequest) (*ApprovalProcessResponse, error) {
	accessToken, err := d.GetAccessToken()
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://oapi.dingtalk.com/topapi/processinstance/create?access_token=%s", accessToken)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response ApprovalProcessResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// SendTextMessage 发送文本消息（工作通知）
func (d *DingTalk) SendTextMessage(userList, deptList string, content string, toAllUser bool) (*SendMessageResponse, error) {
	msg := &Message{
		AgentID:   d.config.AgentID,
		UserList:  userList,
		DeptList:  deptList,
		ToAllUser: toAllUser,
		Msg: &TextMessage{
			MsgType: "text",
			Text: struct {
				Content string `json:"content"`
			}{
				Content: content,
			},
		},
	}

	return d.SendWorkMessage(msg)
}

// SendMarkdownMessage 发送Markdown消息（工作通知）
func (d *DingTalk) SendMarkdownMessage(userList, deptList, title, text string, toAllUser bool) (*SendMessageResponse, error) {
	msg := &Message{
		AgentID:   d.config.AgentID,
		UserList:  userList,
		DeptList:  deptList,
		ToAllUser: toAllUser,
		Msg: &MarkdownMessage{
			MsgType: "markdown",
			Markdown: struct {
				Title string `json:"title"`
				Text  string `json:"text"`
			}{
				Title: title,
				Text:  text,
			},
		},
	}

	return d.SendWorkMessage(msg)
}

// SendRobotTextMessage 发送机器人文本消息
func (d *DingTalk) SendRobotTextMessage(content string, atMobiles []string, atUserIds []string, isAtAll bool) error {
	msg := &RobotMessage{
		MsgType: "text",
		Text: &TextMsg{
			Content: content,
		},
		At: &AtMsg{
			AtMobiles: atMobiles,
			AtUserIds: atUserIds,
			IsAtAll:   isAtAll,
		},
	}

	return d.SendRobotMessage(msg)
}

// SendRobotMarkdownMessage 发送机器人Markdown消息
func (d *DingTalk) SendRobotMarkdownMessage(title, text string, atMobiles []string, atUserIds []string, isAtAll bool) error {
	msg := &RobotMessage{
		MsgType: "markdown",
		Markdown: &MarkdownMsg{
			Title: title,
			Text:  text,
		},
		At: &AtMsg{
			AtMobiles: atMobiles,
			AtUserIds: atUserIds,
			IsAtAll:   isAtAll,
		},
	}

	return d.SendRobotMessage(msg)
}

// GetUserByMobile 根据手机号获取用户信息
func (d *DingTalk) GetUserByMobile(mobile string) (*UserInfo, error) {
	accessToken, err := d.GetAccessToken()
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://oapi.dingtalk.com/topapi/v2/user/getbymobile?access_token=%s", accessToken)

	data := map[string]string{
		"mobile": mobile,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		ErrCode int      `json:"errcode"`
		ErrMsg  string   `json:"errmsg"`
		Result  UserInfo `json:"result"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.ErrCode != 0 {
		return nil, fmt.Errorf("get user by mobile error: %d %s", response.ErrCode, response.ErrMsg)
	}

	return &response.Result, nil
}
