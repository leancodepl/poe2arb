package poeditor

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
