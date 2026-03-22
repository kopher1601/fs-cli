package model

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/kopher1601/fs-cli/internal/api"
)

// DocRow represents a formatted row for the documents table.
type DocRow struct {
	ID          string
	DisplayName string
	State       string
	MimeType    string
	Size        string
	Age         string
}

// DocColumns returns the column headers for the documents table.
func DocColumns() []string {
	return []string{"NAME", "DISPLAY NAME", "STATE", "MIME TYPE", "SIZE", "AGE"}
}

// DocToRow converts an API Document to a display row.
func DocToRow(d *api.Document) DocRow {
	return DocRow{
		ID:          ShortName(d.Name),
		DisplayName: d.DisplayName,
		State:       d.State,
		MimeType:    d.MimeType,
		Size:        humanize.IBytes(uint64(d.SizeBytes)),
		Age:         FormatAge(d.CreateTime),
	}
}

// DocRowToStrings converts a DocRow to a string slice for table rendering.
func DocRowToStrings(r DocRow) []string {
	return []string{r.ID, r.DisplayName, r.State, r.MimeType, r.Size, r.Age}
}

// FormatMetadata formats custom metadata for display.
func FormatMetadata(meta []api.CustomMetadata) []string {
	var lines []string
	for _, m := range meta {
		if m.StringValue != nil {
			lines = append(lines, fmt.Sprintf("  %s: %s", m.Key, *m.StringValue))
		} else if m.NumericValue != nil {
			lines = append(lines, fmt.Sprintf("  %s: %g", m.Key, *m.NumericValue))
		}
	}
	return lines
}
