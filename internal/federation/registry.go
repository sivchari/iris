package federation

import (
	"github.com/vektah/gqlparser/v2/ast"
)

var providers []Provider

// Register adds a federation provider to the registry.
func Register(p Provider) {
	providers = append(providers, p)
}

// Detect checks the schema against all registered providers
// and returns federation info if detected.
func Detect(schema *ast.Schema) *Info {
	for _, p := range providers {
		if p.Detect(schema) {
			return &Info{
				Provider:   p,
				IsSubgraph: true,
				Entities:   extractEntities(schema, p),
			}
		}
	}

	return nil
}

// GetProviders returns all registered providers.
func GetProviders() []Provider {
	return providers
}

func extractEntities(schema *ast.Schema, _ Provider) []string {
	var entities []string

	for _, t := range schema.Types {
		if t.Kind != ast.Object {
			continue
		}

		for _, d := range t.Directives {
			if d.Name == "key" {
				entities = append(entities, t.Name)

				break
			}
		}
	}

	return entities
}
