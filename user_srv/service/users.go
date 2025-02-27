package service

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/Numsina/tk_users/user_srv/dao"
	domain "github.com/Numsina/tk_users/user_srv/domian"
	logger "github.com/Numsina/tk_users/user_srv/logger"
)

var (
	ErrUniqueConflict = dao.ErrUniqueConflict
	ErrRecordNotFound = dao.ErrRecordNotFound
)

type UserService interface {
	SignUp(ctx context.Context, user domain.User) (int32, error)
	Login(ctx context.Context, user domain.User) (domain.User, error)
	Delele(ctx context.Context, uid int32) error
	ModifyUserInfoById(ctx context.Context, user domain.User) (domain.User, error)
	GetUserInfoByEmail(ctx context.Context, email string) (domain.User, error)
}

var _ UserService = &userSvc{}

type userSvc struct {
	d      dao.UserI
	logger *logger.Logger
}

func NewUserSvc(d dao.UserI, logger *logger.Logger) UserService {
	return &userSvc{
		d:      d,
		logger: logger,
	}
}

func (u *userSvc) SignUp(ctx context.Context, user domain.User) (int32, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		// 记录日志
		u.logger.Sugar().Infof("bcrypt加密失败, 失败原因：%s", err)
		return 0, err
	}
	user.Password = string(hash)
	return u.d.CreateUser(ctx, dao.User{
		Email:       user.Email,
		Password:    user.Password,
		NickName:    user.NickName,
		BirthDay:    user.BirthDay,
		Address:     user.Address,
		Description: user.Description,
		Avatar:      user.Avatar,
	})
}

func (u *userSvc) Login(ctx context.Context, user domain.User) (domain.User, error) {
	ue, err := u.d.FindUserByEmail(ctx, user.Email)
	if err != nil {
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(ue.Password), []byte(user.Password))
	return domain.User{
		Id:          ue.Id,
		Email:       ue.Email,
		NickName:    ue.NickName,
		BirthDay:    ue.BirthDay,
		Address:     ue.Address,
		Description: ue.Description,
		Avatar:      ue.Avatar,
	}, err
}

func (u *userSvc) Delele(ctx context.Context, uid int32) error {
	return u.d.DeleteUser(ctx, uid)
}

func (u *userSvc) ModifyUserInfoById(ctx context.Context, user domain.User) (domain.User, error) {
	if user.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			// 记录日志
			u.logger.Sugar().Infof("修改密码加密失败, 失败原因：%s", err)
			return domain.User{}, err
		}
		user.Password = string(hash)
	}
	ue, err := u.d.UpdateUserInfoByUid(ctx, dao.User{
		Id:          user.Id,
		Email:       user.Email,
		NickName:    user.NickName,
		BirthDay:    user.BirthDay,
		Address:     user.Address,
		Description: user.Description,
		Avatar:      user.Avatar,
	})
	return domain.User{
		Id:          ue.Id,
		Email:       ue.Email,
		NickName:    ue.NickName,
		BirthDay:    ue.BirthDay,
		Address:     ue.Address,
		Description: ue.Description,
		Avatar:      ue.Avatar,
	}, err
}

func (u *userSvc) GetUserInfoByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := u.d.FindUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Email:       user.Email,
		NickName:    user.NickName,
		BirthDay:    user.BirthDay,
		Address:     user.Address,
		Description: user.Description,
		Avatar:      user.Avatar,
	}, nil
}
