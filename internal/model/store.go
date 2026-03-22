package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/kopher1601/fs-cli/internal/api"
)

// StoreRow represents a formatted row for the stores table.
type StoreRow struct {
	ID          string
	DisplayName string
	ActiveDocs  string
	PendingDocs string
	FailedDocs  string
	Size        string
	Age         string
}

// StoreColumns returns the column headers for the stores table.
func StoreColumns() []string {
	return []string{"NAME", "DISPLAY NAME", "ACTIVE", "PENDING", "FAILED", "SIZE", "AGE"}
}

// StoreToRow converts an API FileSearchStore to a display row.
func StoreToRow(s *api.FileSearchStore) StoreRow {
	return StoreRow{
		ID:          ShortName(s.Name),
		DisplayName: s.DisplayName,
		ActiveDocs:  fmt.Sprintf("%d", s.ActiveDocumentsCount),
		PendingDocs: fmt.Sprintf("%d", s.PendingDocumentsCount),
		FailedDocs:  fmt.Sprintf("%d", s.FailedDocumentsCount),
		Size:        humanize.IBytes(uint64(s.SizeBytes)),
		Age:         FormatAge(s.CreateTime),
	}
}

// StoreRowToStrings converts a StoreRow to a string slice for table rendering.
func StoreRowToStrings(r StoreRow) []string {
	return []string{r.ID, r.DisplayName, r.ActiveDocs, r.PendingDocs, r.FailedDocs, r.Size, r.Age}
}

// ShortName extracts the short ID from a full resource name.
// e.g., "fileSearchStores/abc123" -> "abc123"
func ShortName(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return name
}

// FormatAge converts an RFC3339 timestamp to a relative time string.
func FormatAge(timestamp string) string {
	if timestamp == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}
	return humanize.Time(t)
}
