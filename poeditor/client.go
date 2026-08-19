// Package poeditor handles communication with POEditor API.
//
// See https://poeditor.com/docs/api
package poeditor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiURL = "https://api.poeditor.com/v2"

type Client struct {
	apiURL string
	token  string

	client http.Client
}

const (
	// PaidAccountUploadRateLimit is from https://poeditor.com/docs/api_rates
	PaidAccountUploadRateLimit = 10 * time.Second
	// FreeAccountUploadRateLimit is from https://poeditor.com/docs/api_rates
	FreeAccountUploadRateLimit = 20 * time.Second
)

func NewClient(token string) *Client {
	return &Client{
		apiURL: apiURL,
		token:  token,
		client: http.Client{},
	}
}

func (c *Client) encodeBody(params map[string]string) io.Reader {
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}

	values.Set("api_token", c.token)

	return strings.NewReader(values.Encode())
}

func (c *Client) request(path string, params map[string]string, respBody interface{}) error {
	reqURL := fmt.Sprintf("%s%s", c.apiURL, path)
	body := c.encodeBody(params)

	req, err := http.NewRequest("POST", reqURL, body)
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("making HTTP request: %w", err)
	}

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	return nil
}

func handleRequestErr(err error, resp baseResponse) error {
	if err != nil {
		return err
	}

	return TryNewErrorFromResponse(resp.Response)
}

func boolField(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func (c *Client) AddLanguage(projectID, languageCode string) error {
	var resp baseResponse
	params := map[string]string{
		"id":       projectID,
		"language": languageCode,
	}
	err := c.request("/languages/add", params, &resp)
	if err := handleRequestErr(err, resp); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetProjectLanguages(projectID string) ([]Language, error) {
	var resp languagesListResponse

	params := map[string]string{"id": projectID}
	err := c.request("/languages/list", params, &resp)
	if err := handleRequestErr(err, resp.baseResponse); err != nil {
		return nil, err
	}

	langs := []Language{}
	for _, lang := range resp.Result.Languages {
		langs = append(langs, Language{
			Name: lang.Name,
			Code: lang.Code,
		})
	}

	return langs, nil
}

// ListTerms fetches all terms in the POEditor project. If languageCode is
// non-empty, each term's Translation field is populated with the translation
// in that language (useful for displaying a preview before deletion).
//
// See: https://poeditor.com/docs/api#terms_list
func (c *Client) ListTerms(projectID, languageCode string) ([]Term, error) {
	var resp termsListResponse

	params := map[string]string{"id": projectID}
	if languageCode != "" {
		params["language"] = languageCode
	}
	err := c.request("/terms/list", params, &resp)
	if err := handleRequestErr(err, resp.baseResponse); err != nil {
		return nil, err
	}

	terms := make([]Term, 0, len(resp.Result.Terms))
	for _, t := range resp.Result.Terms {
		terms = append(terms, Term{
			Term:           t.Term,
			Context:        t.Context,
			Plural:         t.Plural,
			Reference:      t.Reference,
			Tags:           t.Tags,
			Comment:        t.Comment,
			Translation:    t.Translation.Content,
			TranslationRaw: t.Translation.Raw,
		})
	}

	return terms, nil
}

// TermRef identifies a term by its name and (optional) context.
type TermRef struct {
	Term    string `json:"term"`
	Context string `json:"context,omitempty"`
}

// DeleteTerms permanently removes the given terms from the POEditor project.
//
// See: https://poeditor.com/docs/api#terms_delete
func (c *Client) DeleteTerms(projectID string, terms []TermRef) error {
	if len(terms) == 0 {
		return nil
	}

	data, err := json.Marshal(terms)
	if err != nil {
		return fmt.Errorf("encoding terms: %w", err)
	}

	var resp baseResponse
	params := map[string]string{
		"id":   projectID,
		"data": string(data),
	}
	err = c.request("/terms/delete", params, &resp)
	if err := handleRequestErr(err, resp); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetExportURL(projectID, languageCode string) (string, error) {
	var resp projectsExportResponse

	params := map[string]string{
		"id":       projectID,
		"language": languageCode,
		"type":     "json",
	}
	err := c.request("/projects/export", params, &resp)
	if err := handleRequestErr(err, resp.baseResponse); err != nil {
		return "", err
	}

	return resp.Result.URL, nil
}

// UploadOptions configures the Upload call.
type UploadOptions struct {
	// Overwrite, when true, replaces existing translations and term definitions
	// in POEditor with the ones in the uploaded file.
	Overwrite bool
}

func (c *Client) Upload(projectID, languageCode string, file io.Reader, opts UploadOptions) error {
	reqURL := fmt.Sprintf("%s%s", c.apiURL, "/projects/upload")

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	_ = w.WriteField("api_token", c.token)
	_ = w.WriteField("id", projectID)
	_ = w.WriteField("updating", "terms_translations")

	fileWriter, err := w.CreateFormFile("file", "file.json")
	if err != nil {
		return fmt.Errorf("creating form field: %w", err)
	}
	if _, err = io.Copy(fileWriter, file); err != nil {
		return fmt.Errorf("copying file to form field: %w", err)
	}

	_ = w.WriteField("language", languageCode)
	_ = w.WriteField("overwrite", boolField(opts.Overwrite))

	err = w.Close()
	if err != nil {
		return fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", reqURL, &b)
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("making HTTP request: %w", err)
	}

	var respBody baseResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	if err := handleRequestErr(err, respBody); err != nil {
		return err
	}

	return nil
}
