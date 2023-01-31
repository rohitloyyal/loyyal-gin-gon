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
