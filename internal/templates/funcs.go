package templates

import (
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"
)

var ManiplacerFuncs = template.FuncMap{
	"Base64": func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	},
	"ToUpper": func(s string) string {
		return strings.ToUpper(s)
	},
	"ToLower": func(s string) string {
		return strings.ToLower(s)
	},
	"Quote": func(s string) string {
		return fmt.Sprintf("%q", s) // adds quotes around string
	},
}
