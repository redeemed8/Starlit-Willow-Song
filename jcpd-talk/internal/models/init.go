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

}

func createTables() {
	log.Println(constants.Hint("------------------- create tables --------------------"))

	log.Println(constants.Hint("------------------- create success -------------------"))
}
