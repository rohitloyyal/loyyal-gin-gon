package models

import "time"

type Contract struct {
	DocType      string `json:"type"`
	Identifier   string `json:"identifier"`
	ContractId   int64  `json:"contractId"`
	ContractName string `json:"contractName" binding:"required"`

	OperatorId   int64  `json:"operatorId"`
	OperatorName string `json:"operatorName" binding:"required"`

	PartnerId   int64  `json:"partnerId"`
	PartnerName string `json:"partnerName" binding:"required"`

	// [Regular, Promotional]
	ContractType string `json:"contractType" binding:"required"`
	// 5 - highest to 0 - lowest
	Priorty            int64  `json:"priority" binding:"required"`
	ConversionCurrency string `json:"conversionCurrency" binding:"required"`
	// [MILES, POINTS]
	PointsType string `json:"pointsType" binding:"required"`

	ValidFrom   time.Time `json:"validFrom" binding:"required"`
	ValidUntill time.Time `json:"validUntill" binding:"required"`

	// [1:20, 2x]
	EarnConversionRatio int64 `json:"earnConversionRatio" binding:"required"`
	BurnConversionRatio int64 `json:"burnConversionRatio" binding:"required"`
	// Creator records the user that created the record.
	Creator string `json:"creator"`
	// Channel records the channel on which the wallet will be written.
	Channel string `json:"channel"`

	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`
	LastUpdatedBy string    `json:"lastUpdatedBy"`
	IsDeleted     bool      `json:"isDeleted"`
}
