package poeditor

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	client := NewClient("test-token")
	client.apiURL = server.URL
	return client, server
}

func decodeForm(r *http.Request) (url.Values, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return url.ParseQuery(string(body))
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, body)
}

func TestListTerms(t *testing.T) {
	t.Run("returns terms with translations on success", func(t *testing.T) {
		client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/terms/list", r.URL.Path)
			form, err := decodeForm(r)
			require.NoError(t, err)
			assert.Equal(t, "test-token", form.Get("api_token"))
			assert.Equal(t, "proj-1", form.Get("id"))
			assert.Equal(t, "en", form.Get("language"))

			writeJSON(w, `{
				"response":{"status":"success","code":"200","message":"OK"},
				"result":{"terms":[
					{"term":"appTitle","context":"","plural":"","tags":["v1"],"translation":{"content":"My App"}},
					{"term":"items","context":"","plural":".","translation":{"content":{"one":"1 item","other":"{count} items"}}},
					{"term":"untranslated","translation":{"content":""}}
				]}
			}`)
		})
		defer server.Close()

		terms, err := client.ListTerms("proj-1", "en")
		require.NoError(t, err)

		require.Len(t, terms, 3)
		assert.Equal(t, "appTitle", terms[0].Term)
		assert.Equal(t, []string{"v1"}, terms[0].Tags)
		assert.Equal(t, "My App", terms[0].Translation)
		assert.Equal(t, "items", terms[1].Term)
		assert.Equal(t, ".", terms[1].Plural)
		assert.Equal(t, "{count} items", terms[1].Translation)
		assert.Equal(t, "untranslated", terms[2].Term)
		assert.Empty(t, terms[2].Translation)
	})

	t.Run("omits language param when empty", func(t *testing.T) {
		client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
			form, err := decodeForm(r)
			require.NoError(t, err)
			assert.Empty(t, form.Get("language"))

			writeJSON(w, `{"response":{"status":"success","code":"200","message":"OK"},"result":{"terms":[]}}`)
		})
		defer server.Close()

		_, err := client.ListTerms("proj-1", "")
		require.NoError(t, err)
	})

	t.Run("propagates POEditor error", func(t *testing.T) {
		client, server := newTestClient(func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, `{"response":{"status":"fail","code":"403","message":"forbidden"}}`)
		})
		defer server.Close()

		_, err := client.ListTerms("proj-1", "")
		require.Error(t, err)

		var poeErr *Error
		require.True(t, errors.As(err, &poeErr))
		assert.Equal(t, 403, poeErr.Code)
	})
}

func TestDeleteTerms(t *testing.T) {
	t.Run("posts JSON-encoded terms", func(t *testing.T) {
		var capturedData string
		client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/terms/delete", r.URL.Path)
			form, err := decodeForm(r)
			require.NoError(t, err)

			assert.Equal(t, "proj-2", form.Get("id"))
			capturedData = form.Get("data")

			writeJSON(w, `{"response":{"status":"success","code":"200","message":"OK"}}`)
		})
		defer server.Close()

		err := client.DeleteTerms("proj-2", []TermRef{
			{Term: "obsoleteTerm"},
			{Term: "anotherTerm", Context: "ctx"},
		})
		require.NoError(t, err)

		assert.JSONEq(t,
			`[{"term":"obsoleteTerm"},{"term":"anotherTerm","context":"ctx"}]`,
			capturedData,
		)
	})

	t.Run("no-op for empty input", func(t *testing.T) {
		called := false
		client, server := newTestClient(func(http.ResponseWriter, *http.Request) {
			called = true
		})
		defer server.Close()

		err := client.DeleteTerms("proj-2", nil)
		require.NoError(t, err)
		assert.False(t, called, "request should not be made for empty term list")
	})
}

func TestUpload(t *testing.T) {
	parseMultipart := func(t *testing.T, r *http.Request) (map[string]string, []byte) {
		t.Helper()

		_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		require.NoError(t, err)

		mr := multipart.NewReader(r.Body, params["boundary"])
		fields := map[string]string{}
		var fileBytes []byte

		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)

			body, err := io.ReadAll(part)
			require.NoError(t, err)

			if part.FileName() != "" {
				fileBytes = body
			} else {
				fields[part.FormName()] = string(body)
			}
		}

		return fields, fileBytes
	}

	t.Run("default options send overwrite=0", func(t *testing.T) {
		client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/projects/upload", r.URL.Path)
			fields, file := parseMultipart(t, r)
			assert.Equal(t, "test-token", fields["api_token"])
			assert.Equal(t, "proj-1", fields["id"])
			assert.Equal(t, "en", fields["language"])
			assert.Equal(t, "terms_translations", fields["updating"])
			assert.Equal(t, "0", fields["overwrite"])
			assert.Equal(t, `[{"term":"x"}]`, string(file))

			writeJSON(w, `{"response":{"status":"success","code":"200","message":"OK"}}`)
		})
		defer server.Close()

		err := client.Upload("proj-1", "en", strings.NewReader(`[{"term":"x"}]`), UploadOptions{})
		require.NoError(t, err)
	})

	t.Run("Overwrite=true sends overwrite=1", func(t *testing.T) {
		client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
			fields, _ := parseMultipart(t, r)
			assert.Equal(t, "1", fields["overwrite"])

			writeJSON(w, `{"response":{"status":"success","code":"200","message":"OK"}}`)
		})
		defer server.Close()

		err := client.Upload("p", "en", strings.NewReader(`[]`), UploadOptions{Overwrite: true})
		require.NoError(t, err)
	})

	t.Run("propagates rate limit error", func(t *testing.T) {
		client, server := newTestClient(func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, `{"response":{"status":"fail","code":"4048","message":"too fast"}}`)
		})
		defer server.Close()

		err := client.Upload("p", "en", strings.NewReader(`[]`), UploadOptions{})
		require.Error(t, err)

		var poeErr *Error
		require.True(t, errors.As(err, &poeErr))
		assert.Equal(t, RateLimitErrorCode, poeErr.Code)
	})
}

func TestGetProjectLanguages(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/languages/list", r.URL.Path)
		writeJSON(w, `{
			"response":{"status":"success","code":"200","message":"OK"},
			"result":{"languages":[
				{"name":"English","code":"en"},
				{"name":"Polish","code":"pl"}
			]}
		}`)
	})
	defer server.Close()

	langs, err := client.GetProjectLanguages("p")
	require.NoError(t, err)
	assert.Equal(t,
		[]Language{{Name: "English", Code: "en"}, {Name: "Polish", Code: "pl"}},
		langs,
	)
}

func TestAddLanguage(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/languages/add", r.URL.Path)
		form, err := decodeForm(r)
		require.NoError(t, err)
		assert.Equal(t, "p", form.Get("id"))
		assert.Equal(t, "es", form.Get("language"))

		writeJSON(w, `{"response":{"status":"success","code":"200","message":"OK"}}`)
	})
	defer server.Close()

	require.NoError(t, client.AddLanguage("p", "es"))
}
