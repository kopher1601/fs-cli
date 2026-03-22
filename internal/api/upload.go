package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
)

// ExpandHome expands a leading ~ to the user's home directory.
func ExpandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		if home, err := os.UserHomeDir(); err == nil {
			return home + path[1:]
		}
	}
	return path
}

// UploadFile uploads a file to a FileSearchStore.
func (c *Client) UploadFile(ctx context.Context, storeName, filePath string, config *UploadConfig) (*Operation, error) {
	filePath = ExpandHome(filePath)
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Write metadata part
	if config != nil {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Type", "application/json")
		metaPart, err := writer.CreatePart(h)
		if err != nil {
			return nil, fmt.Errorf("create metadata part: %w", err)
		}
		metaJSON, err := json.Marshal(config)
		if err != nil {
			return nil, fmt.Errorf("marshal config: %w", err)
		}
		if _, err := metaPart.Write(metaJSON); err != nil {
			return nil, fmt.Errorf("write metadata: %w", err)
		}
	}

	// Write file part
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("create file part: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer: %w", err)
	}

	url := fmt.Sprintf("%s/%s:uploadToFileSearchStore?key=%s", c.uploadURL, storeName, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute upload: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error.Message != "" {
			return nil, fmt.Errorf("upload error %d (%s): %s", apiErr.Error.Code, apiErr.Error.Status, apiErr.Error.Message)
		}
		return nil, fmt.Errorf("upload error %d: %s", resp.StatusCode, string(respBody))
	}

	var op Operation
	if err := json.Unmarshal(respBody, &op); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &op, nil
}
