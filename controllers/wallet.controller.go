package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/middleware"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/nats"
	"github.com/loyyal/loyyal-be-contract/services"
	"github.com/loyyal/loyyal-be-contract/utils/common"
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

type WalletCreateRequest struct {
	WalletType   string          `json:"walletType" binding:"required"`
	Name         string          `json:"name" binding:"required"`
	Metadata     json.RawMessage `json:"metadata"`
	PreLoadValue int64           `json:"preLoadValue"`
	LinkedTo     string          `json:"linkedTo" binding:"required"`
}

func (controller *WalletController) walletCreate(ctx *gin.Context) {
	fName := "controller/wallet/create"
	var request WalletCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%d ", err))
		return
	}

	var wallet models.Wallet
	wallet.Name = request.Name
	wallet.Metadata = request.Metadata
	wallet.WalletType = request.WalletType

	err := controller.WalletService.Create(&wallet, request.LinkedTo, request.PreLoadValue)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%d ", err))
		return
	}

	// publishing to nats
	if err := controller.Nats.Publish(ctx.Request.Context(), &models.CreateRequest{RefID: wallet.Ref, Amount: wallet.Balance, Channel: wallet.Channel}); err != nil {
		fmt.Print("failed to write wallet to NATS (failing over to retry service): %w", err)
		// TODO: execute the retry flow from here
	}

	common.PrepareCustomResponse(ctx, "wallet created", struct {
		Identifier string `json:"identifier"`
	}{Identifier: wallet.Identifier})
}

func (controller *WalletController) walletGet(ctx *gin.Context) {
	fName := "controller/wallet/get"
	walletId := ctx.Query("walletId")
	if walletId == "" {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: wallet id is required", fmt.Sprintf("got :%s ", walletId))
		return
	}

	wallet, err := controller.WalletService.Get(walletId)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "wallet fetched", wallet)
}

func (controller *WalletController) walletDelete(ctx *gin.Context) {
	fName := "controller/wallet/delete"
	var wallet models.Wallet
	if err := ctx.ShouldBindJSON(&wallet); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: invalid request payload provided", fmt.Sprintf("got :%s ", err))
		return
	}

	if wallet.Identifier == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "identifier is required",
		})
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: wallet identifier is required", fmt.Sprintf("got :%s ", wallet.Identifier))
		return
	}

	err := controller.WalletService.Delete(wallet.Identifier, "admin")
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "wallet deleted", nil)
}

func (controller *WalletController) walletFilter(ctx *gin.Context) {
	fName := "controller/wallet/filter"
	transactions, err := controller.WalletService.Filter("AND isDeleted=false", map[string]interface{}{}, "createdAt", -1)

	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "transactions fetched", transactions)
}

func (controller *WalletController) WalletRoutes(group *gin.RouterGroup) {
	walletRoute := group.Group("/wallet")

	walletRoute.Use(middleware.JWTAuthMiddleware())

	walletRoute.GET("/get", controller.walletGet)
	walletRoute.POST("/filter", controller.walletFilter)
	walletRoute.POST("/create", controller.walletCreate)
	walletRoute.DELETE("/delete", controller.walletDelete)
}
