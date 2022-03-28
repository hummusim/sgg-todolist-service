package tools

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/status"
)

func TestGetValidUUID(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (uuid.UUID, error)
		expect *status.Status
	}{
		{
			name: "Test_Success",
			input: func() (uuid.UUID, error) {
				return GetValidUUID(uuid.NewV4().String())
			},
			expect: nil,
		},
		{
			name: "Test_ErrStatusIdMustBeUUID",
			input: func() (uuid.UUID, error) {
				return GetValidUUID("ASD")
			},
			expect: ErrStatusIdMustBeUUID,
		},
		{
			name: "Test_ErrStatusIdMusBeValidUUID",
			input: func() (uuid.UUID, error) {
				return GetValidUUID(uuid.Nil.String())
			},
			expect: ErrStatusIdMusBeValidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.input()
			errMsg, _ := status.FromError(err)
			if errMsg.Code() != tt.expect.Code() {
				t.Errorf("expect values are equals, but got diferent, output: %v, expect: %v", err, tt.expect)
			}
		})
	}
}
