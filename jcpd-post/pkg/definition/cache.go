package definition

import (
	"context"
	"github.com/go-redis/redis/v8"
	"jcpd.cn/post/internal/constants"
	"jcpd.cn/post/internal/options"
	"log"
	"time"
)

type Cache interface {
	Put(key, value string, expire time.Duration) error
	Get(key string) (string, error)
}

type CacheType string

const (
	CacheRedis CacheType = "redis_"
	CacheMysql CacheType = "mysql_"
	CacheMongo CacheType = "mongo_"
	Memcahce   CacheType = "memcache_"
)

//	任选一种在其下方进行实现 -- 这里使用 redis

var Rc = New()

type RedisCache struct {
	rdb *redis.Client
}

func New() *RedisCache {
	return &RedisCache{}
}

func Init() {
	if Rc.rdb == nil {
		Rc.rdb = options.C.RDB
	}
}

func (Rc *RedisCache) Put(key, value string, expire time.Duration) error {
	Init()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := Rc.rdb.Set(ctx, key, value, expire).Err()
	return err
}

func (Rc *RedisCache) Get(key string) (string, error) {
	Init()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, err := Rc.rdb.Get(ctx, key).Result()
	return result, err
}

func (Rc *RedisCache) HashPut(key, field, value string, expire time.Duration) error {
	Init()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	//	使用管道发送多条命令，减少 网络I/O 和 服务器负载，而且是原子性操作，支持事务
	pipe := Rc.rdb.TxPipeline()

	//	塞入命令
	pipe.HSet(ctx, key, field, value) //	存值
	pipe.Expire(ctx, key, expire)     //	过期时间

	//	事务执行管道命令
	_, err := pipe.Exec(context.Background())
	if err != nil {
		log.Println(constants.Err("redis事务执行出错-hash缓存出错-err = " + err.Error()))
		return err
	}
	return nil
}

func (Rc *RedisCache) HashGet(key, field string) (string, error) {
	Init()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	//  获取 hash值
	value, err := Rc.rdb.HGet(ctx, key, field).Result()

	return value, err
}
