// internal/atproto/client_test.go
package atproto

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "empty config",
			cfg:     Config{},
			wantErr: true,
		},
		{
			name: "valid config",
			cfg: Config{
				PdsHost: "https://bsky.social",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Errorf("NewClient() returned nil client with no error")
			}
			if !tt.wantErr && client.PdsHost() != tt.cfg.PdsHost {
				t.Errorf("NewClient() PdsHost = %v, want %v", client.PdsHost(), tt.cfg.PdsHost)
			}
		})
	}
}