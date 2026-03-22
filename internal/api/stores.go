package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// CreateStore creates a new FileSearchStore.
func (c *Client) CreateStore(ctx context.Context, displayName string) (*FileSearchStore, error) {
	var store FileSearchStore
	err := c.post(ctx, "fileSearchStores", &CreateStoreRequest{DisplayName: displayName}, &store)
	if err != nil {
		return nil, fmt.Errorf("create store: %w", err)
	}
	return &store, nil
}

// ListStores lists all FileSearchStores with pagination.
func (c *Client) ListStores(ctx context.Context, pageSize int, pageToken string) (*ListStoresResponse, error) {
	params := url.Values{}
	if pageSize > 0 {
		params.Set("pageSize", strconv.Itoa(pageSize))
	}
	if pageToken != "" {
		params.Set("pageToken", pageToken)
	}

	var resp ListStoresResponse
	err := c.get(ctx, "fileSearchStores", params, &resp)
	if err != nil {
		return nil, fmt.Errorf("list stores: %w", err)
	}
	return &resp, nil
}

// GetStore retrieves a specific FileSearchStore.
func (c *Client) GetStore(ctx context.Context, name string) (*FileSearchStore, error) {
	var store FileSearchStore
	err := c.get(ctx, name, nil, &store)
	if err != nil {
		return nil, fmt.Errorf("get store: %w", err)
	}
	return &store, nil
}

// DeleteStore deletes a FileSearchStore.
func (c *Client) DeleteStore(ctx context.Context, name string, force bool) error {
	params := url.Values{}
	if force {
		params.Set("force", "true")
	}
	err := c.delete(ctx, name, params)
	if err != nil {
		return fmt.Errorf("delete store: %w", err)
	}
	return nil
}
