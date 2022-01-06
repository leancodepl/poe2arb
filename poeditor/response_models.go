package poeditor

type response struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type languagesListResponse struct {
	Response response `json:"response"`
	Result   struct {
		Languages []struct {
			Name string `json:"name"`
			Code string `json:"code"`
		} `json:"languages"`
	} `json:"result"`
}

type projectsExportResponse struct {
	Response response `json:"response"`
	Result   struct {
		URL string `json:"url"`
	} `json:"result"`
}
