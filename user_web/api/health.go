package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "ok")
	return
}
