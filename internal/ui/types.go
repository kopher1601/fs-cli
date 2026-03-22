package ui

import "github.com/kopher1601/fs-cli/internal/api"

// ViewType represents the type of view currently displayed.
type ViewType int

const (
	ViewStores ViewType = iota
	ViewDocuments
	ViewStoreDetail
	ViewDocDetail
	ViewUpload
	ViewOperations
	ViewHelp
)

// NavigateMsg requests a view change.
type NavigateMsg struct {
	View  ViewType
	Store *api.FileSearchStore
	Doc   *api.Document
}

// BackMsg requests navigating back.
type BackMsg struct{}

// FlashMsg displays a temporary message.
type FlashMsg struct {
	Message string
	Level   FlashLevel
}

// FlashLevel represents the severity of a flash message.
type FlashLevel int

const (
	FlashInfo FlashLevel = iota
	FlashSuccess
	FlashWarn
	FlashError
)

// ErrMsg represents an API error.
type ErrMsg struct {
	Err error
}

func (e ErrMsg) Error() string { return e.Err.Error() }

// RefreshMsg requests a data refresh.
type RefreshMsg struct{}
