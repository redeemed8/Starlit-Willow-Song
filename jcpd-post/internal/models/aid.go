package models

import (
	"encoding/binary"
	"errors"
	"github.com/willf/bloom"
	"gorm.io/gorm"
	"jcpd.cn/post/internal/constants"
	"log"
	"time"
)

type PageArgs struct {
	PageNum  int
	PageSize int
}

// 	-----------------------------------------------------------------

const bloomFilterCount = 20
const bloomFilterCap = 100000
const bloomFilterFp = 0.01

// bloomFilters 布隆过滤器组
type bloomFilters [bloomFilterCount]*bloom.BloomFilter

var BloomFilters bloomFilters //	对外提供服务

// initBloomFilters 初始化布隆过滤器组
func initBloomFilters() {
	filterStatus = filterWorking

	for i := 0; i < bloomFilterCount; i++ {
		// 创建一个容量为100000，假阳性率为0.01的布隆过滤器
		BloomFilters[i] = bloom.NewWithEstimates(bloomFilterCap, bloomFilterFp)
	}

	//	 将数据库中的id 初始化加载到布隆过滤器组
	log.Println(constants.Hint("布隆过滤器加载中ing..."))
	ids, err := PostInfoDao.GetAllIds()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		constants.MysqlErr("获取所有帖子id出错", err)
		log.Fatalln(constants.Err("布隆过滤器加载失败, 从数据库获取id异常"))
		return
	}
	BloomFilters.AddInBatches(ids)
	log.Println(constants.Info("布隆过滤器加载完成..."))
}

func (filter *bloomFilters) Flush() {
	filter.rest()

	//	为了减少一些可能的不必要的问题
	time.Sleep(2 * time.Second)

	for i := 0; i < bloomFilterCount; i++ {
		// 创建一个容量为100000，假阳性率为0.01的布隆过滤器
		BloomFilters[i] = bloom.NewWithEstimates(bloomFilterCap, bloomFilterFp)
	}

	//	 将数据库中的id 重新加载到布隆过滤器组
	log.Println(constants.Hint("布隆过滤器刷新中ing..."))
	ids, err := PostInfoDao.GetAllIds()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		constants.MysqlErr("获取所有帖子id出错", err)
		log.Fatalln(constants.Err("布隆过滤器刷新失败, 从数据库获取id异常"))
		return
	}
	BloomFilters.AddInBatches(ids)
	log.Println(constants.Info("布隆过滤器刷新完成..."))

	filter.work()
}

// 	-----------------------------------------------------------------

func (filter *bloomFilters) makeUint32Bytes(uint32num uint32) []byte {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32num)
	return bytes
}

const filterWorking = "filter_working"
const filterResting = "filter_resting"

var filterStatus string

func (filter *bloomFilters) work() {
	filterStatus = filterWorking
}

func (filter *bloomFilters) rest() {
	filterStatus = filterResting
}

// check 检查过滤器组是否可用， 有可能处在更新状态，将暂停使用, 返回值为是否可用
func (filter *bloomFilters) check() bool {
	return filterStatus == filterWorking
}

// Add 添加元素
func (filter *bloomFilters) Add(postId uint32) {
	if !filter.check() {
		return
	}
	index := postId%bloomFilterCount - 1
	(*filter)[index].Add(filter.makeUint32Bytes(postId))
}

// AddInBatches 批量添加
func (filter *bloomFilters) AddInBatches(postIds []uint32) {
	if !filter.check() {
		return
	}
	for _, id := range postIds {
		index := id%bloomFilterCount - 1
		(*filter)[index].Add(filter.makeUint32Bytes(id))
	}
}

// Exist 判断id是否存在
func (filter *bloomFilters) Exist(postId uint32) bool {
	if !filter.check() {
		return true //	过滤器摆烂了，来者不拒
	}
	index := postId%bloomFilterCount - 1
	return (*filter)[index].Test(filter.makeUint32Bytes(postId))
}
