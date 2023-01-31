package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/middleware"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/nats"
	"github.com/loyyal/loyyal-be-contract/services"
)

type WalletController struct {
	WalletService services.WalletService
	Nats          *nats.Client
}

// constructor calling
func NewWallet(service services.WalletService, nats *nats.Client) WalletController {
	return WalletController{
		WalletService: service,
		Nats:          nats,
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

	// publishing to nats
	if err := controller.Nats.Publish(ctx.Request.Context(), &models.CreateRequest{RefID: wallet.Ref, Amount: wallet.Balance, Channel: wallet.Channel}); err != nil {
		fmt.Print("failed to write wallet to NATS (failing over to retry service): %w", err)
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

func (controller *WalletController) WalletRoutes(group *gin.RouterGroup) {
	walletRoute := group.Group("/wallet")
	walletRoute.Use(middleware.JWTAuthMiddleware())

	walletRoute.GET("/get", controller.WalletGet)
	walletRoute.POST("/create", controller.walletCreate)
	walletRoute.DELETE("/delete", controller.walletDelete)
}
