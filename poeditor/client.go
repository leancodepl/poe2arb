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

	"github.com/pkg/errors"
)

const apiURL = "https://api.poeditor.com/v2"

type Client struct {
	apiURL string
	token  string

	client http.Client
}

func NewClient(token string) *Client {
	return &Client{
		apiURL: apiURL,
		token:  token,
		client: http.Client{},
	}
}

func (c Client) encodeBody(params map[string]string) io.Reader {
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}

	values.Set("api_token", c.token)

	return strings.NewReader(values.Encode())
}

func (c *Client) request(path string, params map[string]string, respBody interface{}) error {
	url := fmt.Sprintf("%s%s", c.apiURL, path)
	body := c.encodeBody(params)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return errors.Wrap(err, "creating HTTP request")
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "making HTTP request")
	}

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return errors.Wrap(err, "decoding response")
	}

	return nil
}

func handleRequestErr(err error, resp baseResponse) error {
	if err != nil {
		return err
	}

	return TryNewErrorFromResponse(resp.Response)
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
			Code: NewLocale(lang.Code),
		})
	}

	return langs, nil
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

func (c *Client) Upload(projectID, languageCode string, file io.Reader) error {
	url := fmt.Sprintf("%s%s", c.apiURL, "/projects/upload")

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.WriteField("api_token", c.token)
	w.WriteField("id", projectID)
	w.WriteField("updating", "terms_translations")

	fileWriter, err := w.CreateFormFile("file", "file.json")
	if err != nil {
		return errors.Wrap(err, "creating form field")
	}
	if _, err = io.Copy(fileWriter, file); err != nil {
		return errors.Wrap(err, "copying file to form field")
	}

	w.WriteField("language", languageCode)
	w.WriteField("overwrite", "0")

	err = w.Close()
	if err != nil {
		return errors.Wrap(err, "closing multipart writer")
	}

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return errors.Wrap(err, "creating HTTP request")
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "making HTTP request")
	}

	var respBody baseResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return errors.Wrap(err, "decoding response")
	}

	if err := handleRequestErr(err, respBody); err != nil {
		return err
	}

	return nil
}
