package nats

import (
	"context"

	nats "github.com/nats-io/nats.go"
	"go.opencensus.io/trace"
)

const DefaultURI = "localhost:4222"

type Client struct {
	nats *nats.Conn
}

func NewClient(natsURI string) (*Client, error) {
	nc, err := nats.Connect(natsURI)
	if err != nil {
		return nil, err
	}

	return &Client{nc}, nil
}

type TopicEncoder interface {
	TopicName() string
	Encode() ([]byte, error)
}

func (c *Client) Publish(ctx context.Context, msg TopicEncoder) error {
	ctx, sp := trace.StartSpan(ctx, "golo/nats/Publish")
	defer sp.End()

	b, err := msg.Encode()
	if err != nil {
		return err
	}
	b = Wrap(ctx, b) // OpenCensus span context

	sp.AddAttributes(
		trace.StringAttribute("message", string(b)),
		trace.StringAttribute("topic", msg.TopicName()),
	)
	msg.TopicName()

	return c.nats.Publish("create.loyyalchannel", b)
}
