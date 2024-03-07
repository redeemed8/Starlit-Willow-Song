package definition

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"jcpd.cn/post/internal/options"
	"time"
)

type Cache interface {
	Put(key, value string, expire time.Duration) error
	Get(key string) (string, error)

	HashMultiPut(string, map[string]string, time.Duration) (error, string)
	HashMultiGet(string) (map[string]string, error, string)
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

func (Rc *RedisCache) HashMultiPut(key string, FV map[string]string, expire time.Duration) (error, string) {
	Init()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	//	使用管道发送多条命令，减少 网络I/O 和 服务器负载，而且是原子性操作，支持事务
	pipe := Rc.rdb.TxPipeline()

	//	塞入命令
	pipe.HMSet(ctx, key, FV)      //	存值
	pipe.Expire(ctx, key, expire) //	过期时间

	//	事务执行管道命令
	_, err := pipe.Exec(context.Background())
	if err != nil && !errors.Is(err, redis.Nil) {
		return err, "redis事务执行出错-hash缓存出错"
	}
	return nil, ""
}

func (Rc *RedisCache) HashMultiGet(key string) (map[string]string, error, string) {
	Init()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	//  获取 hash值
	result, err := Rc.rdb.HGetAll(ctx, key).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return make(map[string]string), err, "redis事务执行出错-获取hash表的field和value出错"
	}

	return result, err, ""
}
