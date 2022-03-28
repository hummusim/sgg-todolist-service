package healthcheck

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	pbPaymentGateway "github.com/overridesh/sgg-todolist-service/proto"
)

// Backend implements the protobuf interface
type healthcheck struct{}

// New initializes a new Healthcheck struct.
func NewGRPC() pbPaymentGateway.HealthcheckServiceServer {
	return &healthcheck{}
}

// Healthcheck
func (b *healthcheck) GetHealthcheck(ctx context.Context, _ *emptypb.Empty) (*pbPaymentGateway.GetHealthcheckResponse, error) {
	return &pbPaymentGateway.GetHealthcheckResponse{
		Ok: true,
	}, nil
}
