package poeditor

type Language struct {
	Name string
	Code string
}

type Term struct {
	Term        string
	Context     string
	Plural      string
	Reference   string
	Tags        []string
	Comment     string
	Translation string
}
