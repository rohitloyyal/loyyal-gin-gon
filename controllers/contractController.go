package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loyyal/loyyal-be-contract/intializers"
	"github.com/loyyal/loyyal-be-contract/models"
)

func ContractCreate(c *gin.Context) {
	// get data from body

	var request struct {
		OperatorId   int64
		OperatorName string
		ContractId   string
		ContractName string
	}

	c.Bind(&request)
	if request.ContractId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "contract id is required",
		})
		return
	}

	// create a contract object
	contract := models.Contract{OperatorId: request.OperatorId, OperatorName: request.OperatorName, ContractId: request.ContractId, ContractName: request.ContractName}
	col := intializers.DB.DefaultCollection()

	_, err := col.Upsert(request.ContractId, contract, nil)
	if err != nil {
		log.Fatal(err)
	}

	//return response
	c.JSON(http.StatusOK, gin.H{
		"message": "contract created",
		"body":    contract,
	})

}

func ContractGet(c *gin.Context) {
	// get data from param
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "contract id is required",
		})
		return
	}
	// create a contract object
	col := intializers.DB.DefaultCollection()
	result, err := col.Get(id, nil)
	if err != nil {
		log.Fatal(err)
	}

	// var contract models.Contract

	//return response
	c.JSON(http.StatusOK, gin.H{
		"message": "contract fetched",
		"body":    result,
	})

}

func ContractUpdate(c *gin.Context) {
	// get data from param
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "contract id is required",
		})
		return
	}
	// create a contract object
	col := intializers.DB.DefaultCollection()
	result, err := col.Get(id, nil)
	if err != nil {
		log.Fatal(err)
	}

	if result == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no contract found",
		})
		return
	}

	var contract models.Contract
	result.Content(&contract)

	contract.IsDeleted = true
	_, err = col.Replace(contract.ContractId, contract, nil)
	if err != nil {
		log.Fatal(err)
	}

	//return response
	c.JSON(http.StatusOK, gin.H{
		"message": "contract updated",
		"body":    contract,
	})

}

func ContractDelete(c *gin.Context) {
	// get data from param
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "contract id is required",
		})
		return
	}
	// create a contract object
	// using transaction way
	// intializers.Cluster.Transactions().Run(func(ctx *gocb.TransactionAttemptContext) error {
	// 	doc, err := ctx.Get(*gocb.Bucket.DefaultCollection(), "")
	// })

	col := intializers.DB.DefaultCollection()
	result, err := col.Get(id, nil)
	if err != nil {
		log.Fatal(err)
	}

	if result == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no contract found",
		})
		return
	}

	var contract models.Contract
	result.Content(&contract)

	contract.IsDeleted = true
	_, err = col.Replace(contract.ContractId, contract, nil)
	if err != nil {
		log.Fatal(err)
	}

	//return response
	c.JSON(http.StatusOK, gin.H{
		"message": "contract deleted",
		"body":    contract,
	})

}
