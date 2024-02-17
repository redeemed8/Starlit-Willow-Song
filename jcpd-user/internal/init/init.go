package init

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"jcpd.cn/user/internal/options"
	"log"
	"os"
	"time"
)

// 初始化 yaml文件
func init() {
	viper.SetConfigName("application") //	配置文件名称
	viper.AddConfigPath("config")      //	配置文件在根目录的位置
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Application one failed to read configuration , cause by : %v ... \n", err)
	}
	options.ReadAppConfig()
	options.ReadMysqlConfig()
	options.ReadRedisConfig()
	log.Println("Application one init configuration successfully ... ")
}

// 初始化 mysql
func init() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, //	慢 sql阈值
			LogLevel:      logger.Info, //	日志级别
			Colorful:      true,        //	彩色
		})
	dsn := options.C.Mysql.DSN.User + ":" +
		options.C.Mysql.DSN.Password + "@tcp(" +
		options.C.Mysql.DSN.Addr + ")/" +
		options.C.Mysql.Basename + "?" +
		options.C.Mysql.Others
	//	初始化连接
	var err error
	options.C.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		log.Fatalf("Application one failed to connect Mysql database , cause by : %v ... \n", err)
	}
	if options.C.DB == nil {
		log.Fatalln("Application one failed to connect Mysql database , cause by the connection is abnormal ...")
	}
	var version string
	if err1 := options.C.DB.Raw("Select version()").Scan(&version).Error; err1 != nil || version == "" {
		log.Fatalf("Application one failed to connect Mysql database , cause by the connection is abnormal by test , err = %v ... \n", err1)
	}
	log.Println("Application one connect Mysql database successfully ...")
}

// 初始化 redis
func init() {
	options.C.RDB = redis.NewClient(
		&redis.Options{
			Addr:         options.C.Redis.Addr,
			Password:     options.C.Redis.Password,
			DB:           options.C.Redis.DB,
			PoolSize:     options.C.Redis.PoolSize,
			MinIdleConns: options.C.Redis.MinIdleConn,
		})
	if options.C.RDB == nil {
		log.Fatalln("Application one failed to connect Redis database , cause by the connection is abnormal ...")
	}

	_, err := options.C.RDB.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Application one failed to connect Redis database , cause by the connection is abnormal by test , err = %v ... \n", err)
	}
	log.Println("Application one connect Redis database successfully ...")
}
