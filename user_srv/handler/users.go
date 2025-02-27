package handler

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Numsina/tk_users/user_srv/dao"
	domain "github.com/Numsina/tk_users/user_srv/domian"
	"github.com/Numsina/tk_users/user_srv/gen/users/v1"
	"github.com/Numsina/tk_users/user_srv/service"
)

var (
	ErrUniqueConflict = dao.ErrUniqueConflict
	ErrRecordNotFound = dao.ErrRecordNotFound
)

var _ users.UserServiceServer = &UserHandler{}

type UserHandler struct {
	users.UnimplementedUserServiceServer
	srv service.UserService
}

func NewUserHandler(srv service.UserService) *UserHandler {
	return &UserHandler{
		srv: srv,
	}
}

func (u *UserHandler) Register(ctx context.Context, req *users.RegisterReq) (*users.RegisterResp, error) {
	// 按理来说不应该再这里校验参数， 应该在web层进行校验(保险起见可以校验参数)
	// 简单校验一下
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return &users.RegisterResp{}, status.Error(codes.InvalidArgument, "参数无效")
	}

	if req.GetPassword() != req.GetConfirmPassword() {
		return &users.RegisterResp{}, status.Error(codes.InvalidArgument, "两次输入的密码不同")
	}

	id, err := u.srv.SignUp(ctx, domain.User{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})

	if errors.Is(err, ErrUniqueConflict) {
		return &users.RegisterResp{}, status.Error(codes.AlreadyExists, "用户已存在")
	}

	if err != nil {
		return &users.RegisterResp{}, status.Error(codes.Internal, err.Error())
	}

	return &users.RegisterResp{
		UserId: id,
	}, nil
}

func (u *UserHandler) Login(ctx context.Context, req *users.LoginReq) (*users.LoginResp, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	tracer := otel.Tracer("tk_user_srv")
	ctx, span := tracer.Start(ctx, "Register",
		trace.WithAttributes(
			attribute.StringSlice("client-id", md.Get("client-id")),
		),
	)
	defer span.End()
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return &users.LoginResp{}, status.Error(codes.InvalidArgument, "参数无效")
	}

	user, err := u.srv.Login(ctx, domain.User{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})

	if errors.Is(err, ErrRecordNotFound) {
		return &users.LoginResp{}, status.Error(codes.NotFound, "用户不存在")
	}

	if err != nil {
		return &users.LoginResp{}, status.Error(codes.Internal, err.Error())
	}

	return &users.LoginResp{
		UserId: user.Id,
	}, nil
}

func (u *UserHandler) GetUserByEmail(ctx context.Context, req *users.GetUserByEmailReq) (*users.GetUserByEmailResp, error) {
	if req.GetEmail() == "" {
		return &users.GetUserByEmailResp{}, status.Error(codes.InvalidArgument, "参数无效")
	}

	user, err := u.srv.GetUserInfoByEmail(ctx, req.GetEmail())

	if err == ErrRecordNotFound {
		return &users.GetUserByEmailResp{}, status.Error(codes.NotFound, "用户不存在")
	}

	if err != nil {
		return &users.GetUserByEmailResp{}, status.Error(codes.Internal, err.Error())
	}

	return &users.GetUserByEmailResp{
		Email:       user.Email,
		NickName:    user.NickName,
		BirthDay:    user.BirthDay,
		Address:     user.Address,
		Description: user.Description,
		Avatar:      user.Avatar,
	}, nil
}
