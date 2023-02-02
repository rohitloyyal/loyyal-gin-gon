package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/middleware"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/services"
	"github.com/loyyal/loyyal-be-contract/utils/common"
)

type IdentityController struct {
	IdentityService services.IdentityService
}

// constructor calling
func NewIdentityController(service services.IdentityService) IdentityController {
	return IdentityController{
		IdentityService: service,
	}
}

func (controller *IdentityController) identityCreate(ctx *gin.Context) {
	fName := "identitycontroller/create"
	var identity models.Identity
	if err := ctx.ShouldBindJSON(&identity); err != nil {
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

	identifier, err := controller.IdentityService.Create(&identity)
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
	identityId := ctx.Query("identityId")
	if identityId == "" {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, "error: identifier is required", fmt.Sprintf("got :%s ", identityId))
		return
	}

	identity, err := controller.IdentityService.Get(identityId)
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "identity fetched", identity)
}

func (controller *IdentityController) IdentityUpdate(ctx *gin.Context) {
	fName := "identitycontroller/update"
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

	err := controller.IdentityService.Update(identity.Identifier, identity.PersonalDetails)
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

	err := controller.IdentityService.Delete(identity.Identifier, "admin")
	if err != nil {
		common.PrepareCustomError(ctx, http.StatusBadRequest, fName, err.Error(), fmt.Sprintf("got :%s ", err))
		return
	}

	common.PrepareCustomResponse(ctx, "identity deleted", nil)
}

func (controller *IdentityController) identityFilter(ctx *gin.Context) {
	fName := "identitycontroller/filter"
	identities, err := controller.IdentityService.Filter("and walletType = $walletType", map[string]interface{}{
		"walletType": "regular_wallet",
	}, "createdAt", 10)

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
	identityRoute.POST("/filter", controller.identityFilter)
	identityRoute.POST("/create", controller.identityCreate)
	identityRoute.PUT("/update", controller.IdentityUpdate)
	identityRoute.DELETE("/delete", controller.identityDelete)

}
