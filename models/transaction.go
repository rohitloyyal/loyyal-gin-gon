package models

import (
	"encoding/json"
	"time"
)

type Transaction struct {
	DocType string `json:"docType"`
	TransactionWritables
	TransactionReadOnlys

	// Creator is the API user creating the transaction. Set based on the
	// API token used.
	Creator string `json:"-"`

	// Channel is the fabric channel that will be written to.
	Channel string `json:"-"`
}

type TransactionExtID struct {
	// Ref is the user facing descriptive key
	Ref string `json:"ref"`
}

type TransactionWritables struct {
	TransactionExtID

	// Metadata stores the user provided data, for example, a description.
	Metadata json.RawMessage `json:"metadata,omitempty"`

	// From is the user facing name of the source wallet of the
	// transaction. If it is blank the transaction has no source wallet and
	// is considered an issuance.
	From string `json:"from,omitempty"`

	// To is the user facing name of the destination wallet of the
	// transaction. It cannot be blank.
	To string `json:"to"`

	// Currency is the user facing name of the currency being used in the
	// transaction.
	Currency string `json:"currency"`

	// Amount is the number of points being transfered or issued in the
	// transaction.
	Amount int64 `json:"amount"`

	// Spend defers the update of the balance of the To wallet in order to
	// unblock chain operations.
	Spend bool `json:"spend"`

	TransactionType string `json:"transactionType"`
	AppliedContract string `json:"appliedContract"`
}

type TransactionReadOnlys struct {
	// Kind is either issue or transfer depending on the value of the From
	// string. READONLY.
	Kind string `json:"kind,omitempty"`

	// Timestamp is the time that the transaction was appended to the
	// chain. Timestamps are in RFC3339 format. READONLY.
	Timestamp *time.Time `json:"timestamp,omitempty"`

	// Error is the user visible message for Rejected transactions.
	Error string `json:"error,omitempty"`
}

type TransactionSearch struct {
	Transaction
	TransactionSearchOpts
}

type TransactionSearchResponse struct {
	Transactions []*Transaction         `json:"transactions"`
	Stats        TransactionSearchStats `json:"stats"`
}

type TransactionSearchStats struct {
	Sum int64 `json:"sum"`
}

type TransactionSearchOpts struct {
	// Since is a search parameter that will select all transactions after
	// the given timestamp, including any with that exact timestamp.
	// Timestamps are in RFC3339 format.
	Since *time.Time `json:"since,omitempty"`

	// Until is a search parameter that will select all transactions before
	// the given timestamp, including any with that exact timestamp.
	// Timestamps are in RFC3339 format.
	Until *time.Time `json:"until,omitempty"`

	// Offset is a search parameter that can be used to paginate records.
	Offset int64 `json:"offset,omitempty"`

	// Limit is a search parameter that will limit the number of records
	// returned.
	Limit int64 `json:"limit,omitempty"`

	// AtLeast is a search parameter for Amount, return only records that
	// are at least this amount. If 0 then all amounts are valid.
	AtLeast int64 `json:"at_least,omitempty"`

	// AtMost is a search parameter for Amount, returning only records that
	// are at most this amount. If 0 then all amounts are valid.
	AtMost int64 `json:"at_most,omitempty"`

	// Status can be used to filter for pending / completed transactions.
	// If it is not set, both types of transactions will be returned.
	Status string

	// Wallet filters for matches on either ToExtID or FromExtID.
	Wallet string
}
