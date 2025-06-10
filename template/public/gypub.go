package public

import (
	"fmt"
	"time"
	"{firstweb}/model"

	"github.com/guyigood/gyweb/core/middleware"
	orm "github.com/guyigood/gyweb/core/orm/mysql"
	"github.com/guyigood/gyweb/core/utils/common"
	"github.com/guyigood/gyweb/core/utils/datatype"
	"github.com/guyigood/gyweb/core/utils/redisutils"
)

var (
	SysConfig model.AppConfig
	Re_Client *redisutils.RedisClient
	Db        *orm.DB
)

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

	dsn := SysConfig.Database.Username + ":" + SysConfig.Database.Password + "@tcp(" + SysConfig.Database.Host + ":" + datatype.TypetoStr(SysConfig.Database.Port) + ")/" + SysConfig.Database.Dbname + "?charset=utf8mb4&parseTime=True&loc=Local"
	Db, err = orm.NewDB(SysConfig.Database.Dialect, dsn)
	if err != nil {
		panic(err)
		return
	}
	Db.SetMaxIdleConns(SysConfig.Database.Pool.Max)
	Db.SetMaxOpenConns(SysConfig.Database.Pool.Idle)
	Db.SetConnMaxLifetime(time.Duration(SysConfig.Database.Pool.Lifetime) * time.Second)
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
}
func GetJwtConfig() *middleware.JWTConfig {
	return &middleware.JWTConfig{
		SecretKey:     "gy7210",               // JWT密钥
		TokenLookup:   "header:Authorization", // 从请求头获取token
		TokenHeadName: "Bearer",               // token前缀
		ExpiresIn:     2 * 24 * time.Hour,     // token过期时间

	}

}
