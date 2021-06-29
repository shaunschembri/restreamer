package provider

import (
	"context"
	"time"
)

type Segment struct {
	URL       string
	KeyMethod string
	KeyURL    string
	IV        string
	Duration  float64
}

type Provider interface {
	Get(ctx context.Context, bandwidth uint32) ([]Segment, time.Duration, error)
	Info() string
}
