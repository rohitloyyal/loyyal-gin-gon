package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/middleware"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/services"
	"github.com/loyyal/loyyal-be-contract/utils/notification"
)

type WalletController struct {
	WalletService services.WalletService
}

// constructor calling
func NewWallet(service services.WalletService) WalletController {
	return WalletController{
		WalletService: service,
	}
}

func (controller *WalletController) walletCreate(ctx *gin.Context) {
	// get data from body
	var wallet models.Wallet
	if err := ctx.ShouldBindJSON(&wallet); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if wallet.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "name is required",
		})
		return
	}

	err := controller.WalletService.Create(&wallet, "", 0)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    wallet,
	})
}

func (controller *WalletController) WalletGet(ctx *gin.Context) {
	// get data from body
	walletId := ctx.Query("walletId")
	if walletId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "wallet id is required",
		})
		return
	}

	wallet, err := controller.WalletService.Get(walletId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    wallet,
	})
}

func (controller *WalletController) walletDelete(ctx *gin.Context) {
	var wallet models.Wallet
	if err := ctx.ShouldBindJSON(&wallet); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if wallet.Identifier == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "identifier is required",
		})
		return
	}

	err := controller.WalletService.Delete(wallet.Identifier, "admin")
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

func (controller *WalletController) SendEmail(ctx *gin.Context) {
	// get data from body
	go notification.SendEmailNotification()
	// if err != nil {
	// 	ctx.JSON(http.StatusBadRequest, gin.H{
	// 		"message": err.Error(),
	// 	})
	// 	return
	// }

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func (controller *WalletController) WalletRoutes(group *gin.RouterGroup) {
	walletRoute := group.Group("/wallet")
	walletRoute.Use(middleware.JWTAuthMiddleware())

	walletRoute.GET("/get", controller.WalletGet)
	walletRoute.POST("/create", controller.walletCreate)
	walletRoute.DELETE("/delete", controller.walletDelete)
}
