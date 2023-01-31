package models

import "time"

type Contract struct {
	DocType  string `json:"type"`
	Identifier   string `json:"identifier"`
	ContractId   int64  `json:"contractId"`
	ContractName string `json:"contractName"`

	OperatorId   int64  `json:"operatorId"`
	OperatorName string `json:"operatorName"`

	PartnerId   int64  `json:"partnerId"`
	PartnerName string `json:"partnerName"`

	// [Regular, Promotional]
	ContractType string `json:"contractType"`
	// 1 - highest to 5 - lowest
	// 0 - unset
	Priorty            int64  `json:"priority"`
	ConversionCurrency string `json:"conversionCurrency"`
	// [MILES, POINTS]
	PointsType string `json:"pointsType"`

	ValidFrom   time.Time `json:"validFrom"`
	ValidUntill time.Time `json:"validUntill"`

	// [1:20, 2x]
	EarnConversionRatio int64 `json:"earnConversionRatio"`
	BurnConversionRatio int64 `json:"burnConversionRatio"`
	// Creator records the user that created the record.
	Creator string `json:"-"`
	// Channel records the channel on which the wallet will be written.
	Channel string `json:"-"`

	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`
	LastUpdatedBy string    `json:"lastUpdatedBy"`
}
