package initialize

import "github.com/Numsina/tk_users/user_web/logger"

var l *logger.Logger

func InitLogger() *logger.Logger {
	if l == nil {
		return logger.NewLogger()
	}
	return l
}
