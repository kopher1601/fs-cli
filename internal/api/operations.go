package api

import (
	"context"
	"fmt"
)

// GetOperation retrieves the status of a long-running operation.
func (c *Client) GetOperation(ctx context.Context, opName string) (*Operation, error) {
	var op Operation
	err := c.get(ctx, opName, nil, &op)
	if err != nil {
		return nil, fmt.Errorf("get operation: %w", err)
	}
	return &op, nil
}
