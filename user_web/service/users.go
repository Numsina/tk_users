package service

import (
	"context"
	"log"

	"github.com/Numsina/tk_users/user_web/domain"
	"github.com/Numsina/tk_users/user_web/gen/users/v1"
)

type UserService struct {
	client users.UserServiceClient
}

func NewService(client users.UserServiceClient) *UserService {
	return &UserService{client: client}
}

func (u *UserService) Signup(ctx context.Context, user domain.User) (int32, error) {
	resp, err := u.client.Register(ctx, &users.RegisterReq{
		Email:           user.Email,
		Password:        user.Password,
		ConfirmPassword: user.ConfirmPassword,
	})
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return resp.GetUserId(), nil
}

func (u *UserService) Login(ctx context.Context, user domain.User) (int32, error) {
	resp, err := u.client.Login(ctx, &users.LoginReq{
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		return 0, err
	}
	return resp.GetUserId(), nil
}

func (u *UserService) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	resp, err := u.client.GetUserByEmail(ctx, &users.GetUserByEmailReq{
		Email: email,
	})

	if err != nil {
		// 打点||打日志
		return domain.User{}, err
	}
	return domain.User{
		Email:       resp.Email,
		NickName:    resp.NickName,
		Description: resp.Description,
		Avatar:      resp.Avatar,
		BirthDay:    resp.BirthDay,
		Address:     resp.Address,
	}, err
}
