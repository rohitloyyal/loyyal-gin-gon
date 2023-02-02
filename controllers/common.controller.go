package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/utils/common"
)

type CommonController struct {
}

func (controller *CommonController) HealtCheck(ctx *gin.Context) {

	ctx.Header("Access-Control-Allow-Methods", "*")
	ctx.Header("Access-Control-Allow-Headers", "*")
	ctx.Header("Access-Control-Allow-Origin", "*")
	common.PrepareCustomResponse(ctx, "success", nil)
}

func (controller *CommonController) CommonRoutes(group *gin.RouterGroup) {
	commonRoute := group.Group("/common")

	commonRoute.GET("/healtcheck", controller.HealtCheck)
}
