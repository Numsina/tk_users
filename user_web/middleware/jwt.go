package middleware

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"github.com/Numsina/tk_users/user_web/initialize"
)

var (
	ErrTokenGenFailed         = errors.New("令牌生成失败")
	ErrTokenExpired           = errors.New("令牌已过期")
	ErrTokenExpiredMaxRefresh = errors.New("令牌已过最大刷新时间")
	ErrRefreshToken           = errors.New("刷新令牌失败")
	ErrTokenInvalid           = errors.New("请求令牌无效")
	ErrTokenNotFound          = errors.New("无法找到令牌")
	ErrSsidGenFailed          = errors.New("生成ssid失败")
	ErrSsidValid              = errors.New("ssid无效")
	ErrSsidExpired            = errors.New("ssid已过期")
)

type UserClaims struct {
	jwt.RegisteredClaims
	UserId    int32
	UserAgent string
	Ssid      string
}

type JWT struct {
	Secret      []byte
	Expire      time.Duration
	MaxRefresh  time.Duration
	RedisClient redis.Cmdable
}

func NewJWT(key []byte) *JWT {
	return &JWT{
		Secret:     key,
		Expire:     time.Hour * 12,
		MaxRefresh: time.Hour * 6,
		RedisClient: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", initialize.Conf.RedisInfo.Host, initialize.Conf.RedisInfo.Port),
			Password: initialize.Conf.RedisInfo.PassWord,
		}),
	}
}

func (j *JWT) SetToken(ctx *gin.Context, id int32, ssid string) (string, error) {
	userClaims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.Expire)),
		},
		UserId:    id,
		UserAgent: ctx.GetHeader("User-Agent"),
		Ssid:      ssid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, userClaims)
	tokenString, err := token.SignedString(j.Secret)
	if err != nil {
		log.Println("生成token失败", err)
		return "", ErrTokenGenFailed
	}

	err = j.RedisClient.Set(ctx.Request.Context(), strconv.Itoa(int(id)), ssid, j.Expire).Err()
	if err != nil {
		return "", ErrSsidGenFailed
	}
	return tokenString, nil
}

func (j *JWT) ParseToken(ctx *gin.Context, tokenString string, claims *UserClaims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return j.Secret, nil
	})

	if err != nil {
		log.Println("解析令牌发生错误，错误原因为：", err)
		return nil, ErrTokenNotFound
	}

	if token == nil || !token.Valid {
		return nil, ErrTokenInvalid
	}

	val, err := j.RedisClient.Get(ctx, strconv.Itoa(int(claims.UserId))).Result()
	if err != nil {
		return nil, ErrSsidExpired
	}

	if val != claims.Ssid {
		return nil, ErrSsidValid
	}
	return token, nil
}

func (j *JWT) RefreshToken(ctx *gin.Context, token *jwt.Token, claims *UserClaims) (*UserClaims, string, error) {
	var tokenString string
	var err error
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(j.Expire))
	tokenString, err = token.SignedString(j.Secret)
	if err != nil {
		log.Println("token 刷新失败, 错误原因：", err)
		return nil, "", ErrRefreshToken
	}

	err = j.RedisClient.Expire(ctx, strconv.Itoa(int(claims.UserId)), j.Expire).Err()
	if err != nil {
		return nil, "", ErrSsidGenFailed
	}
	return claims, tokenString, nil
}

func (j *JWT) DeleteSsid(ctx *gin.Context, claims *UserClaims) error {
	return j.RedisClient.Del(ctx, strconv.Itoa(int(claims.UserId))).Err()
}
