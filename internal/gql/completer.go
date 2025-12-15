// Package gql provides GraphQL utilities including schema loading and completion.
package gql

import (
	"strings"

	prompt "github.com/ktr0731/go-prompt"
	"github.com/vektah/gqlparser/v2/ast"
)

// Completer provides GraphQL-aware completion.
type Completer struct {
	schema *ast.Schema
}

// NewCompleter creates a new Completer.
func NewCompleter(schema *ast.Schema) *Completer {
	return &Completer{schema: schema}
}

// Complete returns suggestions based on the input.
func (c *Completer) Complete(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	if text == "" {
		return c.commands()
	}

	// GraphQL query completion
	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "{") ||
		strings.HasPrefix(trimmed, "query") ||
		strings.HasPrefix(trimmed, "mutation") {
		return c.completeGraphQL(d.GetWordBeforeCursor())
	}

	// Command completion
	words := strings.Fields(text)
	if len(words) == 0 {
		return c.commands()
	}

	cmd := words[0]

	// First word: command
	if len(words) == 1 && !strings.HasSuffix(text, " ") {
		return prompt.FilterHasPrefix(c.commands(), cmd, true)
	}

	// Command arguments
	prefix := d.GetWordBeforeCursor()

	switch cmd {
	case "show":
		return c.completeShow(prefix)
	case "desc", "describe":
		return c.completeTypes(prefix)
	case "call":
		return c.completeCall(prefix)
	}

	return nil
}

func (c *Completer) commands() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "help", Description: "Show help"},
		{Text: "show", Description: "Show schema info"},
		{Text: "desc", Description: "Describe type/field"},
		{Text: "call", Description: "Call query/mutation"},
		{Text: "exit", Description: "Exit"},
	}
}

func (c *Completer) completeShow(prefix string) []prompt.Suggest {
	suggests := []prompt.Suggest{
		{Text: "types", Description: "List types"},
		{Text: "queries", Description: "List queries"},
		{Text: "mutations", Description: "List mutations"},
		{Text: "federation", Description: "Show federation info"},
	}
	if prefix == "" {
		return suggests
	}

	return prompt.FilterHasPrefix(suggests, prefix, true)
}

func (c *Completer) completeTypes(prefix string) []prompt.Suggest {
	suggests := make([]prompt.Suggest, 0, len(c.schema.Types))

	for _, t := range c.schema.Types {
		if strings.HasPrefix(t.Name, "__") {
			continue
		}

		suggests = append(suggests, prompt.Suggest{
			Text:        t.Name,
			Description: string(t.Kind),
		})
	}

	if prefix == "" {
		return suggests
	}

	return prompt.FilterHasPrefix(suggests, prefix, true)
}

func (c *Completer) completeCall(prefix string) []prompt.Suggest {
	var suggests []prompt.Suggest

	// Queries
	if c.schema.Query != nil {
		for _, f := range c.schema.Query.Fields {
			if strings.HasPrefix(f.Name, "__") {
				continue
			}

			suggests = append(suggests, prompt.Suggest{
				Text:        f.Name,
				Description: "query",
			})
		}
	}

	// Mutations
	if c.schema.Mutation != nil {
		for _, f := range c.schema.Mutation.Fields {
			if strings.HasPrefix(f.Name, "__") {
				continue
			}

			suggests = append(suggests, prompt.Suggest{
				Text:        f.Name,
				Description: "mutation",
			})
		}
	}

	if prefix == "" {
		return suggests
	}

	return prompt.FilterHasPrefix(suggests, prefix, true)
}

func (c *Completer) completeGraphQL(prefix string) []prompt.Suggest {
	// Suggest query root fields for now
	var suggests []prompt.Suggest

	if c.schema.Query != nil {
		for _, f := range c.schema.Query.Fields {
			if strings.HasPrefix(f.Name, "__") {
				continue
			}

			suggests = append(suggests, prompt.Suggest{
				Text:        f.Name,
				Description: FormatType(f.Type),
			})
		}
	}

	if prefix == "" {
		return suggests
	}

	return prompt.FilterHasPrefix(suggests, prefix, true)
}
