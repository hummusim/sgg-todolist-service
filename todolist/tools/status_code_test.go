package tools

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSetStatusCode(t *testing.T) {
	tests := []struct {
		name   string
		input  func() error
		expect *status.Status
	}{
		{
			name: "SetHeader_error",
			input: func() error {
				return SetStatusCode(context.Background(), http.StatusServiceUnavailable)
			},
			expect: status.New(codes.Unknown, "grpc: failed to fetch the stream from the context context.Background"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input()
			if err != nil {
				errMsg := status.FromContextError(err)
				if errMsg.Code() != tt.expect.Code() {
					t.Errorf("expect values are equals, but got diferent, output: %v, expect: %v", err, tt.expect)
				}
			}
		})
	}
}
