package models

import (
	"encoding/json"
	"time"
)

type Asset struct {
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

type Wallet struct {
	DocType    string `json:"type"`
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	// Metadata stores the user provided data, for example, a description.
	Metadata json.RawMessage `json:"metadata,omitempty"`
	// TODO: add the default token udner balance to open end the possiblities
	// handling multiple wallets
	Assets     []Asset  `json:"assets,omitempty"`
	WalletType string   `json:"walletType"`
	Status     string   `json:"status"`
	LinkedTo   []string `json:"linkedTo"`

	// Creator records the user that created the record.
	Creator string `json:"-"`
	// Channel records the channel on which the wallet will be written.
	Channel string `json:"-"`
	// Error is the error message returned from chaincode on failing to
	// create a wallet.
	Error string `json:"errmsg,omitempty"`

	// Balance is the hard balance of the wallet as returned by the chain.
	Balance int64 `json:"balance"`

	// SoftBalance (computed on the fly) is the hard balance plus any
	// pending diff.
	// TODO: omitting for now: calculate it for all wallet responses
	SoftBalance int64 `json:"-"`

	// 	// UUID is the chain generated unique ID of the wallet.
	UUID string `json:"uuid"`

	// // Ref is a temporary deduplication reference generated and stored in
	// // couchbase and used for retries. This field is set automatically and
	// // should be read only for internal applications. It is not for
	// // external use and should not be exposed to API users.
	Ref string `json:"refid"`

	CreatedAt     time.Time `json:"createdAt"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`
	LastUpdatedBy string    `json:"lastUpdatedBy"`
	IsDeleted     bool      `json:"isDeleted"`
}
