package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/middleware"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/nats"
	"github.com/loyyal/loyyal-be-contract/services"
	"github.com/loyyal/loyyal-be-contract/utils/common"
)

type TransactionController struct {
	logger             *log.Logger
	TransactionService services.TransactionService
	WalletService      services.WalletService
	ContractService    services.ContractService
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
	ERR_SAME_FROM_AND_TO                = errors.New("error: from and to can not be same address")
	ERR_INSUFICIENT_BALANCE             = errors.New("error: insuffient balance")
	ERR_INVALID_TRANSACTION_ID_PROVIDED = errors.New("error: invalid transaction id provided")
)

// constructor calling
func NewTransactionController(logger *log.Logger, transactionService services.TransactionService, contractService services.ContractService, walletService services.WalletService, nats *nats.Client) TransactionController {
	return TransactionController{
		logger:             logger,
		TransactionService: transactionService,
		ContractService:    contractService,
		WalletService:      walletService,
		Nats:               nats,
	}
}

func (controller *TransactionController) issue(ctx *gin.Context) {
	fName := "transactioncontrller/issue"
	var input TransactionInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: invalid request body provided", fmt.Sprintf("got :%s ", err))
		return
	}

	if input.Amount <= 0 {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_AMOUNT_NEGATIVE_OR_ZERO.Error(), fmt.Sprintf("got :%d ", input.Amount))
		return
	}

	if input.From == input.To {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_AMOUNT_NEGATIVE_OR_ZERO.Error(), fmt.Sprintf("got: from %s & to %s", input.From, input.To))
		return
	}

	// check if From is valid
	walletFrom, err := controller.WalletService.Get(input.From)
	if common.IsStructEmpty(walletFrom) {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_INVALID_WALLET.Error(), fmt.Sprintf("got :%v ", walletFrom))
		return
	}
	// check if From has available balance
	if walletFrom.Balance <= 0 || walletFrom.Balance <= input.Amount {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_INSUFICIENT_BALANCE.Error(), fmt.Sprintf("got :%v ", walletFrom))
		return
	}

	// check if To is valid
	walletTo, err := controller.WalletService.Get(input.To)
	if common.IsStructEmpty(walletTo) {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_INVALID_WALLET.Error(), fmt.Sprintf("got :%v ", walletTo))
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
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%v ", err))
		return
	}

	// publishing to nats
	go controller.ApplyBusinessContractAndPublishToNats(ctx, &transaction)

	common.PrepareCustomResponse(ctx, "points issued", struct {
		Identifier string `json:"identifier"`
	}{Identifier: transaction.ExtID})
}

func (controller *TransactionController) redeem(ctx *gin.Context) {
	fName := "transactioncontrller/redeem"
	var input TransactionInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: invalid request body provided", fmt.Sprintf("got :%s ", err))
		return
	}

	if input.Amount <= 0 {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_AMOUNT_NEGATIVE_OR_ZERO.Error(), fmt.Sprintf("got :%d ", input.Amount))
		return
	}

	if input.From == input.To {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_AMOUNT_NEGATIVE_OR_ZERO.Error(), fmt.Sprintf("got: from %s & to %s", input.From, input.To))
		return
	}

	// check if From is valid
	walletFrom, err := controller.WalletService.Get(input.From)
	if common.IsStructEmpty(walletFrom) {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_INVALID_WALLET.Error(), fmt.Sprintf("got :%v ", walletFrom))
		return
	}
	// check if From has available balance
	if walletFrom.Balance <= 0 || walletFrom.Balance <= input.Amount {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_INSUFICIENT_BALANCE.Error(), fmt.Sprintf("got :%v ", walletFrom))
		return
	}

	// check if To is valid
	walletTo, err := controller.WalletService.Get(input.To)
	if common.IsStructEmpty(walletTo) {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_INVALID_WALLET.Error(), fmt.Sprintf("got :%v ", walletTo))
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
	transaction.TransactionType = "redeem"

	err = controller.TransactionService.Create(&transaction)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%v ", err))
		return
	}

	// publishing to nats
	go controller.publishTransactionToNats(ctx.Request.Context(), &transaction)
	common.PrepareCustomResponse(ctx, "points redeemed", struct {
		Identifier string `json:"identifier"`
	}{Identifier: transaction.ExtID})
}

func (controller *TransactionController) transfer(ctx *gin.Context) {
	fName := "transactioncontrller/transfer"
	var input TransactionInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: invalid request body provided", fmt.Sprintf("got :%s ", err))
		return
	}

	if input.Amount <= 0 {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_AMOUNT_NEGATIVE_OR_ZERO.Error(), fmt.Sprintf("got :%d ", input.Amount))
		return
	}

	if input.From == input.To {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_AMOUNT_NEGATIVE_OR_ZERO.Error(), fmt.Sprintf("got: from %s & to %s", input.From, input.To))
		return
	}

	// check if From is valid
	walletFrom, err := controller.WalletService.Get(input.From)
	if common.IsStructEmpty(walletFrom) {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_INVALID_WALLET.Error(), fmt.Sprintf("got :%v ", walletFrom))
		return
	}
	// check if From has available balance
	if walletFrom.Balance <= 0 || walletFrom.Balance < input.Amount {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_INSUFICIENT_BALANCE.Error(), fmt.Sprintf("got :%v ", walletFrom))
		return
	}

	// check if To is valid
	walletTo, err := controller.WalletService.Get(input.To)
	if common.IsStructEmpty(walletTo) {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, ERR_INVALID_WALLET.Error(), fmt.Sprintf("got :%v ", walletTo))
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
	transaction.TransactionType = "transfer"

	err = controller.TransactionService.Create(&transaction)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%v ", err))
		return
	}

	// publishing to nats
	go controller.ApplyBusinessContractAndPublishToNats(ctx, &transaction)
	common.PrepareCustomResponse(ctx, "points transferred", struct {
		Identifier string `json:"identifier"`
	}{Identifier: transaction.ExtID})
}

func (controller *TransactionController) deposit(ctx *gin.Context) {

}

func (controller *TransactionController) withdraw(ctx *gin.Context) {

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

	results, err := controller.TransactionService.Filter("", map[string]interface{}{}, "createdAt", -1)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type Transaction struct {
		Identifier      string    `json:"identifier"`
		Amount          int64     `json:"amount"`
		From            string    `json:"from"`
		To              string    `json:"to"`
		Currency        string    `json:"currency"`
		TransactionType string    `json:"transactionType"`
		CreatedOn       time.Time `json:"createdOn"`
		Creator         string    `json:"creator"`
	}

	transactions := []Transaction{}

	for _, row := range results {
		var tx Transaction
		tx.Identifier = row.ExtID
		tx.Amount = row.Amount
		tx.From = row.FromExtID
		tx.To = row.ToExtID
		tx.TransactionType = row.TransactionType
		tx.Creator = row.Creator
		tx.CreatedOn = row.CreatedOn

		transactions = append(transactions, tx)

	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    transactions,
	})
}

func (controller *TransactionController) ApplyBusinessContractAndPublishToNats(ctx *gin.Context, transaction *models.Transaction) {
	contracts, err := controller.ContractService.Filter("AND isDeleted=false", map[string]interface{}{}, "createdAt", -1)
	if err != nil {
		controller.logger.Println("failed to query the contract for dynamic application: %w", err)
	}

	location, _ := time.LoadLocation("UTC")
	now, _ := time.Parse(time.RFC1123, time.Now().In(location).Format(time.RFC1123))
	if len(contracts) > 0 {
		// sorting based on priority
		sort.Slice(contracts, func(i, j int) bool {
			return contracts[i].Priorty > contracts[j].Priorty
		})

		filteredContract := []models.Contract{}
		for _, row := range contracts {
			if row.ValidFrom.Before(now) && row.ValidUntill.After(now) {
				filteredContract = append(filteredContract, *row)
			}
		}

		if len(filteredContract) > 0 {
			sort.Slice(filteredContract, func(i, j int) bool {
				return filteredContract[i].Priorty > filteredContract[j].Priorty
			})

			maxPriority := filteredContract[0].Priorty
			maxPriorityContract := []models.Contract{}
			for _, row := range filteredContract {
				if row.Priorty == maxPriority {
					maxPriorityContract = append(maxPriorityContract, row)
				}
			}

			var applicableContract models.Contract
			if len(maxPriorityContract) == 1 {
				applicableContract = filteredContract[0]
			} else {
				sort.Slice(maxPriorityContract, func(i, j int) bool {
					return filteredContract[i].LastUpdatedAt.Unix() > filteredContract[j].LastUpdatedAt.Unix()
				})

				applicableContract = maxPriorityContract[0]
			}

			// apply contract
			transaction.Amount = transaction.Amount * applicableContract.BurnConversionRatio
			transaction.AppliedContract = applicableContract.Identifier
		}

	}
	controller.publishTransactionToNats(ctx.Request.Context(), transaction)

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
		controller.logger.Println("failed to write wallet to NATS (failing over to retry service): %w", err)
	}
}

func (controller *TransactionController) TransactionRoutes(group *gin.RouterGroup) {
	transactionRoute := group.Group("/transaction")

	transactionRoute.Use(middleware.JWTAuthMiddleware())

	transactionRoute.POST("/filter", controller.TransactionFilter)
	transactionRoute.GET("/get", controller.TransactionGet)
	transactionRoute.POST("/earn", controller.issue)
	transactionRoute.POST("/redeem", controller.redeem)
	transactionRoute.POST("/transfer", controller.transfer)
}
