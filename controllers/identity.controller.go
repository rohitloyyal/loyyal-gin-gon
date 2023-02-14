package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.opentelemetry.io/otel"

	"github.com/loyyal/loyyal-be-contract/middleware"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/services"
	"github.com/loyyal/loyyal-be-contract/utils/common"
)

type IdentityController struct {
	logger          *log.Logger
	IdentityService services.IdentityService
	WalletService   services.WalletService
}

// constructor calling
func NewIdentityController(logger *log.Logger, service services.IdentityService, walletservice services.WalletService) IdentityController {
	return IdentityController{
		IdentityService: service,
		WalletService:   walletservice,
		logger:          logger,
	}
}

type DefaultWalletCreate struct {
	IsdefaultWalletRequired bool   `json:"isdefaultWalletRequired"`
	PreLoadValue            int64  `json:"preLoadValue"`
	WalletName              string `json:"walletName"`
}

func (controller *IdentityController) identityCreate(ctx *gin.Context) {
	fName := "identitycontroller/create"
	tracer := otel.Tracer("identityCreate")
	_, span := tracer.Start(ctx.Request.Context(), fName)

	defer span.End()
	var identity models.Identity
	if err := ctx.ShouldBindBodyWith(&identity, binding.JSON); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	if identity.Username == "" {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: username is required", fmt.Sprintf("got :%s ", identity.Username))
		return
	}

	if identity.Password == "" {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: password is required", fmt.Sprintf("got :%s ", identity.Password))
		return
	}

	if identity.IdentityType == "" {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: identity type is required", fmt.Sprintf("got :%s ", identity.IdentityType))
		return
	}

	// checking if default wallet creation is required
	var defaultWallet DefaultWalletCreate
	if err := ctx.ShouldBindBodyWith(&defaultWallet, binding.JSON); err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	if defaultWallet.IsdefaultWalletRequired {
		if defaultWallet.WalletName == "" {
			common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: wallet name is required", fmt.Sprintf("got :%s ", defaultWallet.WalletName))
			return
		}
	}

	identifier, err := controller.IdentityService.Create(ctx.Request.Context(), &identity)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	// creating default wallet
	var wallet models.Wallet
	wallet.Name = defaultWallet.WalletName

	err = controller.WalletService.Create(ctx.Request.Context(), &wallet, identity.Identifier, defaultWallet.PreLoadValue)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "identity created successfully", struct {
		Identifier string `json:"identifier"`
	}{Identifier: identifier})
}

func (controller *IdentityController) IdentityGet(ctx *gin.Context) {
	fName := "identitycontroller/get"
	tracer := otel.Tracer("identityGet")
	_, span := tracer.Start(ctx.Request.Context(), fName)
	defer span.End()

	identityId := ctx.Query("identityId")
	if identityId == "" {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: identifier is required", fmt.Sprintf("got :%s ", identityId))
		return
	}

	identity, err := controller.IdentityService.Get(ctx.Request.Context(), identityId)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "identity fetched", identity)
}

func (controller *IdentityController) IdentityUpdate(ctx *gin.Context) {
	fName := "identitycontroller/update"
	tracer := otel.Tracer("identityUpdate")
	_, span := tracer.Start(ctx.Request.Context(), fName)

	defer span.End()
	var identity models.Identity
	if err := ctx.ShouldBindJSON(&identity); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if identity.Identifier == "" {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: identifier is required", fmt.Sprintf("got :%s ", identity.Identifier))
		return
	}

	err := controller.IdentityService.Update(ctx.Request.Context(), identity.Identifier, identity.PersonalDetails)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    identity,
	})
	common.PrepareCustomResponse(ctx, "identity updated", nil)
}

func (controller *IdentityController) identityDelete(ctx *gin.Context) {
	fName := "identitycontroller/delete"
	tracer := otel.Tracer("identityDelete")
	_, span := tracer.Start(ctx.Request.Context(), fName)
	defer span.End()

	var identity models.Identity
	if err := ctx.ShouldBindJSON(&identity); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if identity.Identifier == "" {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: identifier is required", fmt.Sprintf("got :%s ", identity.Identifier))
		return
	}

	err := controller.IdentityService.Delete(ctx.Request.Context(), identity.Identifier, "admin")
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "identity deleted", nil)
}

func (controller *IdentityController) identityLinkedWallets(ctx *gin.Context) {
	fName := "identitycontroller/linkedwallets"
	tracer := otel.Tracer("identityLinkedWallets")
	_, span := tracer.Start(ctx.Request.Context(), fName)
	defer span.End()

	userId := ctx.Query("userId")
	wallets, err := controller.WalletService.CustomFilterQuery(ctx.Request.Context(), "identifier, name, balance, walletType, status, createdAt",
		"and isDeleted=false and any wallet in testbucket.linkedTo SATISFIES wallet == $userId end", map[string]interface{}{
			"userId": userId,
		}, "createdAt", -1)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "linked wallets fetched", wallets)
}

func (controller *IdentityController) identityFilter(ctx *gin.Context) {
	fName := "identitycontroller/filter"
	tracer := otel.Tracer("identityFilter")
	_, span := tracer.Start(ctx.Request.Context(), fName)
	defer span.End()
	identities, err := controller.IdentityService.Filter(ctx.Request.Context(), "AND isDeleted=false AND identityType!='admin'", map[string]interface{}{}, "createdAt", -1)

	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "identities filtered", identities)
}

func (controller *IdentityController) IdentityRoutes(group *gin.RouterGroup) {
	identityRoute := group.Group("/identity")

	identityRoute.Use(middleware.JWTAuthMiddleware())

	identityRoute.GET("/get", controller.IdentityGet)
	identityRoute.GET("/get-linked-wallets", controller.identityLinkedWallets)
	identityRoute.POST("/filter", controller.identityFilter)
	identityRoute.POST("/create", controller.identityCreate)
	identityRoute.PUT("/update", controller.IdentityUpdate)
	identityRoute.DELETE("/delete", controller.identityDelete)

}
