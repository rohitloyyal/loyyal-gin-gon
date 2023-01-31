package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/middleware"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/nats"
	"github.com/loyyal/loyyal-be-contract/services"
)

type TransactionController struct {
	TransactionService services.TransactionService
	Nats               *nats.Client
}

// constructor calling
func NewTransactionController(service services.TransactionService, nats *nats.Client) TransactionController {
	return TransactionController{
		TransactionService: service,
		Nats:               nats,
	}
}

func (controller *TransactionController) issue(ctx *gin.Context) {
	// get data from body
	var transaction models.Transaction
	if err := ctx.ShouldBindJSON(&transaction); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if transaction.From == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "from is required",
		})
		return
	}

	if transaction.To == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "to is required",
		})
		return
	}

	transaction.TransactionType = "issue"
	err := controller.TransactionService.Create(&transaction)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    transaction.Ref,
	})
}

func (controller *TransactionController) redeem(ctx *gin.Context) {
	// get data from body
	var transaction models.Transaction
	if err := ctx.ShouldBindJSON(&transaction); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if transaction.From == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "from is required",
		})
		return
	}

	if transaction.To == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "to is required",
		})
		return
	}

	transaction.TransactionType = "redeem"
	err := controller.TransactionService.Create(&transaction)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    transaction.Ref,
	})
}

func (controller *TransactionController) transfer(ctx *gin.Context) {
	// get data from body
	var transaction models.Transaction
	if err := ctx.ShouldBindJSON(&transaction); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if transaction.From == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "from is required",
		})
		return
	}

	if transaction.To == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "to is required",
		})
		return
	}

	transaction.TransactionType = "transfer"
	err := controller.TransactionService.Create(&transaction)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    transaction.Ref,
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

func (controller *TransactionController) TransactionRoutes(group *gin.RouterGroup) {
	transactionRoute := group.Group("/transaction")
	transactionRoute.Use(middleware.JWTAuthMiddleware())

	transactionRoute.GET("/get", controller.TransactionGet)
	transactionRoute.POST("/issue", controller.issue)
	transactionRoute.POST("/redeem", controller.issue)
	transactionRoute.POST("/transfer", controller.issue)
}
