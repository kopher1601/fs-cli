package api

import "encoding/json"

// FileSearchStore represents a Google Gemini File Search store.
type FileSearchStore struct {
	Name                  string `json:"name"`
	DisplayName           string `json:"displayName,omitempty"`
	CreateTime            string `json:"createTime,omitempty"`
	UpdateTime            string `json:"updateTime,omitempty"`
	ActiveDocumentsCount  int64 `json:"activeDocumentsCount,omitempty,string"`
	PendingDocumentsCount int64 `json:"pendingDocumentsCount,omitempty,string"`
	FailedDocumentsCount  int64 `json:"failedDocumentsCount,omitempty,string"`
	SizeBytes             int64 `json:"sizeBytes,omitempty,string"`
}

// Document represents a document within a FileSearchStore.
type Document struct {
	Name           string           `json:"name"`
	DisplayName    string           `json:"displayName,omitempty"`
	CustomMetadata []CustomMetadata `json:"customMetadata,omitempty"`
	CreateTime     string           `json:"createTime,omitempty"`
	UpdateTime     string           `json:"updateTime,omitempty"`
	State          string           `json:"state,omitempty"`
	SizeBytes      int64            `json:"sizeBytes,omitempty,string"`
	MimeType       string           `json:"mimeType,omitempty"`
}

// CustomMetadata is a key-value pair attached to a document.
type CustomMetadata struct {
	Key          string  `json:"key"`
	StringValue  *string `json:"stringValue,omitempty"`
	NumericValue *float64 `json:"numericValue,omitempty"`
}

// Operation represents a long-running operation.
type Operation struct {
	Name     string          `json:"name"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
	Done     bool            `json:"done"`
	Error    *Status         `json:"error,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
}

// Status represents an error status from the API.
type Status struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ListStoresResponse is the response for listing file search stores.
type ListStoresResponse struct {
	FileSearchStores []FileSearchStore `json:"fileSearchStores"`
	NextPageToken    string            `json:"nextPageToken,omitempty"`
}

// ListDocumentsResponse is the response for listing documents.
type ListDocumentsResponse struct {
	Documents     []Document `json:"documents"`
	NextPageToken string     `json:"nextPageToken,omitempty"`
}

// CreateStoreRequest is the request body for creating a store.
type CreateStoreRequest struct {
	DisplayName string `json:"displayName,omitempty"`
}

// UploadConfig holds configuration for file uploads.
type UploadConfig struct {
	DisplayName    string           `json:"displayName,omitempty"`
	CustomMetadata []CustomMetadata `json:"customMetadata,omitempty"`
	ChunkingConfig *ChunkingConfig  `json:"chunkingConfig,omitempty"`
}

// ChunkingConfig configures how documents are chunked.
type ChunkingConfig struct {
	WhiteSpaceConfig *WhiteSpaceConfig `json:"whiteSpaceConfig,omitempty"`
}

// WhiteSpaceConfig configures whitespace-based chunking.
type WhiteSpaceConfig struct {
	MaxTokensPerChunk int `json:"maxTokensPerChunk,omitempty"`
	MaxOverlapTokens  int `json:"maxOverlapTokens,omitempty"`
}

// APIError represents an error response from the API.
type APIError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}
