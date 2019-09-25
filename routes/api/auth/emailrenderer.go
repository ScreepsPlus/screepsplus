package auth

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	"github.com/volatiletech/authboss"
)

// EmailRenderer renderer
type EmailRenderer struct {
	templates map[string]*template.Template
}

// NewEmailRenderer renderer
func NewEmailRenderer() *EmailRenderer {
	return &EmailRenderer{
		templates: make(map[string]*template.Template, 0),
	}
}

// Load templates
func (e *EmailRenderer) Load(names ...string) error {
	box := packr.New("templates", "../../../templates")
	for _, n := range names {
		tmpl, _ := box.FindString(fmt.Sprintf("email/%s.tmpl", n))
		temp, err := template.New("authboss_email").Parse(tmpl)
		if err != nil {
			return errors.Wrapf(err, "failed to load template for page %s", n)
		}

		e.templates[n] = temp
	}
	return nil
}

// Render a view
func (e *EmailRenderer) Render(ctx context.Context, page string, data authboss.HTMLData) (output []byte, contentType string, err error) {
	buf := &bytes.Buffer{}

	exe, ok := e.templates[page]
	contentType = "text/html"
	if strings.HasSuffix(page, "_txt") {
		contentType = "text/plain"
	}

	if !ok {
		return nil, "", errors.Errorf("template for page %s not found", page)
	}

	err = exe.Execute(buf, data)
	if err != nil {
		return nil, "", errors.Wrapf(err, "failed to render template for page %s", page)
	}

	return buf.Bytes(), contentType, nil
}
