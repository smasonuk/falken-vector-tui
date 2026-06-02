package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func StatusViewFromSDK(directory string, status falkenvector.Status) StatusViewModel {
	vm := StatusViewModel{
		Loaded:        true,
		Directory:     directory,
		Documents:     status.Documents,
		Chunks:        status.Chunks,
		LastIndexedAt: status.LastIndexedAt,
	}
	vm.Text = FormatStatus(vm)
	return vm
}

func NoIndexStatus(directory string) StatusViewModel {
	return StatusViewModel{
		Loaded:    true,
		NoIndex:   true,
		Directory: directory,
		Text:      fmt.Sprintf("This directory has not been indexed yet.\n\nDirectory: %s\n\nCLI equivalent: falkengo ingest .", directory),
	}
}

func FormatStatus(status StatusViewModel) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Directory: %s\n\n", status.Directory)
	fmt.Fprintf(&b, "Documents: indexed %d · error %d · deleted %d\n", status.Documents.Indexed, status.Documents.Error, status.Documents.Deleted)
	fmt.Fprintf(&b, "Chunks:    active %d · inactive %d\n", status.Chunks.Active, status.Chunks.Inactive)
	if status.LastIndexedAt == nil || status.LastIndexedAt.IsZero() {
		b.WriteString("Last indexed: never")
	} else {
		b.WriteString("Last indexed: ")
		b.WriteString(status.LastIndexedAt.Local().Format("2006-01-02 15:04"))
	}
	return b.String()
}

func IsNoIndexError(err error) bool {
	return errors.Is(err, falkenvector.ErrNoIndex)
}

func statusTime(year int, month time.Month, day int, hour int, min int) *time.Time {
	t := time.Date(year, month, day, hour, min, 0, 0, time.Local)
	return &t
}
