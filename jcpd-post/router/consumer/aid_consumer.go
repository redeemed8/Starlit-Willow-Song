package consumer

import (
	"encoding/json"
	"errors"
	"github.com/IBM/sarama"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"jcpd.cn/post/internal/constants"
	"jcpd.cn/post/internal/models"
	"jcpd.cn/user/pkg/definition"
	"strings"
	"time"
)

type Consumer struct{}

// Setup 在消费者开始消费消息前调用的方法
func (consumer *Consumer) Setup(_ sarama.ConsumerGroupSession) error { return nil }

// Cleanup 在消费者停止消费消息后调用的方法
func (consumer *Consumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim 消费者实际消费消息的方法
func (consumer *Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	flag := 0
	for msg := range claim.Messages() {
		//	过滤掉每次启动的第一个, 因为会重复 ...
		if flag == 0 {
			flag++
			continue
		}

		KafkaListener.work() //	标记状态 为 工作中

		// 进行消息的处理
		msgBody := string(msg.Value)

		if msgBody == UpdateHotPostSymbol {
			//	更新热点帖子缓存
			updateHotPost()
		} else {
			//  更新点赞记录和点赞数
			updateLikes(msgBody)
		}

		// 手动标记偏移量并提交, 防止消息丢失和重复消费
		sess.MarkOffset(msg.Topic, msg.Partition, msg.Offset, "")
		sess.Commit()

		KafkaListener.rest() //	休息等待中
	}
	return nil
}

const KeyFvSplitSymbol = "$$"

const UpdateHotPostSymbol = "&(*&%*#@&"

const Like = "1"
const Dislike = "0"

func dealWithMsgBody(body string) (string, map[string]string, error) {
	fv := make(map[string]string)
	key := ""
	if body == "" {
		return key, fv, errors.New("1")
	}
	arr := strings.Split(body, KeyFvSplitSymbol)
	if len(arr) != 2 {
		return key, fv, errors.New("1")
	}
	//	key为 用户id，field为 帖子id
	keys := strings.Split(arr[0], ":")
	key = keys[len(keys)-1]
	err := json.Unmarshal([]byte(arr[1]), &fv)
	if err != nil {
		return key, fv, errors.New("1")
	}
	return key, fv, nil
}

//  ----------------------------------------------

const UpperLimit = 100

var createArr = NewAggregateArray(UpperLimit)
var deleteArr = NewAggregateArray(UpperLimit)
var idlikesMap = NewIdLikesManager(UpperLimit)

func updateLikes(msgBody string) {
	key, fv, err := dealWithMsgBody(msgBody)
	if err != nil {
		//	这里也可以将解析失败的body添加到数据库，以备后续处理，这里暂时省略
		return
	}
	//	创建点赞记录 或者 删除点赞记录
	for postId, isLike := range fv {
		pair := pair{userId: key, postId: postId}
		if isLike == Like {
			if createArr.exist(pair) {
				continue
			}
			createArr.add(pair)
		} else {
			if deleteArr.exist(pair) {
				continue
			}
			deleteArr.add(pair)
		}
		idlikesMap.add(postId, isLike)
	}
	//	点赞记录要立马创建，不然会影响用户体验
	if createArr.size() > 0 {
		_ = createArr.sqlExecute("insert")
		createArr.clear()
	}
	if deleteArr.size() > 0 {
		_ = deleteArr.sqlExecute("delete")
		deleteArr.clear()
	}
	//	更新点赞数
	if idlikesMap.size() > 0 {
		_ = idlikesMap.sqlExecute()
	}
}

//  ----------------------------------------------

func updateHotPost() {
	//	这里应该和前端商议每次获取的页大小 - 这里先按 50算
	postInfos, err := models.PostInfoDao.SimpleGetPostsPage(models.PageArgs{PageNum: 1, PageSize: 30})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		constants.MysqlErr("分页查询帖子信息出错", err)
		return
	}
	//	将查到的记录 添加到 redis
	err8 := definition.Rc.Put(constants.HotPostSummary, postInfos.ToIdStr(), 90*time.Minute)
	if err8 != nil && !errors.Is(err8, redis.Nil) {
		constants.RedisErr("获取redis缓存帖子id出错", err8)
		return
	}
}
