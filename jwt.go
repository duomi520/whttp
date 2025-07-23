package whttp

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"maps"
)

// JWT
type JWT struct {
	TokenSigningKey []byte
	TokenExpires    time.Duration
}

// TokenParse 令牌解析
func (j JWT) TokenParse(tokenString string) (jwt.MapClaims, error) {
	if len(tokenString) == 0 {
		return nil, errors.New("invalid token")
	}
	// 指定签名算法并验证
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.TokenSigningKey, nil
	})
	// 细化错误处理
	if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return nil, errors.New("malformed token")
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			return nil, errors.New("token expired")
		}
	} else if err != nil {
		return nil, fmt.Errorf("token parse failed: %w", err)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token claims")
}

// CreateToken 生成
func (j JWT) CreateToken(claims map[string]any) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	//设置claims
	tokenClaims := token.Claims.(jwt.MapClaims)
	// 添加自定义声明
	maps.Copy(tokenClaims, claims)
	//设置有效期限
	now := time.Now()
	tokenClaims["exp"] = now.Add(j.TokenExpires).Unix()
	tokenClaims["iat"] = now.Unix()
	// 生成签名
	signedToken, err := token.SignedString(j.TokenSigningKey)
	if err != nil {
		return signedToken, fmt.Errorf("signing failed: %w", err)
	}
	return signedToken, nil
}

// RefreshToken 刷新令牌
func (j JWT) RefreshToken(c *HTTPContext) error {
	claims, exists := c.Get("jwt_claims")
	if !exists {
		return errors.New("no claims found")
	}
	newToken, err := j.CreateToken(claims.(jwt.MapClaims))
	if err != nil {
		return errors.New("token refresh failed")
	}
	c.Writer.Header().Set("Authorization", newToken)
	return nil
}

// JWTMiddleware 中间件 在Header "Authorization" 设置令牌

func (j JWT) JWTMiddleware(requiredClaims ...string) func(*HTTPContext) {
	return func(c *HTTPContext) {
		authHeader := c.Request.Header.Get("Authorization")
		if len(authHeader) == 0 {
			c.String(http.StatusUnauthorized, "缺少令牌")
			return
		}
		claims, err := j.TokenParse(authHeader)
		if err != nil {
			c.String(http.StatusUnauthorized, "令牌无效")
			return
		}
		for _, v := range requiredClaims {
			if a, ok := claims[v]; !ok {
				c.String(http.StatusUnauthorized, "令牌缺失必要信息")
				return
			} else {
				c.Set(v, a)
			}
		}
		// 设置上下文供后续使用
		c.Set("jwt_claims", claims)
		c.Next()
	}
}

// https://zhuanlan.zhihu.com/p/113376580

// authZ 授权
// https://mp.weixin.qq.com/s/ubSfSAT7kVmnCAZly13F0Q
