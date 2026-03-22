package model

import "github.com/kopher1601/fs-cli/internal/api"

// OpRow represents a formatted row for the operations table.
type OpRow struct {
	Name   string
	Store  string
	Type   string
	Status string
}

// OpColumns returns the column headers for the operations table.
func OpColumns() []string {
	return []string{"OPERATION", "STORE", "TYPE", "STATUS"}
}

// TrackedOp tracks an in-progress operation.
type TrackedOp struct {
	Operation *api.Operation
	StoreName string
	OpType    string // "Upload"
}

// OpStatus returns a human-readable status.
func OpStatus(op *api.Operation) string {
	if op.Done {
		if op.Error != nil {
			return "Failed"
		}
		return "Done"
	}
	return "Running"
}
