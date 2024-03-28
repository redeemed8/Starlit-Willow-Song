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
	NewMessageCounterDao()
}

func createTables() {
	log.Println(constants.Hint("------------------- create tables --------------------"))

	MessageInfoDao.CreateTable()
	MessageCounterDao.CreateTable()

	log.Println(constants.Hint("------------------- create success -------------------"))
}
