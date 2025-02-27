package middleware

import (
	"encoding/gob"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type LoginJWTMiddleWareBuilder struct {
	paths []string
	jhl   *JWT
	key   []byte
}

func NewLoginJWTMiddleWareBuilder(jhl *JWT) *LoginJWTMiddleWareBuilder {
	return &LoginJWTMiddleWareBuilder{
		jhl: jhl,
	}
}

func (l *LoginJWTMiddleWareBuilder) IngorePaths(path ...string) *LoginJWTMiddleWareBuilder {
	l.paths = append(l.paths, path...)
	return l
}

func (l *LoginJWTMiddleWareBuilder) Build() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		// 跳过不需要检查的路径
		for _, v := range l.paths {
			if ctx.Request.URL.Path == v || strings.Contains(ctx.Request.URL.Path, "/swagger") {
				return
			}
		}

		auth := ctx.GetHeader("x-jwt-token")

		claims := &UserClaims{}
		token, err := l.jhl.ParseToken(ctx, auth, claims)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userAgent := ctx.Request.Header.Get("User-Agent")
		if claims.UserAgent != userAgent {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now()
		if claims.ExpiresAt.Sub(now) <= l.jhl.MaxRefresh {
			newClaims, tokenString, err := l.jhl.RefreshToken(ctx, token, claims)
			if err == nil {
				ctx.Header("x-jwt-token", tokenString)
				claims = newClaims
			} else {
				log.Println("刷新token失败")
			}
		}
		ctx.Set("claims", claims)
	}
}
