package model

type AppConfig struct {
	Database struct {
		Dbname   string `json:"dbname"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Dialect  string `json:"dialect"`
		Pool     struct {
			Max      int `json:"max"`
			Lifetime int `json:"lifetime"`
			Idle     int `json:"idle"`
		} `json:"pool"`
	} `json:"database"`
	Redis struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Db       int    `json:"db"`
		Password string `json:"password"`
	} `json:"redis"`
	Server struct {
		Name  string `json:"name"`
		Port  int    `json:"port"`
		Debug bool   `json:"debug"`
	} `json:"server"`
}

type LoginUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	RoleId   int    `json:"role_id"`
	RoleName string `json:"role_name"`
	Memo     string `json:"memo"`
}

type PageFilter struct {
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// 条件组，支持 OR 查询
type FilterGroup struct {
	Logic   string                `json:"logic"`   // "and" 或 "or"，默认 "and"
	Filters map[string]PageFilter `json:"filters"` // 组内的条件
}

type Page struct {
	Page         int                   `json:"page"`
	PageSize     int                   `json:"page_size"`
	SortBy       string                `json:"sort_by"`
	Order        string                `json:"order"`
	TableName    string                `json:"table_name"`
	Filters      map[string]PageFilter `json:"filters"`       // 兼容原来的简单查询条件
	FilterGroups []FilterGroup         `json:"filter_groups"` // 支持复杂条件分组
}

// 定义post过来通用的数据保存结构，要有需要效验的字段
type SaveData struct {
	TableName string                 `json:"table_name"`
	Data      map[string]interface{} `json:"data"`
	Required  []string               `json:"required"`
	Optional  []string               `json:"optional"`
}
