package pingservice

import (
	"context"

	"github.com/nazifbara/flagsvc/gen/go/flagsvc/ping/v1"
)

type PingService struct{}

func (p PingService) Ping(ctx context.Context, req *ping.PingRequest) (*ping.PingResponse, error) {
	return &ping.PingResponse{
		Message: req.Message,
	}, nil
}
