package api_init

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"jcpd.cn/user/internal/constants"
	"jcpd.cn/user/pkg/definition"
	grpcService "jcpd.cn/user/pkg/service"
	"log"
	"net"
)

// grpc 服务注册 ...

type grpcConfig struct {
	Addr         string
	Name         string
	RegisterFunc func(*grpc.Server)
}

func RegisterGrpc() *grpc.Server {
	c := grpcConfig{
		Addr: viper.GetString("grpc.addr"),
		Name: viper.GetString("grpc.name"),
		RegisterFunc: func(server *grpc.Server) {
			grpcService.RegisterUserServiceServer(server, grpcService.New())
		},
	}

	//	实现 token认证  --  一个拦截器
	var authInterceptor grpc.UnaryServerInterceptor
	authInterceptor = func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		err = Auth(ctx)
		if err != nil {
			return
		}
		return handler(ctx, req)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor)) //  使用拦截器创建一个新的server服务端
	c.RegisterFunc(s)                                           //	将server放入配置中

	lis, err := net.Listen("tcp", c.Addr) //	 创建监听器
	if err != nil {
		log.Println(constants.Err("Failed to listen : " + c.Addr))
		return s
	}
	go func() {
		err = s.Serve(lis) //	启动 grpc服务
		if err != nil {
			log.Printf(constants.Err(fmt.Sprintf("server %s started error , cause by : %v \n", viper.GetString("grpc.name"), err)))
			return
		}
	}()
	log.Println(constants.Info(fmt.Sprintf("grpc server named %s is registed successfully ...", c.Name)))
	return s
}

const TokenKey = "token_jcpd"
const TokenSecret = "ade52b708082ca98a68f54583fb9d4ef" //  jcpd-love-jichi-pidan

// Auth 用于验证调用者身份
func Auth(ctx context.Context) error {
	//	实际上就是拿到我们需要的 token
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return definition.TokenLack.Err()
	}
	var tokenString string
	if token_, ok := md[TokenKey]; ok {
		tokenString = token_[0]
	}
	//	判断 token是否合法
	if tokenString != TokenSecret {
		return definition.TokenIllegal.Err()
	}
	return nil
}
