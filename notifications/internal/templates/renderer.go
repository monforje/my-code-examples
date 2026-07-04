// Package templates
package templates

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"
)

//go:embed html/*.html text/*.txt
var templateFS embed.FS

type CodeTemplate string

const (
	CodeVerification   CodeTemplate = "code_verification"
	CodePasswordReset  CodeTemplate = "code_password_reset"
	CodePasswordChange CodeTemplate = "code_password_change"
	CodeEmailChange    CodeTemplate = "code_email_change"
	CodeDeleteAccount  CodeTemplate = "code_delete_account"
)

var codeTemplates = []CodeTemplate{
	CodeVerification,
	CodePasswordReset,
	CodePasswordChange,
	CodeEmailChange,
	CodeDeleteAccount,
}

var subjects = map[CodeTemplate]string{
	CodeVerification:   "Код подтверждения Codurity",
	CodePasswordReset:  "Восстановление пароля Codurity",
	CodePasswordChange: "Подтверждение смены пароля",
	CodeEmailChange:    "Подтверждение смены почты",
	CodeDeleteAccount:  "Подтверждение удаления аккаунта",
}

type CodeEmailData struct {
	Email          string
	Code           string
	PrivacyURL     string
	CompanyAddress string
}

type RenderedEmail struct {
	Subject string
	HTML    string
	Text    string
}

type Renderer struct {
	htmlTemplates map[CodeTemplate]*htmltemplate.Template
	textTemplates map[CodeTemplate]*texttemplate.Template
}

func NewRenderer() (*Renderer, error) {
	r := &Renderer{
		htmlTemplates: make(map[CodeTemplate]*htmltemplate.Template, len(codeTemplates)),
		textTemplates: make(map[CodeTemplate]*texttemplate.Template, len(codeTemplates)),
	}

	for _, name := range codeTemplates {
		htmlPath := fmt.Sprintf("html/%s.html", name)
		textPath := fmt.Sprintf("text/%s.txt", name)

		htmlTpl, err := htmltemplate.ParseFS(templateFS, htmlPath)
		if err != nil {
			return nil, fmt.Errorf("parse html template %q: %w", htmlPath, err)
		}

		textTpl, err := texttemplate.ParseFS(templateFS, textPath)
		if err != nil {
			return nil, fmt.Errorf("parse text template %q: %w", textPath, err)
		}

		r.htmlTemplates[name] = htmlTpl
		r.textTemplates[name] = textTpl
	}

	return r, nil
}

func (r *Renderer) RenderCodeEmail(name CodeTemplate, data CodeEmailData) (RenderedEmail, error) {
	subject, ok := subjects[name]
	if !ok {
		return RenderedEmail{}, fmt.Errorf("unknown email subject for template %q", name)
	}

	htmlTpl, ok := r.htmlTemplates[name]
	if !ok {
		return RenderedEmail{}, fmt.Errorf("html template %q not found", name)
	}

	textTpl, ok := r.textTemplates[name]
	if !ok {
		return RenderedEmail{}, fmt.Errorf("text template %q not found", name)
	}

	var htmlBuf bytes.Buffer
	if err := htmlTpl.Execute(&htmlBuf, data); err != nil {
		return RenderedEmail{}, fmt.Errorf("execute html template %q: %w", name, err)
	}

	var textBuf bytes.Buffer
	if err := textTpl.Execute(&textBuf, data); err != nil {
		return RenderedEmail{}, fmt.Errorf("execute text template %q: %w", name, err)
	}

	return RenderedEmail{
		Subject: subject,
		HTML:    htmlBuf.String(),
		Text:    textBuf.String(),
	}, nil
}
