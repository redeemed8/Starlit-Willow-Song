package models

import (
	"jcpd.cn/post/internal/constants"
	"log"
)

func InitPost() {
	//	初始化DAO
	newDao()
	//	初始化mysql表
	createTables()
	//	初始化布隆过滤器
	initBloomFilters()
}

func newDao() {

	NewPostInfoDao()

	NewLikeInfoDao()

	NewCommentInfoDao()

}

func createTables() {
	log.Println(constants.Hint("------------------- create tables --------------------"))

	PostInfoDao.CreateTable()

	LikeInfoDao.CreateTable()

	CommentInfoDao.CreateTable()

	log.Println(constants.Hint("------------------- create success -------------------"))
}
