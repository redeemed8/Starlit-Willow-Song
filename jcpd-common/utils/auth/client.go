package auth

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"jcpd.cn/common/config"
	color "jcpd.cn/common/utils/color"
	grpcService "jcpd.cn/user/pkg/service"
	"log"
)

var UserServiceClient = getGrpcUserClient()

// getGrpcUserClient 创建客户端连接
func getGrpcUserClient() grpcService.UserServiceClient {
	//	创建 token代表我知道什么 0.0
	token := &Authentication{
		Token: TokenSecret,
	}
	//	携带 token进行连接
	conn, err := grpc.Dial(config.M.Get(config.User), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(token))

	if err != nil {
		log.Fatalf(color.Err(fmt.Sprintf("Failed to connect , cause by : %s \n", err.Error())))
	}

	return grpcService.NewUserServiceClient(conn)
}
