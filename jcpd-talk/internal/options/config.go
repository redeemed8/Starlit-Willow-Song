package options

import (
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var C = New()

func New() *Config {
	return &Config{}
}

// Config 顶级配置类
type Config struct {
	App   AppConfig
	Mysql MysqlConfig
	Redis RedisConfig
	DB    *gorm.DB
	RDB   *redis.Client
}

// AppConfig app一级配置类
type AppConfig struct {
	Server Server
}

// MysqlConfig mysql一级配置类
type MysqlConfig struct {
	DSN      DSN
	Basename string
	Others   string
}

// RedisConfig redis一级配置类
type RedisConfig struct {
	Addr        string
	Password    string
	DB          int
	PoolSize    int
	MinIdleConn int
}

// Server app二级配置类
type Server struct {
	Port string
	Name string
}

// DSN mysql二级配置类
type DSN struct {
	User     string
	Password string
	Addr     string
}

func ReadAppConfig() {
	server := Server{
		Port: ":" + viper.GetString("application.server.port"),
		Name: viper.GetString("application.server.name"),
	}
	C.App.Server = server
}

func ReadMysqlConfig() {
	dns := DSN{
		User:     viper.GetString("mysql.dsn.user"),
		Password: viper.GetString("mysql.dsn.password"),
		Addr:     viper.GetString("mysql.dsn.addr"),
	}
	basename := viper.GetString("mysql.basename")
	others := viper.GetString("mysql.others")
	C.Mysql.DSN = dns
	C.Mysql.Basename = basename
	C.Mysql.Others = others
}

func ReadRedisConfig() {
	C.Redis = RedisConfig{
		Addr:        viper.GetString("redis.addr"),
		Password:    viper.GetString("redis.password"),
		DB:          viper.GetInt("redis.DB"),
		PoolSize:    viper.GetInt("redis.poolSize"),
		MinIdleConn: viper.GetInt("redis.minIdleConn"),
	}
}
