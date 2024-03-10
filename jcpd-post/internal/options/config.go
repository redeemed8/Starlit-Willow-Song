package options

import (
	"github.com/IBM/sarama"
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
	App      AppConfig
	Mysql    MysqlConfig
	Redis    RedisConfig
	KafKa    KafkaConfig
	DB       *gorm.DB
	RDB      *redis.Client
	Producer *sarama.AsyncProducer
	Consumer sarama.ConsumerGroup
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

// KafkaConfig kafka一级配置类
type KafkaConfig struct {
	Consumer Consumer
	Producer Producer
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

// Consumer kafka二级配置类
type Consumer struct {
	GroupId          string
	Brokers          []string
	Topics           []string
	ReturnErrors     bool
	OffsetAutoCommit bool
}

// Producer kafka二级配置类
type Producer struct {
	Addr          []string
	Topics        []string
	ReturnSuccess bool
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

func ReadKafkaConfig() {
	producerPrefix := "kafka.producer."
	C.KafKa.Producer = Producer{
		Addr:          viper.GetStringSlice(producerPrefix + "addr"),
		Topics:        viper.GetStringSlice(producerPrefix + "topics"),
		ReturnSuccess: viper.GetBool(producerPrefix + "return-success"),
	}

	consumerPrefix := "kafka.consumer."
	C.KafKa.Consumer = Consumer{
		GroupId:          viper.GetString(consumerPrefix + "group-id"),
		Brokers:          viper.GetStringSlice(consumerPrefix + "brokers"),
		Topics:           viper.GetStringSlice(consumerPrefix + "topics"),
		ReturnErrors:     viper.GetBool(consumerPrefix + "return-errors"),
		OffsetAutoCommit: viper.GetBool(consumerPrefix + "offset-auto-commit"),
	}
}
