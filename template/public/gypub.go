package public

import (
	"fmt"
	"math"
	"reflect"
	"sync"
	"thermometer/model"
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

// AnalyzeSensorStatus 分析陀螺仪和加速度数据对应的状态
// accelX, accelY, accelZ: 加速度数据 (m/s²)
// gyroX, gyroY, gyroZ: 陀螺仪数据 (°/s)
func AnalyzeSensorStatus(accelX, accelY, accelZ, gyroX, gyroY, gyroZ float64) model.SensorStatus {
	// 计算运动强度（加速度的模长）
	movementIntensity := math.Sqrt(accelX*accelX + accelY*accelY + accelZ*accelZ)

	// 计算陀螺仪数据的模长（角速度）
	gyroMagnitude := math.Sqrt(gyroX*gyroX + gyroY*gyroY + gyroZ*gyroZ)

	// 计算稳定性评分（陀螺仪数据越小，稳定性越高）
	stabilityScore := math.Max(0, 1.0-gyroMagnitude/100.0)

	// 计算重力方向（用于判断姿态）
	gravityMagnitude := math.Sqrt(accelX*accelX + accelY*accelY + accelZ*accelZ)

	status := model.SensorStatus{
		MovementIntensity: movementIntensity,
		StabilityScore:    stabilityScore,
		Alert:             false,
	}

	// 跌倒检测（高加速度变化 + 突然的方向改变）
	if movementIntensity > 20.0 || gyroMagnitude > 300.0 {
		status.Status = "falling"
		status.PostureType = "跌倒"
		status.MovementLevel = "急剧"
		status.Confidence = 0.9
		status.Alert = true
		status.AlertMessage = "检测到可能的跌倒事件"
		return status
	}

	// 剧烈运动检测
	if movementIntensity > 15.0 && gyroMagnitude > 200.0 {
		status.Status = "running"
		status.PostureType = "剧烈运动"
		status.MovementLevel = "剧烈"
		status.Confidence = 0.85
		return status
	}

	// 一般运动检测
	if movementIntensity > 8.0 && gyroMagnitude > 100.0 {
		status.Status = "walking"
		status.PostureType = "运动"
		status.MovementLevel = "中等"
		status.Confidence = 0.8
		return status
	}

	// 基于稳定性和运动强度进行姿态判断
	if movementIntensity < 1.5 && stabilityScore > 0.8 {
		// 非常稳定，低运动强度
		if math.Abs(accelZ) > math.Abs(accelX) && math.Abs(accelZ) > math.Abs(accelY) {
			status.Status = "lying"
			status.PostureType = "平躺"
			status.MovementLevel = "静止"
			status.Confidence = 0.9
		} else {
			status.Status = "lying"
			status.PostureType = "侧卧"
			status.MovementLevel = "静止"
			status.Confidence = 0.8
		}
	} else if movementIntensity < 3.0 && stabilityScore > 0.6 {
		// 较稳定，轻微运动
		if gravityMagnitude > 8.0 { // 接近标准重力加速度
			status.Status = "sitting"
			status.PostureType = "坐立"
			status.MovementLevel = "轻微"
			status.Confidence = 0.7
		} else {
			status.Status = "lying"
			status.PostureType = "侧卧"
			status.MovementLevel = "轻微"
			status.Confidence = 0.6
		}
	} else if movementIntensity < 6.0 && stabilityScore > 0.4 {
		// 中等稳定性
		status.Status = "standing"
		status.PostureType = "站立"
		status.MovementLevel = "轻微"
		status.Confidence = 0.6
	} else if movementIntensity < 10.0 {
		// 较大运动
		status.Status = "walking"
		status.PostureType = "行走"
		status.MovementLevel = "中等"
		status.Confidence = 0.7
	} else {
		// 默认为运动状态
		status.Status = "walking"
		status.PostureType = "运动"
		status.MovementLevel = "中等"
		status.Confidence = 0.5
	}

	return status
}

// AnalyzeSensorDataStatus 直接从传感器数据分析状态
func AnalyzeSensorDataStatus(sensorData *model.SensorData) model.SensorStatus {
	// 将传感器数据转换为物理单位
	accelX := float64(sensorData.AccelX) / 1000.0 // 转换为 m/s²
	accelY := float64(sensorData.AccelY) / 1000.0
	accelZ := float64(sensorData.AccelZ) / 1000.0
	gyroX := float64(sensorData.GyroX) / 1000.0 // 转换为 °/s
	gyroY := float64(sensorData.GyroY) / 1000.0
	gyroZ := float64(sensorData.GyroZ) / 1000.0

	return AnalyzeSensorStatus(accelX, accelY, accelZ, gyroX, gyroY, gyroZ)
}

// GetMovementLevel 根据运动强度获取运动级别
func GetMovementLevel(intensity float64) string {
	if intensity < 1.0 {
		return "静止"
	} else if intensity < 3.0 {
		return "轻微"
	} else if intensity < 8.0 {
		return "中等"
	} else if intensity < 15.0 {
		return "活跃"
	} else {
		return "剧烈"
	}
}

// IsAbnormalMovement 检测是否为异常运动（可能需要关注）
func IsAbnormalMovement(status model.SensorStatus) bool {
	return status.Alert ||
		status.Status == "falling" ||
		(status.MovementIntensity > 20.0) ||
		(status.StabilityScore < 0.2 && status.MovementIntensity > 5.0)
}
