package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/middleware"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/services"
)

type ContractController struct {
	ContractService services.ContractService
}

// constructor calling
func NewContractController(service services.ContractService) ContractController {
	return ContractController{
		ContractService: service,
	}
}

func (controller *ContractController) ContractCreate(ctx *gin.Context) {
	// get data from body
	var contract models.Contract
	if err := ctx.ShouldBindJSON(&contract); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if contract.ContractId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "contract id is required",
		})
		return
	}

	err := controller.ContractService.Create(&contract)
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

func (controller *ContractController) ContractGet(ctx *gin.Context) {
	// get data from body
	contractId := ctx.Param("contractId")
	// if contractId == "" {
	// 	ctx.JSON(http.StatusBadRequest, gin.H{
	// 		"message": "contract id is required",
	// 	})
	// 	return
	// }

	contract, err := controller.ContractService.Get(contractId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    contract,
	})
}

func (controller *ContractController) ContractRoutes(group *gin.RouterGroup) {
	contractRoute := group.Group("/contract")
	contractRoute.Use(middleware.JWTAuthMiddleware())

	contractRoute.GET("/get", controller.ContractGet)
	contractRoute.POST("/filter", controller.ContractCreate)
	contractRoute.POST("/create", controller.ContractCreate)
	contractRoute.PUT("/update", controller.ContractCreate)
	contractRoute.DELETE("/delete", controller.ContractCreate)
}
