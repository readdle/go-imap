package imapclient

import (
	"bufio"
	"bytes"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/internal/imapwire"
)

func TestWriteSearchKeyDates(t *testing.T) {
	date := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		criteria imap.SearchCriteria
		format   imap.SearchDateFormat
		want     string
	}{
		{
			name:     "default",
			criteria: imap.SearchCriteria{Since: date},
			want:     `SINCE "1-Sep-2026"`,
		},
		{
			name:     "since",
			criteria: imap.SearchCriteria{Since: date},
			format:   imap.SearchDateFormatUnquoted,
			want:     "SINCE 01-Sep-2026",
		},
		{
			name:     "before",
			criteria: imap.SearchCriteria{Before: date},
			format:   imap.SearchDateFormatUnquoted,
			want:     "BEFORE 01-Sep-2026",
		},
		{
			name: "on",
			criteria: imap.SearchCriteria{
				Since:  date,
				Before: date.Add(24 * time.Hour),
			},
			format: imap.SearchDateFormatUnquoted,
			want:   "ON 01-Sep-2026",
		},
		{
			name:     "sent since",
			criteria: imap.SearchCriteria{SentSince: date},
			format:   imap.SearchDateFormatUnquoted,
			want:     "SENTSINCE 01-Sep-2026",
		},
		{
			name:     "sent before",
			criteria: imap.SearchCriteria{SentBefore: date},
			format:   imap.SearchDateFormatUnquoted,
			want:     "SENTBEFORE 01-Sep-2026",
		},
		{
			name: "sent on",
			criteria: imap.SearchCriteria{
				SentSince:  date,
				SentBefore: date.Add(24 * time.Hour),
			},
			format: imap.SearchDateFormatUnquoted,
			want:   "SENTON 01-Sep-2026",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := imapwire.NewEncoder(bufio.NewWriter(&buf), imapwire.ConnSideClient)
			writeSearchKey(enc, &tt.criteria, tt.format)
			if err := enc.CRLF(); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); got != tt.want+"\r\n" {
				t.Errorf("writeSearchKey() = %q, want %q", got, tt.want+"\r\n")
			}
		})
	}
}
