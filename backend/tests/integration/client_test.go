//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"employee-portal/configs"
	"employee-portal/modules/server"
)

// apiClient talks to a real server over HTTP with one user's session cookie.
type apiClient struct {
	t       *testing.T
	base    string
	cookies []*http.Cookie
}

func newTestServer(t *testing.T) string {
	t.Helper()
	srv, err := server.New(configs.Config{
		JWTSecret: "integration-test-secret-0123456789", JWTTTL: 3600e9, HistoryDays: 7,
		CORSOrigins: []string{"http://localhost:5173"},
		UploadDir:   t.TempDir(), UploadMaxBytes: 5 << 20,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)), appPool)
	require.NoError(t, err)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts.URL
}

func login(t *testing.T, base, code string) *apiClient {
	t.Helper()
	c := &apiClient{t: t, base: base}
	res := c.do("POST", "/api/v1/auth/login", map[string]string{"employeeCode": code, "password": "demo1234"}, nil)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
	c.cookies = res.Cookies()
	return c
}

// do sends body as JSON (nil = no body) and returns the response; the body is closed at test end.
func (c *apiClient) do(method, path string, body any, headers map[string]string) *http.Response {
	c.t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(c.t, err)
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.base+path, r)
	require.NoError(c.t, err)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, ck := range c.cookies {
		req.AddCookie(ck)
	}
	res, err := http.DefaultClient.Do(req)
	require.NoError(c.t, err)
	c.t.Cleanup(func() { res.Body.Close() })
	return res
}

func decodeJSON[T any](t *testing.T, res *http.Response) T {
	t.Helper()
	var v T
	require.NoError(t, json.NewDecoder(res.Body).Decode(&v))
	return v
}

// newAdmin inserts an admin account (with a profile row) and returns its code.
func newAdmin(t *testing.T) string {
	t.Helper()
	userID, _, code := newEmployee(t, "0811111111")
	_, err := ownerPool.Exec(t.Context(), `UPDATE users SET role = 'admin' WHERE id = $1`, userID)
	require.NoError(t, err)
	return code
}
