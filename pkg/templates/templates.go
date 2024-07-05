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

	//go:embed html/meeting-not-found.html
	tmplMeetingNotFoundHTML string

	//go:embed html/attendees-limit-reached.html
	tmplAttendeesLimitReachedHTML string

	//go:embed html/error-page.html
	tmplErrorPageHTML string

	//go:embed xml/default-presentation-body.xml
	tmplDefaultPresentationBodyXML string

	tmplRedirect                *template.Template
	tmplRetryJoin               *template.Template
	tmplMeetingNotFound         *template.Template
	tmplAttendeesLimitReached   *template.Template
	tmplErrorPage               *template.Template
	tmplDefaultPresentationBody *template.Template
)

// Initialize templates
func init() {
	tmplRedirect, _ = template.New("redirect").Parse(tmplRedirectHTML)
	tmplRetryJoin, _ = template.New("retry_join").Parse(tmplRetryJoinHTML)
	tmplMeetingNotFound, _ = template.New("meeting_not_found").
		Parse(tmplMeetingNotFoundHTML)
	tmplAttendeesLimitReached, _ = template.New("attendees_limit_reached").
		Parse(tmplAttendeesLimitReachedHTML)
	tmplErrorPage, _ = template.New("error_page").Parse(tmplErrorPageHTML)
	tmplDefaultPresentationBody, _ = template.New("default_presentation").
		Parse(tmplDefaultPresentationBodyXML)
}

// Redirect applies the redirect template
func Redirect(url string) []byte {
	res := new(bytes.Buffer)
	tmplRedirect.Execute(res, url)
	return res.Bytes()
}

// RetryJoin applies the retry join template
func RetryJoin(url string) []byte {
	res := new(bytes.Buffer)
	tmplRetryJoin.Execute(res, url)
	return res.Bytes()
}

// MeetingNotFound applies the meeting not found template
func MeetingNotFound() []byte {
	res := new(bytes.Buffer)
	tmplMeetingNotFound.Execute(res, nil)
	return res.Bytes()
}

// AttendeesLimitReached applies the attendees limit reached template
func AttendeesLimitReached() []byte {
	res := new(bytes.Buffer)
	tmplAttendeesLimitReached.Execute(res, nil)
	return res.Bytes()
}

// ErrorPage renders the error page template
func ErrorPage(title, message string) []byte {
	res := new(bytes.Buffer)
	tmplErrorPage.Execute(res, map[string]interface{}{
		"Message": message,
		"Title":   title,
	})
	return res.Bytes()
}

// DefaultPresentationBody renders the xml body for
// a default presentation.
func DefaultPresentationBody(u, filename string) []byte {
	res := new(bytes.Buffer)
	tmplDefaultPresentationBody.Execute(res, struct{ URL, Filename string }{
		URL:      u,
		Filename: filename,
	})
	return res.Bytes()
}
