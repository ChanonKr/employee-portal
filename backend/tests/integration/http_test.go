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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/configs"
	"employee-portal/modules/server"
)

// End-to-end through the real router: login → GET /profile → PUT contact.
func TestHTTP_LoginViewAndEditProfile(t *testing.T) {
	_, _, code := newEmployee(t, "0811111111")

	srv, err := server.New(configs.Config{
		JWTSecret: "integration-test-secret-0123456789", JWTTTL: 3600e9, HistoryDays: 7,
		CORSOrigins: []string{"http://localhost:5173"},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)), appPool)
	require.NoError(t, err)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	do := func(method, path, body string, headers map[string]string, cookies []*http.Cookie) *http.Response {
		req, _ := http.NewRequest(method, ts.URL+path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		for _, c := range cookies {
			req.AddCookie(c)
		}
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { res.Body.Close() })
		return res
	}

	// Unauthenticated
	res := do("GET", "/api/v1/profile", "", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

	// Wrong password → generic 401
	res = do("POST", "/api/v1/auth/login", `{"employeeCode":"`+code+`","password":"wrong"}`, nil, nil)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

	// Login
	res = do("POST", "/api/v1/auth/login", `{"employeeCode":"`+code+`","password":"demo1234"}`, nil, nil)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
	cookies := res.Cookies()
	require.Len(t, cookies, 1)
	assert.True(t, cookies[0].HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookies[0].SameSite)

	// GET /profile
	res = do("GET", "/api/v1/profile", "", nil, cookies)
	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `"1"`, res.Header.Get("ETag"))
	var p struct {
		EmployeeCode string `json:"employeeCode"`
		Contact      struct {
			MobilePhone *string `json:"mobilePhone"`
		} `json:"contact"`
		Skills []any `json:"skills"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&p))
	assert.Equal(t, code, p.EmployeeCode)
	assert.Equal(t, "0811111111", *p.Contact.MobilePhone)
	assert.NotNil(t, p.Skills, "skills is [] not null")

	// PUT without If-Match → 428
	res = do("PUT", "/api/v1/profile/contact/mobile", `{"value":"0812345678"}`, nil, cookies)
	assert.Equal(t, http.StatusPreconditionRequired, res.StatusCode)

	// Unknown field → 400
	res = do("PUT", "/api/v1/profile/contact/first_name", `{"value":"X"}`, map[string]string{"If-Match": `"1"`}, cookies)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	// Invalid value → 400 with field detail
	res = do("PUT", "/api/v1/profile/contact/mobile", `{"value":"abc"}`, map[string]string{"If-Match": `"1"`}, cookies)
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
	var envelope struct {
		Error struct {
			Code      string            `json:"code"`
			Fields    map[string]string `json:"fields"`
			RequestID string            `json:"requestId"`
		} `json:"error"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&envelope))
	assert.Equal(t, "VALIDATION_ERROR", envelope.Error.Code)
	assert.Contains(t, envelope.Error.Fields, "mobile")
	assert.NotEmpty(t, envelope.Error.RequestID)

	// Valid update
	res = do("PUT", "/api/v1/profile/contact/mobile", `{"value":"0812345678"}`, map[string]string{"If-Match": `"1"`}, cookies)
	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `"2"`, res.Header.Get("ETag"))

	// Same stale version again → 409
	res = do("PUT", "/api/v1/profile/contact/mobile", `{"value":"0899999999"}`, map[string]string{"If-Match": `"1"`}, cookies)
	assert.Equal(t, http.StatusConflict, res.StatusCode)

	// No route can change the name
	res = do("PUT", "/api/v1/profile/name", `{"firstName":"X"}`, map[string]string{"If-Match": `"2"`}, cookies)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
