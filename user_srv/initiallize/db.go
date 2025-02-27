package initiallize

import (
	"fmt"
	"github.com/Numsina/tk_users/user_srv/pkg/gormx"
	"gorm.io/plugin/prometheus"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func InitDB() *gorm.DB {
	if db == nil {
		dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			Conf.MysqlInfo.UserName, Conf.MysqlInfo.PassWord, Conf.MysqlInfo.Host, Conf.MysqlInfo.Port,
			Conf.MysqlInfo.DBName)

		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second, // Slow SQL threshold
				LogLevel:      logger.Info, // Log level
			},
		)

		var err error
		db, err = gorm.Open(mysql.New(mysql.Config{
			DSN:                       dsn,
			SkipInitializeWithVersion: true,
		}), &gorm.Config{
			Logger:                 newLogger,
			SkipDefaultTransaction: true,
		})

		if err != nil {
			panic(err)
		}
		//gormx.InitJaeger()
		use(db)
	}
	return db
}

func use(db *gorm.DB) {
	// 监控mysql线程的运行数量
	db.Use(prometheus.New(prometheus.Config{
		DBName:          "tk_user_srv",
		RefreshInterval: 15,
		StartServer:     false,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	}))

	// 利用prometheus监控sql执行时长
	db.Use(gormx.NewCallbacks())
	// jaeger 埋点
	db.Use(gormx.NewJaegerTracer())
}

//
//
//}
