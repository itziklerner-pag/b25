package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"
	"text/template" as textTemplate "text/template"

	"github.com/b25/services/notification/internal/models"
)

// TemplateEngine handles template rendering
type TemplateEngine struct {
	htmlTemplates map[string]*template.Template
	textTemplates map[string]*textTemplate.Template
	mu            sync.RWMutex
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		htmlTemplates: make(map[string]*template.Template),
		textTemplates: make(map[string]*textTemplate.Template),
	}
}

// RegisterHTMLTemplate registers an HTML template
func (te *TemplateEngine) RegisterHTMLTemplate(name, content string) error {
	te.mu.Lock()
	defer te.mu.Unlock()

	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template %s: %w", name, err)
	}

	te.htmlTemplates[name] = tmpl
	return nil
}

// RegisterTextTemplate registers a text template
func (te *TemplateEngine) RegisterTextTemplate(name, content string) error {
	te.mu.Lock()
	defer te.mu.Unlock()

	tmpl, err := textTemplate.New(name).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse text template %s: %w", name, err)
	}

	te.textTemplates[name] = tmpl
	return nil
}

// RegisterTemplate registers a template from a NotificationTemplate model
func (te *TemplateEngine) RegisterTemplate(tmpl *models.NotificationTemplate) error {
	// For email, we assume HTML if the template starts with <
	isHTML := tmpl.Channel == models.ChannelEmail && len(tmpl.BodyTemplate) > 0 && tmpl.BodyTemplate[0] == '<'

	if isHTML {
		return te.RegisterHTMLTemplate(tmpl.Name, tmpl.BodyTemplate)
	}
	return te.RegisterTextTemplate(tmpl.Name, tmpl.BodyTemplate)
}

// RenderHTML renders an HTML template with the given data
func (te *TemplateEngine) RenderHTML(name string, data interface{}) (string, error) {
	te.mu.RLock()
	defer te.mu.RUnlock()

	tmpl, exists := te.htmlTemplates[name]
	if !exists {
		return "", fmt.Errorf("HTML template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute HTML template %s: %w", name, err)
	}

	return buf.String(), nil
}

// RenderText renders a text template with the given data
func (te *TemplateEngine) RenderText(name string, data interface{}) (string, error) {
	te.mu.RLock()
	defer te.mu.RUnlock()

	tmpl, exists := te.textTemplates[name]
	if !exists {
		return "", fmt.Errorf("text template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute text template %s: %w", name, err)
	}

	return buf.String(), nil
}

// Render renders a template (auto-detects HTML vs text)
func (te *TemplateEngine) Render(name string, data interface{}) (string, error) {
	te.mu.RLock()
	defer te.mu.RUnlock()

	// Try HTML first
	if tmpl, exists := te.htmlTemplates[name]; exists {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("failed to execute HTML template %s: %w", name, err)
		}
		return buf.String(), nil
	}

	// Try text template
	if tmpl, exists := te.textTemplates[name]; exists {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("failed to execute text template %s: %w", name, err)
		}
		return buf.String(), nil
	}

	return "", fmt.Errorf("template %s not found", name)
}

// HasTemplate checks if a template is registered
func (te *TemplateEngine) HasTemplate(name string) bool {
	te.mu.RLock()
	defer te.mu.RUnlock()

	_, htmlExists := te.htmlTemplates[name]
	_, textExists := te.textTemplates[name]

	return htmlExists || textExists
}

// ClearTemplates clears all registered templates
func (te *TemplateEngine) ClearTemplates() {
	te.mu.Lock()
	defer te.mu.Unlock()

	te.htmlTemplates = make(map[string]*template.Template)
	te.textTemplates = make(map[string]*textTemplate.Template)
}

// GetTemplateCount returns the number of registered templates
func (te *TemplateEngine) GetTemplateCount() int {
	te.mu.RLock()
	defer te.mu.RUnlock()

	return len(te.htmlTemplates) + len(te.textTemplates)
}
