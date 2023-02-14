package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/services"
	"github.com/loyyal/loyyal-be-contract/utils/common"
	"go.opentelemetry.io/otel"
)

type AuthController struct {
	IdentityService services.IdentityService
}

type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func NewAuthController(service services.IdentityService) AuthController {
	return AuthController{
		IdentityService: service,
	}
}

func (controller *AuthController) Login(ctx *gin.Context) {
	fName := "controllers/authController/login"
	tracer := otel.Tracer("Login")
	_, span := tracer.Start(ctx.Request.Context(), fName)
	defer span.End()

	var input RegisterInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: invalid request body provided", fmt.Sprintf("got :%s ", input))
		return
	}

	var user models.Identity
	user.Username = input.Username
	user.Password = input.Password

	token, err := controller.IdentityService.Login(ctx.Request.Context(), &user)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "logged in successfully", struct {
		Token string `json:"token"`
	}{Token: token})
}

// func (controller *AuthController) Register(ctx *gin.Context) {
// 	fName := "controllers/authController/register"
// 	var input RegisterInput
// 	if err := ctx.ShouldBindJSON(&input); err != nil {
// 		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: invalid request body provided", fmt.Sprintf("got :%s ", input))

// 		return
// 	}

// 	var user models.User
// 	user.Username = input.Username
// 	user.Password = input.Password

// 	err := controller.UserService.Register(&user)
// 	if err != nil {
// 		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
// 		return
// 	}

// 	common.PrepareCustomResponse(ctx, "registered successfully", nil)
// }

func (controller *AuthController) AuthRoutes(group *gin.RouterGroup) {
	contractRoute := group.Group("/auth")

	contractRoute.POST("/login", controller.Login)
	// contractRoute.POST("/register", controller.Register)
}
