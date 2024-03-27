package models

import (
	"jcpd.cn/talk/internal/constants"
	"log"
)

func Init() {
	//	初始化DAO
	newDao()
	//	初始化mysql表
	createTables()
}

func newDao() {
	NewMessageInfoDao()
}

func createTables() {
	log.Println(constants.Hint("------------------- create tables --------------------"))

	MessageInfoDao.CreateTable()

	log.Println(constants.Hint("------------------- create success -------------------"))
}
