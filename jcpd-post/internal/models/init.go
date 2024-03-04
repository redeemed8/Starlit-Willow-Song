package models

import (
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
	log.Println("------------------- create tables --------------------")

	PostInfoDao.CreateTable()

	log.Println("------------------- create success -------------------")
}
