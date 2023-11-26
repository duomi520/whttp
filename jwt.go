package whttp

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWT
type JWT struct {
	TokenSigningKey []byte
	TokenExpires    time.Duration
}

// TokenParse 令牌解析
func (j JWT) TokenParse(tokenString string) (jwt.MapClaims, error) {
	if len(tokenString) == 0 {
		return nil, errors.New("token is nil")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return j.TokenSigningKey, nil
	})
	if err != nil {
		return nil, errors.New(err.Error())
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("authentication failed")
}

// CreateToken 生成
func (j JWT) CreateToken(key []string, obj []any) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	//设置claims
	claims := token.Claims.(jwt.MapClaims)
	//设置有效期限
	claims["exp"] = time.Now().Add(j.TokenExpires).Unix()
	claims["iat"] = time.Now().Unix()
	for i := range key {
		claims[key[i]] = obj[i]
	}
	s, err := token.SignedString(j.TokenSigningKey)
	if err != nil {
		return s, fmt.Errorf("create token faile: %w", err)
	}
	return s, nil
}

// JWTMiddleware 中间件
func (j JWT) JWTMiddleware() func(*HTTPContext) {
	return func(c *HTTPContext) {
		tokenString := c.Request.Header.Get("Authorization")
		if len(tokenString) == 0 {
			c.String(http.StatusUnauthorized, "缺少令牌")
			return
		}
		claims, err := j.TokenParse(tokenString)
		if err != nil {
			c.String(http.StatusUnauthorized, "令牌无效")
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}

// https://zhuanlan.zhihu.com/p/113376580

// authZ 授权
// https://mp.weixin.qq.com/s/ubSfSAT7kVmnCAZly13F0Q
