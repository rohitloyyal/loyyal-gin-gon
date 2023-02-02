package models

import "encoding/json"

const (
	TopicCreate   = "create"
	TopicIssue    = "issue"
	TopicTransfer = "transfer"
	TopicBalance  = "balance"
	TopicUUID     = "uuid"

	callbackJoiner = "callingback"
)

type CreateRequest struct {
	Creator     string `json:"creator"`
	RefID       string `json:"refid"`
	Channel     string `json:"channel"`            // channel to write to
	Amount      int64  `json:"amount"`             // optional amount to start account with
	SubmittedAt int64  `json:"submitted"`          // when request was submitted
	Backfill    bool   `json:"backfill,omitempty"` // use submitted time as transaction time
}

func (c CreateRequest) Encode() ([]byte, error) {
	return json.Marshal(c)
}

func (c *CreateRequest) Decode(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c CreateRequest) TopicName() string {
	return TopicCreate + "." + c.Channel
}

// IssueRequest is a request to issue points to an account
type IssueRequest struct {
	//TraceID TraceID `json:"trace_id,omitempty"`
	ID          string `json:"id,omitempty"`
	RefID       string `json:"reference_id"` // external reference id (must be unique for each transaction)
	CostBasis   string `json:"cost_basis"`   // cost basis tag of the issue
	Channel     string `json:"channel"`      // channel to write to
	Amount      int64  `json:"amount"`
	SubmittedAt int64  `json:"submitted"`          // when request was submitted
	Backfill    bool   `json:"backfill,omitempty"` // use submitted time as transaction time
}

// Encode converts the request into GOB bytes
func (r IssueRequest) Encode() ([]byte, error) {
	return json.Marshal(r)
}

// Decode converts GOB bytes back to the request
func (r *IssueRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// TopicName returns the topic associated with the request
func (r IssueRequest) TopicName() string {
	return TopicIssue + "." + r.Channel
}



type TransferRequest struct {
	Creator     string `json:"creator,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	RefID       string `json:"reference_id"`         // external reference id (must be unique for each transaction)
	Channel     string `json:"channel"`              // channel to write to
	CostBasis   string `json:"cost_basis,omitempty"` // tagging of transaction
	Amount      int64  `json:"amount"`
	SubmittedAt int64  `json:"submitted"`            // timestamp request was submitted
	Backfill    bool   `json:"backfill,omitempty"`   // use submitted time as transaction time
	SkipCheck   bool   `json:"skip_check,omitempty"` // skip checking if balance is adequate before submitting
	Update      bool   `json:"update,omitempty"`     // update "To's" balance if set true (default: false)
}

// Encodes marshals the request to binary
func (r TransferRequest) Encode() ([]byte, error) {
	return json.Marshal(r)
}

// Decode unmarshals binary data
func (r *TransferRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

func (r TransferRequest) TopicName() string {
	return TopicTransfer + "." + r.Channel
}
