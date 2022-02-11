// Package poeditor handles communication with POEditor API.
//
// See https://poeditor.com/docs/api
package poeditor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetProjectLanguages(projectID string) ([]Language, error) {
	var resp languagesListResponse

	params := map[string]string{"id": projectID}
	err := c.request("/languages/list", params, &resp)
	if err != nil {
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

func (c *Client) GetExportURL(projectID, languageCode string) (string, error) {
	var resp projectsExportResponse

	params := map[string]string{
		"id":       projectID,
		"language": languageCode,
		"type":     "json",
	}
	err := c.request("/projects/export", params, &resp)
	if err != nil {
		return "", err
	}

	return resp.Result.URL, nil
}
