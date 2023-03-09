package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/middleware"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/nats"
	"github.com/loyyal/loyyal-be-contract/services"
	"github.com/loyyal/loyyal-be-contract/utils/common"
	"go.opentelemetry.io/otel"
)

type WalletController struct {
	WalletService      services.WalletService
	TransactionService services.TransactionService
	Nats               *nats.Client
}

// constructor calling
func NewWallet(service services.WalletService, transactionService services.TransactionService, nats *nats.Client) WalletController {
	return WalletController{
		WalletService:      service,
		TransactionService: transactionService,
		Nats:               nats,
	}
}

type WalletCreateRequest struct {
	WalletType   string          `json:"walletType" binding:"required"`
	Name         string          `json:"name" binding:"required"`
	Metadata     json.RawMessage `json:"metadata"`
	PreLoadValue int64           `json:"preLoadValue"`
	LinkedTo     string          `json:"linkedTo" binding:"required"`
}

type WalletMergeRequest struct {
	From []string `json:"from" binding:"required"`
	To   string   `json:"to" binding:"required"`
}

func (controller *WalletController) walletCreate(ctx *gin.Context) {
	fName := "controller/wallet/create"
	tracer := otel.Tracer("walleCreate")
	_, span := tracer.Start(ctx.Request.Context(), fName)
	defer span.End()

	var request WalletCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: invlaid request body provided", fmt.Sprintf("got :%d ", err))
		return
	}

	var wallet models.Wallet
	wallet.Name = request.Name
	wallet.Metadata = request.Metadata
	wallet.WalletType = request.WalletType

	err := controller.WalletService.Create(ctx.Request.Context(), &wallet, request.LinkedTo, request.PreLoadValue)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error while creating wallet", fmt.Sprintf("got :%d ", err))
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

func (controller *WalletController) walletMerge(ctx *gin.Context) {
	fName := "controllers/wallet/merge"
	tracer := otel.Tracer("walleMerge")
	_, span := tracer.Start(ctx.Request.Context(), fName)
	defer span.End()

	var request WalletMergeRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: invalid request body provided", fmt.Sprintf("got :%d ", err))
		return
	}

	if len(request.From) < 1 {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: atlest one wallet is reqired", fmt.Sprintf("got :%s ", request.From))
		return
	}

	toWallet, err := controller.WalletService.Get(ctx.Request.Context(), request.To)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%d ", err))
		return
	}

	if common.IsStructEmpty(toWallet) {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: no wallet found with "+request.To, fmt.Sprintf("got :%s ", request.To))
		return
	}

	// recursively, get all FROM wallet and transfer their balance to TO wallet
	for _, walletIdentifier := range request.From {
		wallet, err := controller.WalletService.Get(ctx.Request.Context(), walletIdentifier)
		if err != nil {
			common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%d ", err))
			return
		}

		if common.IsStructEmpty(wallet) {
			common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: no wallet found with "+walletIdentifier, fmt.Sprintf("got :%s ", walletIdentifier))
			return
		}

		var transaction models.Transaction
		transaction.FromExtID = wallet.Identifier
		transaction.ToExtID = toWallet.Identifier
		transaction.FromUUID = wallet.UUID
		transaction.ToUUID = toWallet.UUID
		transaction.Amount = wallet.Balance
		transaction.TransactionType = "merge"
		transaction.Remarks = "merged into the wallet " + toWallet.Identifier

		err = controller.TransactionService.Create(ctx.Request.Context(), &transaction)
		if err != nil {
			common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%v ", err))
			return
		}

		// publishing to nats
		// do not need apply the business contract
		go controller.publishTxToNats(ctx.Request.Context(), &transaction)

		// markging wallet as disabled
		controller.WalletService.Update(ctx.Request.Context(), wallet.Identifier, "admin", "disabled")

	}
	common.PrepareCustomResponse(ctx, "wallet merged", nil)
}

func (controller *WalletController) walletGet(ctx *gin.Context) {
	fName := "controller/wallet/get"
	tracer := otel.Tracer("walletGet")
	_, span := tracer.Start(ctx.Request.Context(), fName)
	defer span.End()

	walletId := ctx.Query("walletId")
	if walletId == "" {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: wallet id is required", fmt.Sprintf("got :%s ", walletId))
		return
	}

	wallet, err := controller.WalletService.Get(ctx.Request.Context(), walletId)
	if wallet.UUID != "" {
		wallet.IsCommitedOnBlockchain = true
	}
	wallet.UUID = ""
	wallet.Ref = ""
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "wallet fetched", wallet)
}

func (controller *WalletController) walletDelete(ctx *gin.Context) {
	fName := "controller/wallet/delete"
	tracer := otel.Tracer("walletDelete")
	_, span := tracer.Start(ctx.Request.Context(), fName)
	defer span.End()

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

	err := controller.WalletService.Delete(ctx.Request.Context(), wallet.Identifier, "admin")
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "wallet deleted", nil)
}

func (controller *WalletController) walletFilter(ctx *gin.Context) {
	fName := "controller/wallet/filter"
	tracer := otel.Tracer("walletFilter")
	_, span := tracer.Start(ctx.Request.Context(), fName)
	defer span.End()

	queryString := "AND isDeleted=false"

	var wallet models.Wallet
	if err := ctx.BindJSON(&wallet); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: invalid request payload provided", fmt.Sprintf("got :%s ", err))
		return
	}

	if wallet.Identifier != "" {
		queryString += " AND identifier=$identifier"
	}

	if wallet.Name != "" {
		queryString += " AND name=$name"
	}

	wallets, err := controller.WalletService.Filter(ctx.Request.Context(), queryString, map[string]interface{}{
		"identifier": wallet.Identifier,
		"name":       wallet.Name,
	}, "createdAt", -1)

	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "wallets fetched", wallets)
}

func (controller *WalletController) publishTxToNats(ctx context.Context, transaction *models.Transaction) {

	var request nats.TopicEncoder
	request = &models.TransferRequest{
		RefID:   transaction.RefID,
		From:    transaction.FromUUID,
		To:      transaction.ToUUID,
		Channel: transaction.Channel,
		Amount:  transaction.Amount,
		Update:  !transaction.Spend,
	}

	if err := controller.Nats.Publish(ctx, request); err != nil {
		fmt.Print("failed to write merge transaction to NATS (failing over to retry service): %w", err)
		// TODO: execute the retry flow from here
	}
}

func (controller *WalletController) WalletRoutes(group *gin.RouterGroup) {
	walletRoute := group.Group("/wallet")

	walletRoute.Use(middleware.JWTAuthMiddleware())

	walletRoute.GET("/get", controller.walletGet)
	walletRoute.POST("/filter", controller.walletFilter)
	walletRoute.POST("/create", controller.walletCreate)
	walletRoute.POST("/merge", controller.walletMerge)
	walletRoute.DELETE("/delete", controller.walletDelete)
}
