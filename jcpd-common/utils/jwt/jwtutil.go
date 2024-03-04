package commonJWT

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"time"
)

const (
	ExpireTime_ = 24 * time.Hour * 30 //	一个月
	JWTKey      = "sd789r7gt9eb44hy874jt9b4f47uk4"
)

// MakeToken 获取 token
func MakeToken(userClaim UserClaims) (string, error) {
	expireTime := time.Now().Add(ExpireTime_)
	claims := &Claims{
		UserClaim: userClaim,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, tokenErr := token.SignedString([]byte(JWTKey))
	if tokenErr != nil {
		log.Println("Failed to make token , cause by : ", tokenErr)
		return "token err", tokenErr
	}
	return tokenString, nil
}

const (
	TokenHeader   = "X-auth"
	TokenPathArgs = "auth"
	TableName     = "5613_userinfo"
)

var NotLoginError = errors.New("未登录或登录已过期")
var DBException = errors.New("查询数据库异常")

var DB *gorm.DB

func NewDB(db *gorm.DB) {
	DB = db
	if DB != nil {
		log.Println("JWT mysql connection is inited ...")
	}
}

func ParseToken(ctx *gin.Context) (UserClaims, error) {
	//	从请求头中获取 tokenString
	var tokenString string
	tokenString = ctx.Request.Header.Get(TokenHeader)
	//	如果请求头中不存在，则查询路径参数
	if tokenString == "" {
		tokenString = ctx.Query(TokenPathArgs)
	}
	//	还不存在的话就返回错误
	if tokenString == "" {
		return UserClaims{}, NotLoginError
	}
	//	存在的话，进行解析
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) { return []byte(JWTKey), nil })
	if err != nil || !token.Valid {
		return UserClaims{}, NotLoginError
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return UserClaims{}, NotLoginError
	}
	//	判断 身份信息 是否真实
	id := claims.UserClaim.Id
	type UUID_ struct {
		Username string
		UUID     string
	}

	var uuid_ UUID_

	//	此处方法最后应更换为用 grpc和 proto文件 来调用 user模块中的方法来获取 uuid和 username
	err1 := DB.Table(TableName+" a").Select("a.uuid,a.username").Where("a.id = ?", id).First(&uuid_).Error

	if err1 != nil && !errors.Is(err1, gorm.ErrRecordNotFound) {
		log.Println("查询数据库 uuid异常 , cause by : ", err1)
		//	TODO 放入消息队列进行通知 ...
		//  ...
		return UserClaims{}, DBException
	}
	if uuid_.UUID != claims.UserClaim.UUID || uuid_.Username != claims.UserClaim.Username {
		return UserClaims{}, NotLoginError
	}
	return UserClaims{Id: id, Username: claims.UserClaim.Username, UUID: claims.UserClaim.UUID}, nil
}
