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
	"go.opentelemetry.io/otel"
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
	tracer := otel.Tracer(fName)
	_, span := tracer.Start(ctx.Request.Context(), fName)
	defer span.End()

	var contract models.Contract
	if err := ctx.ShouldBindJSON(&contract); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: invalid request body provided", fmt.Sprintf("got %s ", err.Error()))
		return
	}

	location, _ := time.LoadLocation("UTC")
	currenTimestamp, _ := time.Parse(time.RFC1123, time.Now().In(location).Format(time.RFC1123))
	validFrom, _ := time.Parse(time.RFC1123, contract.ValidFrom.In(location).Format(time.RFC1123))
	validUpto, _ := time.Parse(time.RFC1123, contract.ValidUntill.In(location).Format(time.RFC1123))

	if validFrom.Unix() < currenTimestamp.Unix() || validUpto.Unix() < currenTimestamp.Unix() {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: previous date contract can not be created", fmt.Sprintf("previous date contract can not be created. got from: %d and untill :%d ", contract.ValidFrom, contract.ValidUntill))
		return
	}

	contract.ValidFrom = validFrom
	contract.ValidUntill = validUpto

	identifier, err := controller.ContractService.CreateContract(ctx.Request.Context(), &contract, "admin", "loyyalchannel")
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

	contract, err := controller.ContractService.GetContract(ctx.Request.Context(), contractId)
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
	tracer := otel.Tracer("ContractDelete")
	_, span := tracer.Start(ctx.Request.Context(), fName)
	defer span.End()

	var contract struct {
		Identifier string `json:"identifier" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&contract); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: invalid request body provided", fmt.Sprintf("got %s ", err.Error()))
		return
	}

	if contract.Identifier == "" {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: contrct identifier is required", fmt.Sprintf("got %s ", contract.Identifier))
	}

	err := controller.ContractService.DeleteContract(ctx.Request.Context(), contract.Identifier, "admin")
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "contract deleted", nil)
}

func (controller *ContractController) ContractFilter(ctx *gin.Context) {
	fName := "controller/ContractFilter"
	tracer := otel.Tracer("ContractFilter")
	_, span := tracer.Start(ctx.Request.Context(), fName)
	defer span.End()
	span.AddEvent("filtering contracts")

	// span.SetAttributes(attribute.KeyValue{Key: "32", Value: })

	// TODO: should make filter as dynamic based on the filters passed in request body
	contracts, err := controller.ContractService.Filter(ctx.Request.Context(), "AND isDeleted=false", map[string]interface{}{}, "createdAt", -1)

	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "contract filtered", contracts)
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
