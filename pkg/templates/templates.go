package templates

import (
	"bytes"
	"html/template"
	// Use go16 embedding instead of inline templates
	_ "embed"
)

var (
	//go:embed html/redirect.html
	tmplRedirectHTML string

	//go:embed html/retry-join.html
	tmplRetryJoinHTML string

	//go:embed html/default-presentation-body.xml
	tmplDefaultPresentationBodyXML string

	tmplRedirect                *template.Template
	tmplRetryJoin               *template.Template
	tmplDefaultPresentationBody *template.Template
)

// Redirect applies the redirect template
func Redirect(url string) []byte {
	if tmplRedirect == nil {
		tmplRedirect, _ = template.New("redirect").Parse(tmplRedirectHTML)
	}

	// Render template
	res := new(bytes.Buffer)
	tmplRedirect.Execute(res, url)
	return res.Bytes()
}

// RetryJoin applies the retry join template
func RetryJoin(url string) []byte {
	if tmplRedirect == nil {
		tmplRedirect, _ = template.New("retry_join").Parse(tmplRetryJoinHTML)
	}

	// Render template
	res := new(bytes.Buffer)
	tmplRedirect.Execute(res, url)
	return res.Bytes()
}

// DefaultPresentationBody renders the xml body for
// a default presentation.
func DefaultPresentationBody(url, filename string) []byte {
	if tmplDefaultPresentationBody == nil {
		tmplDefaultPresentationBody, _ = template.New("default_presentation").
			Parse(tmplDefaultPresentationBodyXML)
	}

	// Render template
	res := new(bytes.Buffer)
	tmplDefaultPresentationBody.Execute(res, map[string]string{
		"url":      url,
		"filename": filename,
	})
	return res.Bytes()
}
