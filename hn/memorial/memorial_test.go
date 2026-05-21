package memorial

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	withMemorialBar = `<table><tr><td bgcolor="#000000"><img height="5"></td></tr>` +
		`<tr><td bgcolor="#ff6600">header</td></tr></table>`
	withoutMemorialBar = `<table><tr><td bgcolor="#ff6600">header</td></tr></table>`
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		status  int
		want    bool
		wantErr bool
	}{
		{name: "black bar present", body: withMemorialBar, status: http.StatusOK, want: true},
		{name: "no black bar", body: withoutMemorialBar, status: http.StatusOK, want: false},
		{name: "server error", body: "", status: http.StatusInternalServerError, wantErr: true},
		// Markers past the read limit are never seen, so the bar reads as absent.
		{name: "black bar beyond read limit", body: strings.Repeat("x", readLimit+100) + withMemorialBar, status: http.StatusOK, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			old := baseURL

			baseURL = server.URL
			defer func() { baseURL = old }()

			got, err := Detect(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				assert.False(t, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetect_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer server.Close()

	old := baseURL

	baseURL = server.URL
	defer func() { baseURL = old }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got, err := Detect(ctx)

	require.Error(t, err)
	assert.False(t, got)
}

func TestDetect_Timeout(t *testing.T) {
	// Handler blocks until the client gives up, so the request fails on the
	// caller's deadline rather than a server response.
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	old := baseURL

	baseURL = server.URL
	defer func() { baseURL = old }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	got, err := Detect(ctx)

	require.Error(t, err)
	assert.False(t, got)
}

func TestHasMemorialBar(t *testing.T) {
	assert.True(t, hasMemorialBar(withMemorialBar))
	assert.False(t, hasMemorialBar(withoutMemorialBar))
	assert.False(t, hasMemorialBar(""))
	// Orange before black should not count.
	assert.False(t, hasMemorialBar(`<tr bgcolor="#ff6600"></tr><tr bgcolor="#000000"></tr>`))
}
