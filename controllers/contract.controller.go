package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/middleware"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/services"
	"github.com/loyyal/loyyal-be-contract/utils/common"
	"github.com/loyyal/loyyal-be-contract/utils/notification"
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
	fName := "controllers/ContractCreate"
	var contract models.Contract
	if err := ctx.ShouldBindJSON(&contract); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: invalid request body provided", fmt.Sprintf("got %s ", err.Error()))
		return
	}

	if contract.ContractId <= 0 {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: contract id is required", fmt.Sprintf("contract id is required. got %d ", contract.ContractId))
		return
	}

	if contract.ValidFrom.Unix() < time.Now().Unix() || contract.ValidUntill.Unix() < time.Now().Unix() {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: previous date contract can not be created", fmt.Sprintf("previous date contract can not be created. got from: %d and untill :%d ", contract.ValidFrom, contract.ValidUntill))
		return
	}

	identifier, err := controller.ContractService.CreateContract(&contract, "creator", "loyyalchannel")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got : %s ", err.Error()))
		return
	}

	common.PrepareCustomResponse(ctx, "contract created successfully", struct {
		Identifier string `json:"identifier"`
	}{Identifier: identifier})
}

func (controller *ContractController) ContractGet(ctx *gin.Context) {
	fName := "controllers/ConractGet"
	contractId := ctx.Query("contractId")
	if contractId == "" {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: contract id is required", fmt.Sprintf("got :%s ", contractId))
		return
	}

	contract, err := controller.ContractService.GetContract(contractId)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    contract,
	})
	common.PrepareCustomResponse(ctx, "contract fetched", nil)
}

func (controller *ContractController) ContractDelete(ctx *gin.Context) {
	fName := "controllers/ContractDelete"
	contractId := ctx.Param("contractId")
	if contractId == "" {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: contractId is required", fmt.Sprintf("got :%s ", contractId))
		return
	}

	err := controller.ContractService.DeleteContract(contractId, "admin")
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "contract deleted", nil)
}

func (controller *ContractController) ContractFilter(ctx *gin.Context) {
	fName := "controller/ContractFilter"
	contracts, err := controller.ContractService.Filter("and contractType = $contractType", map[string]interface{}{
		"contractType": "Regular",
	}, "createdAt", 10)

	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "contract deleted", contracts)
}

func (controller *ContractController) SendEmail(ctx *gin.Context) {
	go notification.SendEmailNotification("rohit@loyyal.com", "Testing Sendgrid Email")

	common.PrepareCustomResponse(ctx, "email sent", nil)
}

func (controller *ContractController) ContractRoutes(group *gin.RouterGroup) {
	contractRoute := group.Group("/contract")
	contractRoute.Use(middleware.JWTAuthMiddleware())

	contractRoute.GET("/get", controller.ContractGet)
	contractRoute.POST("/filter", controller.ContractFilter)
	contractRoute.POST("/create", controller.ContractCreate)
	contractRoute.DELETE("/delete", controller.ContractDelete)
	contractRoute.GET("/send-email", controller.SendEmail)
}
