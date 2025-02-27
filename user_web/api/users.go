package api

import (
	"net/http"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Numsina/tk_users/user_web/constant"
	"github.com/Numsina/tk_users/user_web/domain"
	"github.com/Numsina/tk_users/user_web/logger"
	"github.com/Numsina/tk_users/user_web/middleware"
	"github.com/Numsina/tk_users/user_web/service"
	"github.com/Numsina/tk_users/user_web/tools"
)

type UserHandler struct {
	emailRegexp    *regexp.Regexp
	passwordRegexp *regexp.Regexp
	svc            *service.UserService
	logger         *logger.Logger
	jhl            *middleware.JWT
}

func NewUserHandler(svc *service.UserService, logger *logger.Logger, jhl *middleware.JWT) *UserHandler {
	return &UserHandler{
		emailRegexp:    regexp.MustCompile(constant.UserEmail, regexp.None),
		passwordRegexp: regexp.MustCompile(constant.UserPassword, regexp.None),
		svc:            svc,
		logger:         logger,
		jhl:            jhl,
	}
}

func (u *UserHandler) RegisterRouters(router *gin.Engine) {
	router.GET("/health", Health)
	userGroup := router.Group("/v1/users")
	{
		userGroup.POST("/signup", u.signup)
		userGroup.POST("/login", u.login)
		userGroup.POST("/logout", u.logout)
		userGroup.GET("/info", u.getUserByEmail)
	}
}

func (u *UserHandler) signup(ctx *gin.Context) {
	var user domain.User
	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, tools.Result{
			Code: 3,
			Msg:  "参数错误",
		})
		return
	}

	if ok, _ := u.emailRegexp.MatchString(user.Email); !ok {
		ctx.JSON(http.StatusBadRequest, tools.Result{
			Code: 3,
			Msg:  "邮箱格式有误!!!",
			Data: nil,
		})
		return
	}

	if ok, _ := u.passwordRegexp.MatchString(user.Password); !ok {
		ctx.JSON(http.StatusBadRequest, tools.Result{
			Code: 3,
			Msg:  "密码格式有误!!!",
			Data: nil,
		})
		return
	}

	if user.Password != user.ConfirmPassword {
		ctx.JSON(http.StatusOK, tools.Result{
			Code: 3,
			Msg:  "两次输入密码不一致!!!",
			Data: nil,
		})
		return
	}

	uid, err := u.svc.Signup(ctx.Request.Context(), domain.User{
		Email:           user.Email,
		Password:        user.Password,
		ConfirmPassword: user.ConfirmPassword,
	})
	if err != nil {
		checkError(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, tools.Result{
		Code: 0,
		Msg:  "用户创建成功",
		Data: uid,
	})
	return
}

func (u *UserHandler) login(ctx *gin.Context) {
	type login_req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req login_req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, "参数错误")
		return
	}

	if ok, _ := u.emailRegexp.MatchString(req.Email); !ok {
		ctx.JSON(http.StatusBadRequest, tools.Result{
			Code: 3,
			Msg:  "邮箱或密码不正确",
			Data: nil,
		})
		return
	}

	if ok, _ := u.passwordRegexp.MatchString(req.Password); !ok {
		ctx.JSON(http.StatusBadRequest, tools.Result{
			Code: 3,
			Msg:  "邮箱或密码不正确",
			Data: nil,
		})
		return
	}

	id, err := u.svc.Login(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		checkError(err, ctx)
		return
	}

	// 设置token或者cookie
	// 生成session，并设置token
	uid := uuid.New()
	tokenString, err := u.jhl.SetToken(ctx, id, uid.String())
	if err != nil {
		u.logger.Error("生成token失败")
		ctx.JSON(http.StatusOK, "登录失败")
		return
	}

	ctx.Header("x-jwt-token", tokenString)

	ctx.JSON(http.StatusOK, tools.Result{
		Code: 0,
		Msg:  "登陆成功",
		Data: id,
	})
	return
}

func (u *UserHandler) logout(ctx *gin.Context) {
	claims := ctx.Value("claims").(*middleware.UserClaims)
	err := u.jhl.DeleteSsid(ctx, claims)
	if err != nil {
		u.logger.Info("删除ssid失败")
		ctx.JSON(http.StatusOK, "退出登录失败")
		return
	}
	ctx.JSON(http.StatusOK, "退出成功")
	return
}

func (u *UserHandler) getUserByEmail(ctx *gin.Context) {
	var email = ctx.Query("email")
	if ok, _ := u.emailRegexp.MatchString(email); !ok {
		ctx.JSON(http.StatusOK, "邮箱格式有误!!!")
		return
	}

	user, err := u.svc.GetUserByEmail(ctx.Request.Context(), email)
	if err != nil {
		checkError(err, ctx)
		return
	}
	claims := ctx.Value("claims").(*middleware.UserClaims)
	user.Id = claims.UserId
	ctx.JSON(http.StatusOK, tools.Result{
		Code: 0,
		Msg:  "查询成功",
		Data: user,
	})
	return
}
