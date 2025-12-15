// Package federation provides GraphQL Federation support through a plugin architecture.
package federation

import (
	"context"

	"github.com/vektah/gqlparser/v2/ast"

	"github.com/sivchari/iris/internal/client"
)

// Provider defines the interface for Federation implementations.
type Provider interface {
	// Name returns the name of this federation implementation.
	Name() string

	// Detect checks if the schema uses this federation implementation.
	Detect(schema *ast.Schema) bool

	// GetServiceSDL retrieves the service SDL from the subgraph.
	// This typically calls the _service query.
	GetServiceSDL(ctx context.Context, c *client.Client) (string, error)

	// GetFederationDirectives returns the list of federation-specific directives.
	GetFederationDirectives() []string

	// FormatEntityInfo formats entity information for display.
	FormatEntityInfo(schema *ast.Schema) string
}

// Info contains detected federation information.
type Info struct {
	Provider   Provider
	IsSubgraph bool
	Entities   []string
}
