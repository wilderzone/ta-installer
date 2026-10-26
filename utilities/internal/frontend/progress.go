package frontend

import (
	"fmt"
	"strings"
)

// PROGRESS_SCALE is the progress dialog's maximum value. Finer than a percentage, so a large
// download still visibly moves.
const PROGRESS_SCALE = 1000

// line is one of the JSON lines the CLI writes under `-m` and `-p`. Only the fields a
// utility reads are declared.
type line struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Index   int    `json:"index,omitempty"`
	Total   int    `json:"total,omitempty"`
	Bytes   uint64 `json:"bytes,omitempty"`
	Size    uint64 `json:"size,omitempty"`
}

// readable renders a line the way the CLI would have printed it without `-m`. It returns
// an empty string for the lines that only exist for an interface to read.
func (l line) readable() string {
	switch l.Type {
		case "DOWNLOAD_STARTED", "DOWNLOAD_PROGRESS", "DOWNLOAD_COMPLETE":
			return ""

		case "PROGRESS":
			return fmt.Sprintf("[%d/%d] %s", l.Index, l.Total, l.Message)
	}
	return l.Message
}

// progress is what the progress dialog shows, kept up to date from the CLI's lines.
type progress struct {
	value       int // Out of PROGRESS_SCALE.
	text        string
	downloading bool
}

// apply updates the progress from one of the CLI's lines, and reports whether anything
// visible changed.
//
// While a measured download is in flight its bytes drive the bar, because they move far
// more smoothly than the count of finished items. Between downloads, the count does.
func (p *progress) apply(l line) bool {
	before := *p
	switch l.Type {
		case "DOWNLOAD_STARTED", "DOWNLOAD_PROGRESS":
			p.downloading = true
			if l.Size != 0 {
				p.value = int(min(l.Bytes, l.Size) * PROGRESS_SCALE / l.Size)
			}
			p.text = "Downloading..."
			if l.Message != "" {
				p.text += " " + l.Message
			}

		case "DOWNLOAD_COMPLETE":
			p.downloading = false

		case "PROGRESS":
			if p.downloading {
				break
			}
			if l.Total != 0 {
				p.value = l.Index * PROGRESS_SCALE / l.Total
			}
			p.text = l.readable()

		default:
			if p.downloading || l.Message == "" {
				break
			}
			p.text = l.Message
	}

	// A full bar enables the dialog's OK button, and pressing that would close the dialog
	// as surely as Cancel does - so the bar stops just short until the work is over.
	p.value = min(p.value, PROGRESS_SCALE - 1)
	// The text is sent to a Linux dialog tool as a line of its own, so it must stay on one.
	p.text = strings.Join(strings.Fields(p.text), " ")

	return *p != before
}
