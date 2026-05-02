package threads

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toozej/rss2socials/pkg/config"
)

func TestPost_MissingCredentials(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		token        string
	}{
		{name: "Missing Client ID", clientID: "", clientSecret: "secret123", token: "token123"},
		{name: "Missing Client Secret", clientID: "clientid123", clientSecret: "", token: "token123"},
		{name: "Both Missing", clientID: "", clientSecret: "", token: "token123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := config.Config{
				ThreadsClientID:     tt.clientID,
				ThreadsClientSecret: tt.clientSecret,
				ThreadsToken:        tt.token,
			}
			err := Post(context.Background(), conf, "Hello Threads")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "client ID and client secret are required")
		})
	}
}

func TestNewClient_MissingCredentials(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		clientSecret string
	}{
		{name: "Missing Client ID", clientID: "", clientSecret: "secret123"},
		{name: "Missing Client Secret", clientID: "clientid123", clientSecret: ""},
		{name: "Both Missing", clientID: "", clientSecret: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := config.Config{
				ThreadsClientID:     tt.clientID,
				ThreadsClientSecret: tt.clientSecret,
			}
			client, err := NewClient(conf)
			assert.Error(t, err)
			assert.Nil(t, client)
			assert.Contains(t, err.Error(), "client ID and client secret are required")
		})
	}
}

func TestNewClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	conf := config.Config{
		ThreadsClientID:     "test-client-id",
		ThreadsClientSecret: "test-client-secret",
		ThreadsRedirectURI:  "https://example.com/callback",
		ThreadsToken:        "test-token",
	}
	client, err := NewClient(conf)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestPost_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	conf := config.Config{
		ThreadsClientID:     "test-client-id",
		ThreadsClientSecret: "test-client-secret",
		ThreadsRedirectURI:  "https://example.com/callback",
		ThreadsToken:        "test-token",
	}
	err := Post(context.Background(), conf, "Integration test post")
	assert.Error(t, err)
}
