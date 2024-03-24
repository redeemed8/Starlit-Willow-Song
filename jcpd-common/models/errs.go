package common_models

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type NormalErr struct {
	Code int
	Msg  string
}

func MakeNormalErr(code int, msg string) *NormalErr {
	return &NormalErr{Code: code, Msg: msg}
}

func (e *NormalErr) Err() error {
	return status.Error(codes.Code(e.Code), e.Msg)
}

func ToNormalErr(err error) NormalErr {
	fromError, _ := status.FromError(err)
	return NormalErr{Code: int(fromError.Code()), Msg: fromError.Message()}
}

func RedisException(err error) {
	log.Println("Error : Redis exception , cause by : ", err)
	//	通过消息队列 通知其他服务
}

func MysqlException(err error) {
	log.Println("Error : Mysql exception , cause by : ", err)
	//	通过消息队列 通知其他服务
}
