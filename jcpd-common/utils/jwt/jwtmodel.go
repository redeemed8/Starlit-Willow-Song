package commonJWT

import "github.com/dgrijalva/jwt-go"

type UserClaims struct {
	Id   uint32 `json:"id"`
	UUID string `json:"uuid"`
}

type Claims struct {
	UserClaim UserClaims
	jwt.StandardClaims
}
