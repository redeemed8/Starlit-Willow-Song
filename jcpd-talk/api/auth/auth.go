package auth

import "context"

type Authentication struct {
	Token string
}

const (
	TokenKey    = "token_jcpd"
	TokenSecret = "ade52b708082ca98a68f54583fb9d4ef" //  jcpd-love-jichi-pidan
)

// GetRequestMetadata 用于客户端的返回
func (a *Authentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{TokenKey: a.Token}, nil
}

// RequireTransportSecurity 是否开启安全连接
func (a *Authentication) RequireTransportSecurity() bool {
	return false
}
