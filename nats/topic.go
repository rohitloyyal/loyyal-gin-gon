package nats

import (
	"context"
	"log"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)


func Wrap(ctx context.Context, b []byte) []byte {
	oc := propagation.Binary(trace.FromContext(ctx).SpanContext())
	if len(oc) == 0 {
		return append([]byte{0}, b...)
	}
	if len(oc) > 255 {
		log.Fatalf("span context too big: %d (max 255)", len(oc)) // catch in C/I
	}
	size := byte(len(oc))
	a := make([]byte, 0, len(b)+len(oc)+1)
	a = append([]byte{size}, oc...)
	a = append(a, b...)
	return a
}

