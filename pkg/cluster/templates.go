package cluster

import (
	"html/template"
)

func compileTmplRedirect() *template.Template {
	tmpl := `
	<html>
	  <head>
		<meta http-equiv="Refresh" content="0; url={{.}}" />
	  </head>
	  <body>
		<h1>You are beging redirected to the meeting</h1>
		<p>The meeting can be found at
			<a href="{{.}}">{{.}}</a>.
		</p>
	  </body>
	</html>
	`
	t, _ := template.New("redirect").Parse(tmpl)
	return t
}
