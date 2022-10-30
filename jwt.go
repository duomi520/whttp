package whttp

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

//JWT
type JWT struct {
	TokenSigningKey []byte
	TokenExpires    time.Duration
}

//JWTMiddleware 中间件
func (j JWT) JWTMiddleware() func(*HTTPContext) {
	return func(c *HTTPContext) {
		tokenString := c.Request.Header.Get("Authorization")
		if len(tokenString) == 0 {
			c.String(http.StatusUnauthorized, "token need")
			return
		}
		c.Next()
	}
}

//TokenParse 令牌解析
func (j JWT) TokenParse(tokenString string) (any, error) {
	if len(tokenString) == 0 {
		return 0, errors.New("token is nil")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return j.TokenSigningKey, nil
	})
	if err != nil {
		return 0, errors.New(err.Error())
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		data := claims["object"]
		return data, nil
	}
	return 0, errors.New("authentication failed")
}

//CreateToken 生成
func (j JWT) CreateToken(obj any) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	//设置claims
	claims := token.Claims.(jwt.MapClaims)
	//设置有效期限
	claims["exp"] = time.Now().Add(j.TokenExpires).Unix()
	claims["iat"] = time.Now().Unix()
	claims["object"] = obj
	s, err := token.SignedString(j.TokenSigningKey)
	if err != nil {
		return s, fmt.Errorf("create token faile: %w", err)
	}
	return s, nil
}
