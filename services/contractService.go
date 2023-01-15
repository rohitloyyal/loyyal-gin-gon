package services

type ContractService interface {
	FindContracts(page int, limit int)
}
