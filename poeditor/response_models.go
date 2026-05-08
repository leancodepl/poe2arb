package poeditor

import "encoding/json"

type response struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type baseResponse struct {
	Response response `json:"response"`
}

type languagesListResponse struct {
	baseResponse
	Result struct {
		Languages []struct {
			Name string `json:"name"`
			Code string `json:"code"`
		} `json:"languages"`
	} `json:"result"`
}

type projectsExportResponse struct {
	baseResponse
	Result struct {
		URL string `json:"url"`
	} `json:"result"`
}

type termsListResponse struct {
	baseResponse
	Result struct {
		Terms []struct {
			Term        string          `json:"term"`
			Context     string          `json:"context"`
			Plural      string          `json:"plural"`
			Reference   string          `json:"reference"`
			Tags        []string        `json:"tags"`
			Comment     string          `json:"comment"`
			Translation termTranslation `json:"translation"`
		} `json:"terms"`
	} `json:"result"`
}

// termTranslation matches both the simple-string form and the plural object
// form returned by POEditor's /terms/list endpoint when a language is specified.
type termTranslation struct {
	Content string
}

func (t *termTranslation) UnmarshalJSON(data []byte) error {
	var raw struct {
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw.Content) == 0 {
		return nil
	}

	var asString string
	if err := json.Unmarshal(raw.Content, &asString); err == nil {
		t.Content = asString
		return nil
	}

	// Plural form: object with one/other/etc.
	var asPlural map[string]string
	if err := json.Unmarshal(raw.Content, &asPlural); err == nil {
		// Prefer "other" as the canonical single-line preview.
		if v, ok := asPlural["other"]; ok {
			t.Content = v
			return nil
		}
		for _, v := range asPlural {
			t.Content = v
			return nil
		}
	}

	return nil
}
