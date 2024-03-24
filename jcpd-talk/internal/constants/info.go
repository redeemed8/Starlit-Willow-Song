package constants

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"jcpd.cn/talk/internal/options"
	"log"
	"sync"
	"time"
)

type Err_ struct{}

func (*Err_) CheckMysqlErr(err error) bool {
	return err != nil && !errors.Is(err, gorm.ErrRecordNotFound)
}

func (*Err_) CheckRedisErr(err error) bool {
	return err != nil && !errors.Is(err, redis.Nil)
}

// 两个设备对应的 once，让对应设备的异常恢复只有一次执行，防止处理异常的协程过多
var mysqlOnce = newOnce()
var redisOnce = newOnce()

func newOnce() *sync.Once {
	return &sync.Once{}
}

var RedisStatus string
var MysqlStatus string

type DeviceType string

const (
	OK        = "ok"
	Exception = "exception"

	MYSQL DeviceType = "mysql"
	REDIS DeviceType = "redis"
)

func MysqlErr(msg string, err error) {
	if MysqlStatus == Exception {
		return
	}
	//	保证处理函数只会执行一次
	mysqlOnce.Do(func() {
		MysqlStatus = Exception
		log.Printf(Err(fmt.Sprintf("Error : Mysql exception , %s , cause by : %v \n", msg, err)))
		handleErr(msg, err, MYSQL)
	})
}

func RedisErr(msg string, err error) {
	if RedisStatus == Exception {
		return
	}
	//	保证处理函数只会执行一次
	redisOnce.Do(func() {
		RedisStatus = Exception
		log.Printf(Err(fmt.Sprintf("Error : Redis exception , %s , cause by : %v \n", msg, err)))
		handleErr(msg, err, REDIS)
	})
}

const Mobile = "xxxxxx" //	可以写成邮箱，亦可以在配置文件中定义

func alertErr(msg string, err error) {
	//	TODO 此处可以通过消息队列异步通知异常到运维人员的邮箱或手机
	//	...
	log.Printf(Hint(fmt.Sprintf("发送异常通知消息到 手机号 : %s , 异常信息 : %s , 错误 : %v ... \n ", Mobile, msg, err)))
}

func handleErr(msg string, err error, device DeviceType) {
	//	服务异常的处理机制
	alertErr(msg, err)
	//	重开一个协程，监听服务的恢复
	go func() {
		listenRecover(device)
	}()
	//	让handleErr结束，从而结束 once.DO
	return
}

const TickerTime = 15 * time.Second

func listenRecover(device DeviceType) {
	// 定时任务，每 TickerTime 执行一次
	ticker := time.NewTicker(TickerTime)
	defer ticker.Stop()
	//	最后恢复成功只需要修改 对应的 Status常量
	for {
		if device == MYSQL && MysqlStatus == OK {
			return
		} else if device == REDIS && RedisStatus == OK {
			return
		}
		select {
		case <-ticker.C:
			//  测试对应连接
			recover_(device)
		}
	}
}

func recover_(device DeviceType) {
	switch device {
	case MYSQL:
		recoverMysql()
	case REDIS:
		recoverRedis()
	}
}

var (
	MysqlLogger *logger.Interface
	MysqlDsn    string
)

func recoverMysql() {
	//	初始化连接
	db, err := gorm.Open(mysql.Open(MysqlDsn), &gorm.Config{Logger: *MysqlLogger})
	if err != nil {
		log.Printf("Application two failed to recover Mysql database , cause by : %v ... \n", err)
		return
	}
	if db == nil {
		log.Println(Hint(fmt.Sprintf("Application two failed to recover Mysql database , cause by the connection is abnormal , db == nil ...")))
		return
	}
	//	尝试查询
	var version string
	if err1 := db.Raw("Select version()").Scan(&version).Error; err1 != nil || version == "" {
		log.Printf(Hint(fmt.Sprintf("Application two failed to recover Mysql database , cause by the connection is abnormal by test , err = %v ... \n", err1)))
		return
	}
	//	查询到 version , 说明成功恢复
	options.C.DB = db
	//	恢复 MysqlStatus 到 ok
	MysqlStatus = OK
	//	等待是为了 让同时段的请求都执行完，防止有的请求慢，导致的下面的once.DO函数被无缝衔接
	time.Sleep(5 * time.Second)
	//	恢复 MysqlOnce , 等待下次重用
	mysqlOnce = newOnce()
	log.Println(Info(fmt.Sprintf("Application two recover Mysql database Successfully ... ")))
}

var RedisOptions *redis.Options

func recoverRedis() {
	//	初始化连接
	rdb := redis.NewClient(RedisOptions)
	if rdb == nil {
		log.Println(Info(fmt.Sprintf("Application two failed to recover Redis database , cause by the connection is abnormal , rdb == nil ... ")))
		return
	}
	//	测试连接
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Printf(Info(fmt.Sprintf("Application two failed to recover Redis database , cause by the connection is abnormal by test , err = %v ... \n", err)))
		return
	}
	//	查询到 version , 说明成功恢复
	options.C.RDB = rdb
	//	恢复 MysqlStatus 到 ok
	RedisStatus = OK
	//	等待是为了 让同时段的请求都执行完，防止有的请求慢，导致的下面的once.DO函数被无缝衔接
	time.Sleep(5 * time.Second)
	//	恢复 MysqlOnce , 等待下次重用
	redisOnce = newOnce()
	log.Println(Info(fmt.Sprintf("Application two recover Redis database Successfully ... ")))

}
