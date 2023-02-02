package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/services"
)

type AuthController struct {
	UserService services.UserService
}

type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}


func NewAuthController(service services.UserService) AuthController {
	return AuthController{
		UserService: service,
	}
}

func (controller *AuthController) Login(ctx *gin.Context) {

	var input RegisterInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var user models.User
	user.Username = input.Username
	user.Password = input.Password

	token, err := controller.UserService.Login(&user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    token,
	})
}

func (controller *AuthController) Register(ctx *gin.Context) {

	var input RegisterInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var user models.User
	user.Username = input.Username
	user.Password = input.Password

	err := controller.UserService.Register(&user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func (controller *AuthController) AuthRoutes(group *gin.RouterGroup) {
	contractRoute := group.Group("/auth")

	contractRoute.POST("/login", controller.Login)
	contractRoute.POST("/register", controller.Register)
}
