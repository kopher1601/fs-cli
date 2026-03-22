package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ListDocuments lists documents in a FileSearchStore.
func (c *Client) ListDocuments(ctx context.Context, storeName string, pageSize int, pageToken string) (*ListDocumentsResponse, error) {
	params := url.Values{}
	if pageSize > 0 {
		params.Set("pageSize", strconv.Itoa(pageSize))
	}
	if pageToken != "" {
		params.Set("pageToken", pageToken)
	}

	var resp ListDocumentsResponse
	path := fmt.Sprintf("%s/documents", storeName)
	err := c.get(ctx, path, params, &resp)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	return &resp, nil
}

// GetDocument retrieves a specific document.
func (c *Client) GetDocument(ctx context.Context, docName string) (*Document, error) {
	var doc Document
	err := c.get(ctx, docName, nil, &doc)
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}
	return &doc, nil
}

// DeleteDocument deletes a document from a FileSearchStore.
func (c *Client) DeleteDocument(ctx context.Context, docName string, force bool) error {
	params := url.Values{}
	if force {
		params.Set("force", "true")
	}
	err := c.delete(ctx, docName, params)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	return nil
}

