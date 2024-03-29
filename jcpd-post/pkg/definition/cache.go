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

	HSet(key, field, value string, expire time.Duration) error
	HGet(key, field string) (string, error)

	HIncrBy(key, field string, addValue int64) (int64, error)

	HashMultiPut(string, map[string]string, time.Duration) (error, string)
	HashMultiGet(string) (map[string]string, error, string)

	SetNX(key, value string) error
	Delete(string) error
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

func (Rc *RedisCache) HSet(key, field, value string, expire time.Duration) error {
	Init()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	pipe := Rc.rdb.Pipeline()

	pipe.HSet(ctx, key, field, value)
	pipe.Expire(ctx, key, expire)

	//	执行管道
	_, err := pipe.Exec(ctx)

	return err
}

func (Rc *RedisCache) HGet(key, field string) (string, error) {
	Init()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := Rc.rdb.HGet(ctx, key, field).Result()
	return result, err
}

func (Rc *RedisCache) HIncrBy(key, field string, addValue int64) (int64, error) {
	Init()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := Rc.rdb.HIncrBy(ctx, key, field, addValue).Result()
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

func (Rc *RedisCache) SetNX(key, value string) error {
	Init()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := Rc.rdb.SetNX(ctx, key, value, 15).Result() //	不设置为永不过期，防止刚设置锁就宕机导致的死锁

	return err //	没抢到锁会返回 redis.Nil
}

func (Rc *RedisCache) Delete(key string) error {
	Init()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := Rc.rdb.Del(ctx, key).Result()

	return err
}
