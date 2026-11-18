//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// End-to-end through the real router: login → GET /profile → PUT contact.
func TestHTTP_LoginViewAndEditProfile(t *testing.T) {
	_, _, code := newEmployee(t, "0811111111")

	baseURL := newTestServer(t)

	do := func(method, path, body string, headers map[string]string, cookies []*http.Cookie) *http.Response {
		req, _ := http.NewRequest(method, baseURL+path, bytes.NewBufferString(body))
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

	// Skill catalog: active seed skills in display order
	res = do("GET", "/api/v1/skills", "", nil, cookies)
	require.Equal(t, http.StatusOK, res.StatusCode)
	var catalog []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&catalog))
	require.GreaterOrEqual(t, len(catalog), 2)
	assert.Equal(t, "Go", catalog[0].Name)

	// Missing skillIds → 400; valid set → new version
	res = do("PUT", "/api/v1/profile/skills", `{}`, map[string]string{"If-Match": `"2"`}, cookies)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	body := `{"skillIds":[` + strconv.FormatInt(catalog[0].ID, 10) + `,` + strconv.FormatInt(catalog[1].ID, 10) + `]}`
	res = do("PUT", "/api/v1/profile/skills", body, map[string]string{"If-Match": `"2"`}, cookies)
	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `"3"`, res.Header.Get("ETag"))

	res = do("GET", "/api/v1/profile", "", nil, cookies)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&p))
	assert.Len(t, p.Skills, 2)

	// History shows both edits (mobile, skills), newest first, in the documented shape
	res = do("GET", "/api/v1/profile/history", "", nil, cookies)
	require.Equal(t, http.StatusOK, res.StatusCode)
	var hist struct {
		Days     int  `json:"days"`
		HasOlder bool `json:"hasOlder"`
		Items    []struct {
			Field     string          `json:"field"`
			OldValue  json.RawMessage `json:"oldValue"`
			NewValue  json.RawMessage `json:"newValue"`
			ChangedAt string          `json:"changedAt"`
			ChangedBy struct {
				ID   int64  `json:"id"`
				Role string `json:"role"`
			} `json:"changedBy"`
		} `json:"items"`
		Page struct {
			Page     int  `json:"page"`
			PageSize int  `json:"pageSize"`
			HasNext  bool `json:"hasNext"`
		} `json:"page"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&hist))
	assert.Equal(t, 7, hist.Days)
	assert.False(t, hist.HasOlder)
	require.Len(t, hist.Items, 2)
	assert.Equal(t, "skills", hist.Items[0].Field)
	assert.JSONEq(t, `[]`, string(hist.Items[0].OldValue))
	assert.Equal(t, "mobile_phone", hist.Items[1].Field)
	assert.JSONEq(t, `"0811111111"`, string(hist.Items[1].OldValue))
	assert.Equal(t, "employee", hist.Items[1].ChangedBy.Role)
	assert.Equal(t, 20, hist.Page.PageSize)

	res = do("GET", "/api/v1/profile/history?field=mobile", "", nil, cookies)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode, "filter uses the stored field name")
	res = do("GET", "/api/v1/profile/history?page=abc", "", nil, cookies)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	// ── photo upload (multipart)
	upload := func(content []byte, version string) *http.Response {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		part, _ := mw.CreateFormFile("file", "me.jpg")
		_, _ = part.Write(content)
		_ = mw.Close()
		req, _ := http.NewRequest("PUT", baseURL+"/api/v1/profile/photo", &body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		req.Header.Set("If-Match", version)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { res.Body.Close() })
		return res
	}

	res = upload([]byte("plain text renamed to .jpg"), `"3"`)
	assert.Equal(t, http.StatusUnsupportedMediaType, res.StatusCode)

	res = upload(bytes.Repeat([]byte{0xFF}, 6<<20), `"3"`)
	assert.Equal(t, http.StatusRequestEntityTooLarge, res.StatusCode)

	res = upload(testPNG(t, 800, 800), `"3"`)
	require.Equal(t, http.StatusOK, res.StatusCode)
	var photo struct {
		Version  int32  `json:"version"`
		PhotoURL string `json:"photoUrl"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&photo))
	assert.EqualValues(t, 4, photo.Version)
	assert.Regexp(t, `^/api/v1/files/photos/[0-9a-f-]{36}\.jpg$`, photo.PhotoURL)

	res = do("GET", photo.PhotoURL, "", nil, cookies)
	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "image/jpeg", res.Header.Get("Content-Type"))
	assert.Contains(t, res.Header.Get("Cache-Control"), "immutable")

	res = do("GET", photo.PhotoURL, "", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "photos need a session")

	// The photo history carries URLs (the stored values are file names).
	res = do("GET", "/api/v1/profile/history?field=photo", "", nil, cookies)
	require.Equal(t, http.StatusOK, res.StatusCode)
	var photoHist struct {
		Items []struct {
			OldValue *string `json:"oldValue"`
			NewValue string  `json:"newValue"`
		} `json:"items"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&photoHist))
	require.Len(t, photoHist.Items, 1)
	assert.Nil(t, photoHist.Items[0].OldValue, "first photo")
	assert.Equal(t, photo.PhotoURL, photoHist.Items[0].NewValue)
	res = do("GET", "/api/v1/files/photos/..%2F..%2Fetc%2Fpasswd", "", nil, cookies)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	res = do("GET", "/api/v1/profile", "", nil, cookies)
	var withPhoto struct {
		PhotoURL *string `json:"photoUrl"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&withPhoto))
	require.NotNil(t, withPhoto.PhotoURL)
	assert.Equal(t, photo.PhotoURL, *withPhoto.PhotoURL)

	// JSON endpoints keep the small body limit even though uploads allow 5 MB.
	big := `{"value":"` + strings.Repeat("a", 100<<10) + `"}`
	res = do("PUT", "/api/v1/profile/contact/address", big, map[string]string{"If-Match": `"4"`}, cookies)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
