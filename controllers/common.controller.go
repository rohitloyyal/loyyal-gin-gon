package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommonController struct {
}

func (controller *CommonController) HealtCheck(ctx *gin.Context) {

	ctx.Header("Access-Control-Allow-Methods", "*")
	ctx.Header("Access-Control-Allow-Headers", "*")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func (controller *CommonController) CommonRoutes(group *gin.RouterGroup) {
	commonRoute := group.Group("/common")

	commonRoute.GET("/healtcheck", controller.HealtCheck)
}
