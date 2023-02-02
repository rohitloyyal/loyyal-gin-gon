package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/middleware"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/services"
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
	// get data from body
	var identity models.Identity
	if err := ctx.ShouldBindJSON(&identity); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if identity.Username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "username is required",
		})
		return
	}

	if identity.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "password is required",
		})
		return
	}

	identifier, err := controller.IdentityService.Create(&identity)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    identifier,
	})
}

func (controller *IdentityController) IdentityGet(ctx *gin.Context) {
	// get data from body
	identityId := ctx.Query("identityId")
	if identityId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "identity id is required",
		})
		return
	}

	identity, err := controller.IdentityService.Get(identityId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    identity,
	})
}

func (controller *IdentityController) IdentityUpdate(ctx *gin.Context) {
	// get data from body
	var identity models.Identity
	if err := ctx.ShouldBindJSON(&identity); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if identity.Identifier == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "identifier is required",
		})
		return
	}

	err := controller.IdentityService.Update(identity.Identifier, identity.PersonalDetails)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    identity,
	})
}

func (controller *IdentityController) identityDelete(ctx *gin.Context) {
	var identity models.Identity
	if err := ctx.ShouldBindJSON(&identity); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if identity.Identifier == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "identifier is required",
		})
		return
	}

	err := controller.IdentityService.Delete(identity.Identifier, "admin")
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

func (controller *IdentityController) identityFilter(ctx *gin.Context) {

	identities, err := controller.IdentityService.Filter("and walletType = $walletType", map[string]interface{}{
		"walletType": "regular_wallet",
	}, "createdAt", 10)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"body":    identities,
	})
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
