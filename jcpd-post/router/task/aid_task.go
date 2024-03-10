package task

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"jcpd.cn/post/internal/constants"
	"jcpd.cn/post/internal/options"
	"log"
)

//	由于 定时任务 updateHotPost 代码太多，不适合写在定时任务器的整体架构中, 所以决定创建辅助任务
//  结尾带有下划线的就是对应的辅助函数

// updateHotPostAid 辅助类
type updateHotPostAid struct{}

// updateHotPost_ 定时任务，将 redis中的点赞数，同步到mysql，同时更新热点帖子 id
func updateHotPost_(ctx context.Context) {

	//	思路：使用游标循环从0开始获取点赞的key，处理后删除这些key，再从0开始获取 ...

	aid := updateHotPostAid{}

	for cursor := uint64(0); ; {
		keys, nextCursor, err1 := aid.tryScan(ctx, cursor, constants.PostLikePrefix, EveryScanNum)
		if err1 != nil {
			return
		}
		cmds, err2 := aid.dealWithKeys(ctx, keys)
		if err2 != nil {
			return
		}
		err3 := aid.parseCmdsAndNotify(cmds, keys)
		if err3 != nil {
			return
		}
		err4 := aid.flushCursor(&cursor, nextCursor)
		if err4 != nil {
			return
		}
	}

}

const EveryScanNum = 1000

var rdb = options.C.RDB

// tryScan 尝试得到一些key
func (*updateHotPostAid) tryScan(ctx context.Context, cursor uint64, prefix string, everyScanNum int64) ([]string, uint64, error) {

	keys, nextCursor, err := rdb.Scan(ctx, cursor, prefix+"*", everyScanNum).Result()

	if err != nil {
		constants.RedisErr("redis的Scan批量获取key出错", err)
	}
	return keys, nextCursor, err
}

// dealWithKeys 对得到的key进行处理，得到他们中 对应的字段和值
func (*updateHotPostAid) dealWithKeys(ctx context.Context, keys []string) ([]redis.Cmder, error) {
	selectPipe := rdb.Pipeline()
	defer func(selectPipe redis.Pipeliner) {
		err := selectPipe.Close()
		if err != nil {
			log.Println(constants.Err("redis管道关闭异常，err = " + err.Error()))
		}
	}(selectPipe)

	//	批量获取
	for _, key := range keys {
		selectPipe.HGetAll(ctx, key)
	}

	//	批量删除
	for _, key := range keys {
		selectPipe.Del(ctx, key)
	}

	cmds, err := selectPipe.Exec(ctx)
	if err != nil {
		constants.RedisErr("redis的pipe批量查询key出错", err)
	}
	return cmds, err
}

// parseCmds 解析批量查询的结果，这里是已知结果类型是map
func (*updateHotPostAid) parseCmdsAndNotify(cmds []redis.Cmder, keys []string) error {
	//  获取到每个hash类型key对应的字段和值
	for i, cmd := range cmds {
		resultMap, err := cmd.(*redis.StringStringMapCmd).Result()
		if err != nil {
			constants.RedisErr("解析redis批量获取值map时出错", err)
			return err
		}
		//	将 对应的key 和 map 放入消息队列异步处理

		fmt.Println("key = ", keys[i], "   map = ", resultMap)

	}

	//	 发送给消息队列一个 刷新热点帖子缓存 的标记

	return nil
}

// flushCursor 刷新游标
func (*updateHotPostAid) flushCursor(curCursor *uint64, nextCursor uint64) error {
	if nextCursor == 0 {
		return errors.New("finish")
	}
	*curCursor = 0
	return nil
}
