package models

type Contract struct {
	ContractId   string `json:"contractId"`
	ContractName string `json:"contractName"`

	OperatorId   int64  `json:"operatorId"`
	OperatorName string `json:"operatorName"`

	IsDeleted bool `json:"isDeleted"`
}
