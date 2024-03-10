package consumer

import (
	"errors"
	"jcpd.cn/post/internal/constants"
	"jcpd.cn/post/internal/models"
	"log"
	"strconv"
)

type pair struct {
	userId string
	postId string
}

type AggregateArray struct {
	Pairs      []pair
	Record     map[string]bool
	UpperLimit int
}

func NewAggregateArray(upperLimit int) *AggregateArray {
	return &AggregateArray{Pairs: make([]pair, 0), Record: make(map[string]bool), UpperLimit: upperLimit}
}

func (array *AggregateArray) size() int {
	return len(array.Pairs)
}

func (array *AggregateArray) enough() bool {
	return array.size() > array.UpperLimit
}

func (array *AggregateArray) add(pair pair) {
	array.Pairs = append(array.Pairs, pair)
	array.Record[pair.userId+"#"+pair.postId] = true
}

func (array *AggregateArray) clear() {
	array.Pairs = make([]pair, 0)
	array.Record = make(map[string]bool)
}

func (array *AggregateArray) exist(pair pair) bool {
	return array.Record[pair.userId+"#"+pair.postId]
}

func (array *AggregateArray) sqlExecute(start string) error {
	var sql = start
	if sql == "insert" {
		sql += " into " + models.LikeInfoTN + " (user_id,post_id) values "
		for _, pair := range array.Pairs {
			sql += "(" + pair.userId + "," + pair.postId + "),"
		}
		sql = sql[:len(sql)-1]

		err := models.LikeInfoDao.DB.Exec(sql).Error
		if err != nil {
			log.Println(constants.Err("上传点赞记录失败, err = " + err.Error()))
		}
		return err
	} else if sql == "delete" {
		sql += " from " + models.LikeInfoTN + " where "
		for i, pair := range array.Pairs {
			sql += "(user_id = " + pair.userId + " and post_id = " + pair.postId + ")"
			if i != array.size()-1 {
				sql += " or "
			}
		}
		err := models.LikeInfoDao.DB.Exec(sql).Error
		if err != nil {
			log.Println(constants.Err("删除点赞记录失败, err = " + err.Error()))
			//	然后执行失败的sql应该被保存起来，不然就丢了
		}
		return err
	}
	return errors.New("1")
}

//  ----------------------------------------------

type IdLikesManager struct {
	IdLikeMap  map[string]int
	UpperLimit int
}

func NewIdLikesManager(upperLimit int) *IdLikesManager {
	return &IdLikesManager{IdLikeMap: make(map[string]int), UpperLimit: upperLimit}
}

func (manager *IdLikesManager) size() int {
	return len(manager.IdLikeMap)
}

func (manager *IdLikesManager) enough() bool {
	return manager.size() > manager.UpperLimit
}

func (manager *IdLikesManager) add(postId string, isLike string) {
	if isLike == Like {
		manager.IdLikeMap[postId] = manager.IdLikeMap[postId] + 1
	} else if isLike == Dislike {
		manager.IdLikeMap[postId] = manager.IdLikeMap[postId] - 1
	}
}

func (manager *IdLikesManager) clear() {
	manager.IdLikeMap = make(map[string]int)
}

func (manager *IdLikesManager) sqlExecute() error {
	var sql = "update " + models.PostInfoTN + " set likes = case"
	var ids = ""

	for postId, likeChange := range manager.IdLikeMap {
		sql += " when id = " + postId + " then likes + " + strconv.Itoa(likeChange)
		ids += postId + ","
	}

	ids = ids[:len(ids)-1]

	sql += " else likes end where id in (" + ids + ");"

	err := models.LikeInfoDao.DB.Exec(sql).Error
	if err != nil {
		log.Println(constants.Err("更新点赞信息失败, err = " + err.Error()))
		//	然后执行失败的sql应该被保存起来，不然就丢了
	}
	return err
}
