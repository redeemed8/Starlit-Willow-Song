package models

import (
	commonJWT "jcpd.cn/common/utils/jwt"
	"jcpd.cn/user/internal/options"
	"log"
)

func CreateTables() {

	//	初始化 jwt的数据库连接
	commonJWT.NewDB(options.C.DB)

	log.Println("------------------- create tables --------------------")

	UserInfoDao.CreateTable()

	log.Println("------------------- create success -------------------")
}