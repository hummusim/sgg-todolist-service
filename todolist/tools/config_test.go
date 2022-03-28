package tools

import (
	"os"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestGetConfig(t *testing.T) {
	tests := []struct {
		name    string
		input   func() error
		wantErr bool
	}{
		{
			name: "GetConfig_Success",
			input: func() error {
				defer os.Clearenv()
				config := struct {
					Test string `envconfig:"TEST" required:"true"`
				}{}

				os.Setenv("TEST", uuid.NewV4().String())
				return GetConfig("", &config)
			},
			wantErr: false,
		},
		{
			name: "GetConfig_Failed",
			input: func() error {
				defer os.Clearenv()
				config := struct {
					Test string `envconfig:"TEST" required:"true"`
				}{}
				return GetConfig("", &config)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input()
			if !tt.wantErr && err != nil {
				t.Errorf("expect error nil, but got %v", err)
			}
		})
	}
}
