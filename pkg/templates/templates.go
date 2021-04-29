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

	tmplRedirect  *template.Template
	tmplRetryJoin *template.Template
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
