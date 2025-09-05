package public

import (
	"fmt"
	"reflect"
	"sync"
	"{project_name}/model"
	"time"

	"github.com/guyigood/gyweb/core/utils/datatype"

	"github.com/guyigood/gyweb/core/middleware"
	orm "github.com/guyigood/gyweb/core/orm/mysql"
	"github.com/guyigood/gyweb/core/utils/common"
	"github.com/guyigood/gyweb/core/utils/redisutils"
)

var (
	SysConfig model.AppConfig
	Re_Client *redisutils.RedisClient
	Tbinfo    []model.GLobalTbInfo
	//DB        *orm.DB
	dbPool *DBPool // 数据库连接池

)

// DBPool 数据库连接池结构
type DBPool struct {
	pool    chan *orm.DB
	mutex   sync.Mutex
	maxSize int
	created int
}

// NewDBPool 创建数据库连接池
func NewDBPool(maxSize int) *DBPool {
	return &DBPool{
		pool:    make(chan *orm.DB, maxSize),
		maxSize: maxSize,
		created: 0,
	}
}

// GetConnection 从连接池获取数据库连接
func (p *DBPool) GetConnection() *orm.DB {
	select {
	case db := <-p.pool:
		// 从池中获取连接
		return db
	default:
		// 池中没有可用连接，创建新连接
		p.mutex.Lock()
		defer p.mutex.Unlock()

		if p.created < p.maxSize {
			db := p.createConnection()
			if db != nil {
				p.created++
				return db
			}
		}

		// 如果达到最大连接数，等待可用连接
		return <-p.pool
	}
}

// ReturnConnection 将连接归还到连接池
func (p *DBPool) ReturnConnection(db *orm.DB) {
	if db == nil {
		return
	}

	select {
	case p.pool <- db:
		// 成功归还到池中
	default:
		// 池已满，关闭连接
		// 注意：这里不实际关闭，因为orm.DB可能没有Close方法
		// 实际使用中可能需要根据具体的ORM实现来处理
	}
}

// createConnection 创建新的数据库连接
func (p *DBPool) createConnection() *orm.DB {
	dsn := SysConfig.Database.Username + ":" + SysConfig.Database.Password + "@tcp(" + SysConfig.Database.Host + ":" + datatype.TypetoStr(SysConfig.Database.Port) + ")/" + SysConfig.Database.Dbname + "?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := orm.NewDB(SysConfig.Database.Dialect, dsn)
	if err != nil {
		fmt.Printf("创建数据库连接失败: %v\n", err)
		return nil
	}

	db.SetMaxIdleConns(SysConfig.Database.Pool.Max)
	db.SetMaxOpenConns(SysConfig.Database.Pool.Idle)
	db.SetConnMaxLifetime(time.Duration(SysConfig.Database.Pool.Lifetime) * time.Second)

	return db
}

// DBConnection 数据库连接包装器，用于自动归还连接
type DBConnection struct {
	db   *orm.DB
	pool *DBPool
}

// GetDB 获取数据库连接（自动管理连接池）
func (conn *DBConnection) GetDB() *orm.DB {
	return conn.db
}

// Close 归还连接到池中
func (conn *DBConnection) Close() {
	if conn.pool != nil && conn.db != nil {
		conn.pool.ReturnConnection(conn.db)
	}
}

// Table 包装数据库表操作，确保每次调用都是独立的
func (conn *DBConnection) Table(tableName string) *orm.DB {
	return conn.db.Table(tableName)
}

// 初始化连接池
func initDBPool() {
	if dbPool == nil {
		dbPool = NewDBPool(SysConfig.Database.Pool.Max) // 创建最大20个连接的连接池
	}
}

// GetDbConnection 获取数据库连接（新的推荐方式）
func GetDbConnection() *DBConnection {
	if dbPool == nil {
		initDBPool()
	}

	db := dbPool.GetConnection()
	return &DBConnection{
		db:   db,
		pool: dbPool,
	}
}

const (
	PublicKey  = "04298364ec840088475eae92a591e01284d1abefcda348b47eb324bb521bb03b0b2a5bc393f6b71dabb8f15c99a0050818b56b23f31743b93df9cf8948f15ddb54"
	PrivateKey = "3037723d47292171677ec8bd7dc9af696c7472bc5f251b2cec07e65fdef22e25"
)

func SysInit() {
	err := common.ReadJsonFile("conf/config.json", &SysConfig)
	if err != nil {
		panic(err)
		return
	}

	redisUrl := fmt.Sprintf("%s:%d", SysConfig.Redis.Host, SysConfig.Redis.Port)
	Re_Client, err = redisutils.NewRedisClient(redisUrl, SysConfig.Redis.Password, SysConfig.Redis.Db)
	if err != nil {
		panic(err)
		return
	}
	err1 := Re_Client.Ping()
	if err != nil {
		fmt.Println(err1)
	}
	fmt.Println(redisUrl)
	GetTbInfo()
}

/*func GetNewDb() *orm.DB {

	dsn := SysConfig.Database.Username + ":" + SysConfig.Database.Password + "@tcp(" + SysConfig.Database.Host + ":" + datatype.TypetoStr(SysConfig.Database.Port) + ")/" + SysConfig.Database.Dbname + "?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	Db, err := orm.NewDB(SysConfig.Database.Dialect, dsn)
	if err != nil {
		panic(err)
	}
	Db.SetMaxIdleConns(SysConfig.Database.Pool.Max)
	Db.SetMaxOpenConns(SysConfig.Database.Pool.Idle)
	Db.SetConnMaxLifetime(time.Duration(SysConfig.Database.Pool.Lifetime) * time.Second)
	return Db
}*/

func GetJwtConfig() *middleware.JWTConfig {
	return &middleware.JWTConfig{
		SecretKey:     "gy7210",               // JWT密钥
		TokenLookup:   "header:Authorization", // 从请求头获取token
		TokenHeadName: "Bearer",               // token前缀
		ExpiresIn:     2 * 24 * time.Hour,     // token过期时间

	}
}

func StructToMap(data interface{}, tag string) map[string]interface{} {
	//结构体转map
	result := make(map[string]interface{})

	// 如果传入的是 nil，直接返回空 map
	if data == nil {
		return result
	}

	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	// 如果是指针，需要解引用
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return result
		}
		v = v.Elem()
		t = t.Elem()
	}

	// 只处理结构体类型
	if v.Kind() != reflect.Struct {
		return result
	}

	// 遍历结构体的所有字段
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 跳过私有字段（首字母小写）
		if !fieldType.IsExported() {
			continue
		}

		// 获取字段名
		fieldName := fieldType.Name

		// 检查 json tag，如果有的话使用 json tag 作为 key
		if jsonTag := fieldType.Tag.Get(tag); jsonTag != "" && jsonTag != "-" {
			// 处理 json tag，只取逗号前的部分
			if commaIndex := len(jsonTag); commaIndex > 0 {
				for j, char := range jsonTag {
					if char == ',' {
						commaIndex = j
						break
					}
				}
				if jsonTag[:commaIndex] != "" {
					fieldName = jsonTag[:commaIndex]
				}
			}
		}

		// 获取字段值
		fieldValue := field.Interface()

		// 如果字段也是结构体，递归处理
		if field.Kind() == reflect.Struct {
			fieldValue = StructToMap(fieldValue, tag)
		} else if field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct {
			fieldValue = StructToMap(field.Interface(), tag)
		}

		result[fieldName] = fieldValue
	}

	return result
}
