// Package apollo provides Apollo Federation support.
package apollo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"

	"github.com/sivchari/iris/internal/client"
	"github.com/sivchari/iris/internal/federation"
)

func init() {
	federation.Register(&Provider{})
}

// Provider implements Apollo Federation support.
type Provider struct{}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "Apollo Federation"
}

// Detect checks if the schema uses Apollo Federation.
func (p *Provider) Detect(schema *ast.Schema) bool {
	// Check for Federation directives
	for _, d := range schema.Directives {
		if isFederationDirective(d.Name) {
			return true
		}
	}

	// Check for _service query
	if schema.Query != nil {
		for _, f := range schema.Query.Fields {
			if f.Name == "_service" {
				return true
			}
		}
	}

	return false
}

// GetServiceSDL retrieves the SDL from _service query.
func (p *Provider) GetServiceSDL(ctx context.Context, c *client.Client) (string, error) {
	query := `query { _service { sdl } }`

	resp, err := c.Execute(ctx, &client.Request{Query: query})
	if err != nil {
		return "", fmt.Errorf("execute _service query: %w", err)
	}

	if len(resp.Errors) > 0 {
		return "", fmt.Errorf("_service query error: %s", resp.Errors[0].Message)
	}

	var result struct {
		Service struct {
			SDL string `json:"sdl"`
		} `json:"_service"` //nolint:tagliatelle // GraphQL field name is _service
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	return result.Service.SDL, nil
}

// GetFederationDirectives returns Apollo Federation directives.
func (p *Provider) GetFederationDirectives() []string {
	return []string{
		"key",
		"external",
		"requires",
		"provides",
		"extends",
		"shareable",
		"inaccessible",
		"override",
		"tag",
	}
}

// FormatEntityInfo formats entity information for display.
func (p *Provider) FormatEntityInfo(schema *ast.Schema) string {
	var sb strings.Builder

	sb.WriteString("Entities:\n")

	for _, t := range schema.Types {
		if t.Kind != ast.Object {
			continue
		}

		var keys []string

		for _, d := range t.Directives {
			if d.Name == "key" {
				for _, arg := range d.Arguments {
					if arg.Name == "fields" {
						keys = append(keys, arg.Value.Raw)
					}
				}
			}
		}

		if len(keys) > 0 {
			sb.WriteString(fmt.Sprintf("  %s @key(fields: %s)\n", t.Name, strings.Join(keys, ", ")))
		}
	}

	return sb.String()
}

func isFederationDirective(name string) bool {
	switch name {
	case "key", "external", "requires", "provides", "extends",
		"shareable", "inaccessible", "override", "tag":
		return true
	}

	return false
}
