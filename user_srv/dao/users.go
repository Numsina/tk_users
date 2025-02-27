package dao

import (
	"context"
	"errors"
	"time"

	"github.com/Numsina/tk_users/user_srv/logger"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var ErrRecordNotFound = errors.New("记录不存在")
var ErrUniqueConflict = errors.New("唯一主键冲突")

type UserI interface {
	CreateUser(ctx context.Context, user User) (int32, error)
	UpdateUserInfoByUid(ctx context.Context, user User) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	DeleteUser(ctx context.Context, uid int32) error
}

var _ UserI = &user{}

type user struct {
	db     *gorm.DB
	logger *logger.Logger
}

func NewUserDao(db *gorm.DB, logger *logger.Logger) UserI {
	return &user{
		db:     db,
		logger: logger,
	}
}

func (u *user) CreateUser(ctx context.Context, user User) (int32, error) {
	now := time.Now().UnixMilli()
	user.CreateAt = now
	user.UpdateAt = now
	err := u.db.WithContext(ctx).Create(&user).Error
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		const uniqueConflictErr uint16 = 1062
		if mysqlErr.Number == uniqueConflictErr {
			u.logger.Sugar().Infof("唯一主键冲突, 冲突主键: %s", user.Email)
			return 0, ErrUniqueConflict
		}
	}

	if err != nil {
		u.logger.Sugar().Warnf("数据库错误, 错误原因: %s", err)
	}
	return user.Id, nil
}

func (u *user) DeleteUser(ctx context.Context, uid int32) error {
	err := u.db.WithContext(ctx).Delete(&User{Id: uid}).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRecordNotFound
	}

	if err != nil {
		u.logger.Sugar().Info("删除用户失败, 造成数据库错误, 错误原因: %v", err)
		return err
	}
	return nil
}

func (u *user) UpdateUserInfoByUid(ctx context.Context, user User) (User, error) {
	now := time.Now().UnixMilli()
	user.CreateAt = now
	user.UpdateAt = now
	err := u.db.WithContext(ctx).Where("id = ?", user.Id).Updates(&user).Error

	if err == gorm.ErrRecordNotFound {
		return User{}, ErrRecordNotFound
	}

	if err != nil {
		// 可能是数据库错误， 记录日志，
		u.logger.Sugar().Warnf("数据库错误, 错误原因: %s", err)
	}

	return user, nil
}

func (u *user) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var ue User
	err := u.db.WithContext(ctx).Where("email = ?", email).First(&ue).Error
	if err == gorm.ErrRecordNotFound {
		return User{}, ErrRecordNotFound
	}

	if err != nil {
		u.logger.Sugar().Warnf("数据库内部错误, 错误原因：%s", err)
		return ue, err
	}

	return ue, nil
}
