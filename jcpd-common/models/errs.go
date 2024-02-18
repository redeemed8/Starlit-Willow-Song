package common_models

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NormalErr struct {
	Code int
	Msg  string
}

func (e *NormalErr) Err() error {
	return status.Error(codes.Code(e.Code), e.Msg)
}

func CodeAndMsg(err error) (int, string) {
	fromError, _ := status.FromError(err)
	return int(fromError.Code()), fromError.Message()
}
