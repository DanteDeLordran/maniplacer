package cli

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

type notifyingWriter struct {
	bytes.Buffer
	written chan struct{}
}

func (w *notifyingWriter) Write(p []byte) (int, error) {
	n, err := w.Buffer.Write(p)
	select {
	case <-w.written:
	default:
		close(w.written)
	}
	return n, err
}

func TestParseDocsPort(t *testing.T) {
	tests := []struct {
		value   string
		want    int
		wantErr bool
	}{
		{value: "1", want: 1},
		{value: "8000", want: 8000},
		{value: "65535", want: 65535},
		{value: "0", wantErr: true},
		{value: "65536", wantErr: true},
		{value: "-1", wantErr: true},
		{value: "http", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := parseDocsPort(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseDocsPort(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseDocsPort(%q) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestDocsHandler(t *testing.T) {
	handler := docsHandler()
	tests := []struct {
		name         string
		method       string
		path         string
		wantStatus   int
		wantLocation string
	}{
		{name: "root redirects", method: http.MethodGet, path: "/", wantStatus: http.StatusTemporaryRedirect, wantLocation: "/docs/"},
		{name: "docs redirects to trailing slash", method: http.MethodGet, path: "/docs", wantStatus: http.StatusTemporaryRedirect, wantLocation: "/docs/"},
		{name: "docs page", method: http.MethodGet, path: "/docs/", wantStatus: http.StatusOK},
		{name: "unknown docs path", method: http.MethodGet, path: "/docs/missing", wantStatus: http.StatusNotFound},
		{name: "method not allowed", method: http.MethodPost, path: "/docs/", wantStatus: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(tt.method, tt.path, nil))

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if location := recorder.Header().Get("Location"); location != tt.wantLocation {
				t.Errorf("Location = %q, want %q", location, tt.wantLocation)
			}
			if tt.wantStatus == http.StatusOK {
				if contentType := recorder.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
					t.Errorf("Content-Type = %q", contentType)
				}
				if body := recorder.Body.String(); !strings.Contains(body, "The canonical workflow") || !strings.Contains(body, "Base64 is encoding, not encryption") {
					t.Error("docs page is missing workflow or secret-safety guidance")
				}
			}
		})
	}
}

func TestRunDocsServerBindsBeforeReportingSuccess(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not reserve test port: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	var output bytes.Buffer
	if err := runDocsServer(context.Background(), port, &output); err == nil {
		t.Fatal("runDocsServer() error = nil on an occupied port")
	}
	if output.Len() != 0 {
		t.Errorf("server reported success before binding: %q", output.String())
	}
}

func TestRunDocsServerShutsDownOnCancellation(t *testing.T) {
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not allocate test port: %v", err)
	}
	port := probe.Addr().(*net.TCPAddr).Port
	if err := probe.Close(); err != nil {
		t.Fatalf("could not release test port: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	output := &notifyingWriter{written: make(chan struct{})}
	result := make(chan error, 1)
	go func() {
		result <- runDocsServer(ctx, port, output)
	}()

	select {
	case <-output.written:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not report startup")
	}

	wantURL := fmt.Sprintf("http://127.0.0.1:%s/docs/", strconv.Itoa(port))
	if !strings.Contains(output.String(), wantURL) {
		t.Errorf("startup output %q does not contain %q", output.String(), wantURL)
	}

	cancel()
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("runDocsServer() returned an error during shutdown: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down after context cancellation")
	}
}
