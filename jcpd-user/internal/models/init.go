package models

import (
	commonJWT "jcpd.cn/common/utils/jwt"
	"jcpd.cn/user/internal/options"
	"log"
)

func InitUser() {
	//	初始化 jwt的数据库连接
	commonJWT.NewDB(options.C.DB)

	newDao()

	createTables()
}

func newDao() {
	NewUserInfoDao()
	NewPointInfoDao()
	NewJoinApplyDao()
}

func createTables() {
	log.Println("------------------- create tables --------------------")

	UserInfoDao.CreateTable()
	PointInfoDao.CreateTable()
	JoinApplyDao.CreateTable()

	log.Println("------------------- create success -------------------")
}
