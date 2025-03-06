// internal/atproto/client.go
package atproto

import (
	"context"
	"errors"
	"fmt"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
)

// Client represents an AT Protocol client
type Client struct {
	xrpcClient *xrpc.Client
	pdsHost    string
	loggedIn   bool
}

// Config holds the configuration for an AT Protocol client
type Config struct {
	PdsHost string // PDS host, e.g., "https://bsky.social"
}

// NewClient creates a new AT Protocol client
func NewClient(cfg Config) (*Client, error) {
	if cfg.PdsHost == "" {
		return nil, errors.New("PDS host is required")
	}

	xrpcClient := &xrpc.Client{
		Host: cfg.PdsHost,
	}

	return &Client{
		xrpcClient: xrpcClient,
		pdsHost:    cfg.PdsHost,
		loggedIn:   false,
	}, nil
}

// Login authenticates with the PDS using app password
func (c *Client) Login(ctx context.Context, identifier, password string) error {
	if identifier == "" || password == "" {
		return errors.New("identifier and password are required")
	}

	session, err := atproto.ServerCreateSession(ctx, c.xrpcClient, &atproto.ServerCreateSession_Input{
		Identifier: identifier,
		Password:   password,
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Update client with authentication info
	c.xrpcClient.Auth = &xrpc.AuthInfo{
		AccessJwt:  session.AccessJwt,
		RefreshJwt: session.RefreshJwt,
		Handle:     session.Handle,
		Did:        session.Did,
	}
	c.loggedIn = true

	return nil
}

// GetProfile fetches a user's profile by handle
func (c *Client) GetProfile(ctx context.Context, handle string) (*bsky.ActorDefs_ProfileViewDetailed, error) {
	if !c.loggedIn {
		return nil, errors.New("client not authenticated")
	}

	profile, err := bsky.ActorGetProfile(ctx, c.xrpcClient, handle)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return profile, nil
}

// IsLoggedIn returns whether the client is currently authenticated
func (c *Client) IsLoggedIn() bool {
	return c.loggedIn
}

// PdsHost returns the current PDS host
func (c *Client) PdsHost() string {
	return c.pdsHost
}