package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Numsina/tk_users/user_web/tools"
)

func checkError(err error, ctx *gin.Context) {
	if err != nil {
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.AlreadyExists:
				ctx.JSON(http.StatusNotFound, tools.Result{
					Code: int(s.Code()),
					Msg:  err.Error(),
				})
			case codes.NotFound:
				ctx.JSON(http.StatusNotFound, tools.Result{
					Code: int(s.Code()),
					Msg:  "未发现该资源",
				})
			case codes.Internal:
				ctx.JSON(http.StatusInternalServerError, tools.Result{
					Code: int(s.Code()),
					Msg:  err.Error(),
				})

			case codes.DeadlineExceeded:
				ctx.JSON(http.StatusRequestTimeout, tools.Result{
					Code: int(s.Code()),
					Msg:  "请求资源超时",
				})
			case codes.InvalidArgument:
				ctx.JSON(http.StatusBadRequest, tools.Result{
					Code: int(s.Code()),
					Msg:  "请求参数错误",
				})
			default:
				ctx.JSON(http.StatusInternalServerError, tools.Result{
					Code: int(s.Code()),
					Msg:  err.Error(),
				})
			}
		}
		return
	}
}
