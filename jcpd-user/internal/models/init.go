package models

import (
	"jcpd.cn/user/internal/constants"
	"log"
)

func InitUser() {

	newDao()

	createTables()
}

func newDao() {
	NewUserInfoDao()
	NewPointInfoDao()
	NewJoinApplyDao()
	NewGroupInfoDao()
}

func createTables() {
	log.Println(constants.Hint("------------------- create tables --------------------"))

	UserInfoDao.CreateTable()
	PointInfoDao.CreateTable()
	JoinApplyDao.CreateTable()
	GroupInfoDao.CreateTable()

	log.Println(constants.Hint("------------------- create success -------------------"))
}
