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
}

func newDao() {

	NewPostInfoDao()

}

func createTables() {
	log.Println(constants.Hint("------------------- create tables --------------------"))

	PostInfoDao.CreateTable()

	log.Println(constants.Hint("------------------- create success -------------------"))
}
