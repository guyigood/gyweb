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

		LogicDeleteField string      `json:"logic-delete-field"`
		LogicDeleteValue interface{} `json:"logic-delete-value"`
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
	MQTT struct {
		Broker   string   `json:"broker"`
		Port     int      `json:"port"`
		ClientID string   `json:"client_id"`
		Username string   `json:"username"`
		Password string   `json:"password"`
		Topics   []string `json:"topics"`
	} `json:"mqtt"`
}

type LoginUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	RoleId   int    `json:"role_id"`
	RoleName string `json:"role_name"`
	Memo     string `json:"memo"`
}

// 定义post过来通用的数据保存结构，要有需要效验的字段

// TableInfo 数据库表信息结构
type TableInfo struct {
	TableName    string `json:"table_name"`
	TableComment string `json:"table_comment"`
	TableSchema  string `json:"table_schema"`
}

// FieldInfo 数据库字段信息结构
type FieldInfo struct {
	TableName    string      `json:"table_name"`
	FieldName    string      `json:"field_name"`
	FieldType    string      `json:"field_type"`
	FieldLength  interface{} `json:"field_length"`
	DefaultValue interface{} `json:"default_value"`
	FieldComment string      `json:"field_comment"`
	IsNullable   string      `json:"is_nullable"`
	ColumnKey    string      `json:"column_key"`
	Extra        string      `json:"extra"`
	OrdinalPos   int         `json:"ordinal_position"`
}

type GLobalTbInfo struct {
	TableName  string         `json:"table_name"`
	ModuleName string         `json:"module_name"`
	JoinTable  string         `json:"join_table"`
	JoinField  string         `json:"join_field"`
	FdInfo     []GLobalFdInfo `json:"fd_info"`
	PrimaryKey string         `json:"primary_key"`
}

type GLobalFdInfo struct {
	FieldName    string `json:"field_name"`
	FieldType    string `json:"field_type"`
	IsSearchable bool   `json:"is_searchable"`
	IsRequired   bool   `json:"is_required"`
	IsUnique     bool   `json:"is_unique"`
	IsActive     bool   `json:"is_active"`
	QueryType    string `json:"query_type"`
	IsPk         bool   `json:"is_pk"`
}
