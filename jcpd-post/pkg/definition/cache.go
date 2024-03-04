package definition

import (
	"context"
	"github.com/go-redis/redis/v8"
	"jcpd.cn/post/internal/options"
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
