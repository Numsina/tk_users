package dao

import (
	"log"

	"gorm.io/gorm"
)

type User struct {
	Id          int32  `gorm:"primaryKey, autoIncrement"`
	Email       string `gorm:"unique"`
	Password    string
	NickName    string
	Description string
	Avatar      string
	Address     string
	BirthDay    int64
	CreateAt    int64
	UpdateAt    int64
	DeleteAt    int64
}

func InitAutoMigrateTable(db *gorm.DB) error {
	err := db.AutoMigrate(&User{})
	if err != nil {
		log.Printf("迁移表失败, 失败原因：%v", err)
		return err
	}
	return nil
}
