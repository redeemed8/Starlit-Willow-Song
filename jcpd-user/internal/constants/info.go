package constants

import "log"

func MysqlErr(msg string, err error) {
	log.Printf("Error : Mysql exception , %s , cause by : %v \n", msg, err)
}

func RedisErr(msg string, err error) {
	log.Printf("Error : Redis exception , %s , cause by : %v \n", msg, err)
}
