package tools

import (
	"context"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func SetStatusCode(ctx context.Context, statusCode int) error {
	return grpc.SetHeader(ctx, metadata.Pairs("X-http-code", strconv.Itoa(statusCode)))
}
