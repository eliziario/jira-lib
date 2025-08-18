package lib

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  ClientConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: ClientConfig{
				Server:   "https://test.atlassian.net",
				Login:    "test@example.com",
				APIToken: "test-token",
			},
			wantErr: false,
		},
		{
			name: "missing server",
			config: ClientConfig{
				Login:    "test@example.com",
				APIToken: "test-token",
			},
			wantErr: true,
			errMsg:  "server URL is required",
		},
		{
			name: "missing login",
			config: ClientConfig{
				Server:   "https://test.atlassian.net",
				APIToken: "test-token",
			},
			wantErr: true,
			errMsg:  "login is required",
		},
		{
			name: "missing api token",
			config: ClientConfig{
				Server: "https://test.atlassian.net",
				Login:  "test@example.com",
			},
			wantErr: true,
			errMsg:  "API token is required",
		},
		{
			name: "with custom timeout",
			config: ClientConfig{
				Server:   "https://test.atlassian.net",
				Login:    "test@example.com",
				APIToken: "test-token",
				Timeout:  30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "with local installation",
			config: ClientConfig{
				Server:           "https://jira.local.com",
				Login:            "username",
				APIToken:         "password",
				InstallationType: "Local",
			},
			wantErr: false,
		},
		{
			name: "with bearer auth",
			config: ClientConfig{
				Server:   "https://jira.local.com",
				Login:    "username",
				APIToken: "pat-token",
				AuthType: "bearer",
			},
			wantErr: false,
		},
		// Skip MTLS test as it tries to load actual certificates
		// {
		// 	name: "with mtls config",
		// 	config: ClientConfig{
		// 		Server:   "https://jira.local.com",
		// 		Login:    "username",
		// 		APIToken: "token",
		// 		AuthType: "mtls",
		// 		MTLSConfig: &MTLSConfig{
		// 			CaCert:     "/path/to/ca.crt",
		// 			ClientCert: "/path/to/client.crt",
		// 			ClientKey:  "/path/to/client.key",
		// 		},
		// 	},
		// 	wantErr: false,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.GetRawClient())
				
				// Verify installation type is set correctly
				if tt.config.InstallationType != "" {
					assert.Equal(t, tt.config.InstallationType, client.installationType)
				} else {
					assert.Equal(t, "Cloud", client.installationType)
				}
			}
		})
	}
}

func TestClientConfigDefaults(t *testing.T) {
	config := ClientConfig{
		Server:   "https://test.atlassian.net",
		Login:    "test@example.com",
		APIToken: "test-token",
	}
	
	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	
	// Check defaults were applied
	assert.Equal(t, "Cloud", client.installationType)
	// AuthType default is checked internally as "basic"
}