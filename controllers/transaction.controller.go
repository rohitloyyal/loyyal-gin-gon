package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/middleware"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/nats"
	"github.com/loyyal/loyyal-be-contract/services"
	"github.com/loyyal/loyyal-be-contract/utils/common"
)

type TransactionController struct {
	TransactionService services.TransactionService
	WalletService      services.WalletService
	Nats               *nats.Client
}

type TransactionInput struct {
	From                   string          `json:"from" binding:"required"`
	To                     string          `json:"to" binding:"required"`
	Amount                 int64           `json:"amount" binding:"required"`
	Metadata               json.RawMessage `json:"metadata"`
	TransactionInitiatedBy string          `json:"transactionInitiatedBy"`
}

var (
	ERR_INVALID_WALLET                  = errors.New("error: invalid wallet provided")
	ERR_AMOUNT_NEGATIVE_OR_ZERO         = errors.New("error: amount can not be empty or zero")
	ERR_INSUFICIENT_BALANCE             = errors.New("error: insuffient balance")
	ERR_INVALID_TRANSACTION_ID_PROVIDED = errors.New("error: invalid transaction id provided")
)

// constructor calling
func NewTransactionController(transactionService services.TransactionService, walletService services.WalletService, nats *nats.Client) TransactionController {
	return TransactionController{
		TransactionService: transactionService,
		WalletService:      walletService,
		Nats:               nats,
	}
}

func (controller *TransactionController) issue(ctx *gin.Context) {
	var input TransactionInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if input.Amount <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": ERR_AMOUNT_NEGATIVE_OR_ZERO,
		})
		return
	}

	// check if From is valid
	walletFrom, err := controller.WalletService.Get(input.From)
	if common.IsStructEmpty(walletFrom) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": ERR_INVALID_WALLET,
		})
		return
	}
	// check if From has available balance
	if walletFrom.Balance <= 0 || walletFrom.Balance <= input.Amount {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": ERR_INSUFICIENT_BALANCE,
		})
		return
	}

	// check if To is valid
	walletTo, err := controller.WalletService.Get(input.To)
	if common.IsStructEmpty(walletTo) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": ERR_INVALID_WALLET,
		})
		return
	}

	var transaction models.Transaction
	transaction.FromExtID = input.From
	transaction.ToExtID = input.To
	transaction.TransactionInitiatedBy = input.TransactionInitiatedBy
	transaction.FromUUID = walletFrom.UUID
	transaction.ToUUID = walletTo.UUID

	transaction.Amount = input.Amount
	transaction.Metadata = input.Metadata
	transaction.TransactionType = "issue"

	err = controller.TransactionService.Create(&transaction)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// publishing to nats
	go controller.publishTransactionToNats(ctx.Request.Context(), &transaction)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    transaction.ExtID,
	})
}

func (controller *TransactionController) redeem(ctx *gin.Context) {
	var input TransactionInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var transaction models.Transaction
	transaction.FromExtID = input.From
	transaction.ToExtID = input.To
	transaction.Amount = input.Amount
	transaction.Metadata = input.Metadata
	transaction.TransactionType = "redeem"
	transaction.CreatedOn = time.Now()

	err := controller.TransactionService.Create(&transaction)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// publishing to nats
	controller.publishTransactionToNats(ctx.Request.Context(), &transaction)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    transaction.RefID,
	})
}

func (controller *TransactionController) transfer(ctx *gin.Context) {
	var input TransactionInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var transaction models.Transaction
	transaction.FromExtID = input.From
	transaction.ToExtID = input.To
	transaction.Amount = input.Amount
	transaction.Metadata = input.Metadata
	transaction.TransactionType = "transfer"
	transaction.CreatedOn = time.Now()

	err := controller.TransactionService.Create(&transaction)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// publishing to nats
	controller.publishTransactionToNats(ctx.Request.Context(), &transaction)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    transaction.RefID,
	})
}

func (controller *TransactionController) TransactionGet(ctx *gin.Context) {
	transactionId := ctx.Query("transactionId")
	if transactionId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "transaction id is required",
		})
		return
	}

	transaction, err := controller.TransactionService.Get(transactionId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    transaction,
	})
}

func (controller *TransactionController) TransactionFilter(ctx *gin.Context) {

	transaction, err := controller.TransactionService.Filter("and amount = $amount", map[string]interface{}{
		"amount": 4,
	}, "createdAt", 10)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    transaction,
	})
}

func (controller *TransactionController) publishTransactionToNats(ctx context.Context, transaction *models.Transaction) {
	var request nats.TopicEncoder
	request = &models.TransferRequest{
		RefID:   transaction.RefID,
		From:    transaction.FromUUID,
		To:      transaction.ToUUID,
		Channel: transaction.Channel,
		Amount:  transaction.Amount,
		Update:  !transaction.Spend,
	}

	// in case of issue request
	if transaction.FromExtID == "" {
		request = &models.IssueRequest{
			ID:      transaction.ToUUID,
			RefID:   transaction.RefID,
			Amount:  transaction.Amount,
			Channel: transaction.Channel,
		}
	}
	if err := controller.Nats.Publish(ctx, request); err != nil {
		fmt.Print("failed to write wallet to NATS (failing over to retry service): %w", err)
	}
}

func (controller *TransactionController) TransactionRoutes(group *gin.RouterGroup) {
	transactionRoute := group.Group("/transaction")
	transactionRoute.Use(middleware.JWTAuthMiddleware())

	transactionRoute.POST("/filter", controller.TransactionFilter)
	transactionRoute.GET("/get", controller.TransactionGet)
	transactionRoute.POST("/issue", controller.issue)
	transactionRoute.POST("/redeem", controller.issue)
	transactionRoute.POST("/transfer", controller.issue)
}
