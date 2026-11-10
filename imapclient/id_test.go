package imapclient_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

func TestID_ResponseParsing(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     imap.IDData
	}{
		{
			name:     "list",
			response: `* ID ("name" "Cyrus" "version" "3.8")`,
			want:     imap.IDData{Name: "Cyrus", Version: "3.8"},
		},
		{
			// RFC 2971 defines a field value as an nstring.
			name:     "NIL value",
			response: `* ID ("name" "Cyrus" "version" NIL "vendor" "Project Cyrus")`,
			want:     imap.IDData{Name: "Cyrus", Vendor: "Project Cyrus"},
		},
		{
			name:     "unknown keys",
			response: `* ID ("name" "Y" "host" "imap.example.com" "remote-host" "203.0.113.7")`,
			want:     imap.IDData{Name: "Y"},
		},
		{
			name:     "NIL",
			response: `* ID NIL`,
			want:     imap.IDData{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientConn, serverConn := net.Pipe()
			_ = clientConn.SetDeadline(time.Now().Add(10 * time.Second))
			_ = serverConn.SetDeadline(time.Now().Add(10 * time.Second))

			client := imapclient.New(clientConn, nil)
			server := newScriptedServer(t, serverConn)
			defer server.Close()

			serverDone := make(chan error, 1)
			go func() {
				serverDone <- func() error {
					if err := server.writeLine("* OK Server ready"); err != nil {
						return err
					}
					tag, body, err := server.expectCommand()
					if err != nil {
						return fmt.Errorf("read ID: %w", err)
					}
					if want := `ID ("name" "go-imap")`; body != want {
						return fmt.Errorf("command = %q, want %q", body, want)
					}
					return server.writeLines(tt.response, tag+" OK ID completed")
				}()
			}()

			data, err := client.ID(&imap.IDData{Name: "go-imap"}).Wait()
			if err != nil {
				t.Fatalf("ID() = %v", err)
			}
			if *data != tt.want {
				t.Errorf("ID() = %+v, want %+v", *data, tt.want)
			}
			if err := <-serverDone; err != nil {
				t.Fatalf("server: %v", err)
			}

			// A parsed response must leave no error behind in the decoder: a
			// stale one replaces the real cause of every later read failure.
			if err := client.Close(); err != nil {
				t.Errorf("Close() = %v", err)
			}
		})
	}
}
