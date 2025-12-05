package gql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/sivchari/iris/internal/client"
)

// LoadSchemaFromIntrospection loads a GraphQL schema from introspection.
func LoadSchemaFromIntrospection(ctx context.Context, c *client.Client) (*ast.Schema, error) {
	// Execute introspection query
	req := &client.Request{
		Query: introspectionQuery,
	}

	resp, err := c.Execute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute introspection: %w", err)
	}

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("introspection error: %s", resp.Errors[0].Message)
	}

	// Parse introspection response
	var introspection introspectionResponse
	if err := json.Unmarshal(resp.Data, &introspection); err != nil {
		return nil, fmt.Errorf("failed to parse introspection response: %w", err)
	}

	// Convert to SDL
	sdl := introspectionToSDL(&introspection.Schema)

	// Parse SDL using gqlparser
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  "introspection",
		Input: sdl,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	return schema, nil
}

// introspectionResponse is the response from an introspection query.
type introspectionResponse struct {
	Schema introspectionSchema `json:"__schema"` //nolint:tagliatelle // GraphQL spec
}

type introspectionSchema struct {
	QueryType        *typeName           `json:"queryType"`
	MutationType     *typeName           `json:"mutationType"`
	SubscriptionType *typeName           `json:"subscriptionType"`
	Types            []introspectionType `json:"types"`
	Directives       []directive         `json:"directives"`
}

type typeName struct {
	Name string `json:"name"`
}

type introspectionType struct {
	Kind          string       `json:"kind"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	Fields        []field      `json:"fields"`
	InputFields   []inputValue `json:"inputFields"`
	Interfaces    []typeRef    `json:"interfaces"`
	EnumValues    []enumValue  `json:"enumValues"`
	PossibleTypes []typeRef    `json:"possibleTypes"`
}

type field struct {
	Name              string       `json:"name"`
	Description       string       `json:"description"`
	Args              []inputValue `json:"args"`
	Type              typeRef      `json:"type"`
	IsDeprecated      bool         `json:"isDeprecated"`
	DeprecationReason string       `json:"deprecationReason"`
}

type inputValue struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Type         typeRef `json:"type"`
	DefaultValue *string `json:"defaultValue"`
}

type typeRef struct {
	Kind   string   `json:"kind"`
	Name   string   `json:"name"`
	OfType *typeRef `json:"ofType"`
}

type enumValue struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	IsDeprecated      bool   `json:"isDeprecated"`
	DeprecationReason string `json:"deprecationReason"`
}

type directive struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Locations   []string     `json:"locations"`
	Args        []inputValue `json:"args"`
}

const introspectionQuery = `
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types {
      ...FullType
    }
    directives {
      name
      description
      locations
      args {
        ...InputValue
      }
    }
  }
}

fragment FullType on __Type {
  kind
  name
  description
  fields(includeDeprecated: true) {
    name
    description
    args {
      ...InputValue
    }
    type {
      ...TypeRef
    }
    isDeprecated
    deprecationReason
  }
  inputFields {
    ...InputValue
  }
  interfaces {
    ...TypeRef
  }
  enumValues(includeDeprecated: true) {
    name
    description
    isDeprecated
    deprecationReason
  }
  possibleTypes {
    ...TypeRef
  }
}

fragment InputValue on __InputValue {
  name
  description
  type {
    ...TypeRef
  }
  defaultValue
}

fragment TypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
              }
            }
          }
        }
      }
    }
  }
}
`

// FormatType returns a string representation of a Type.
func FormatType(t *ast.Type) string {
	if t == nil {
		return ""
	}

	var result string
	if t.Elem != nil {
		result = "[" + FormatType(t.Elem) + "]"
	} else {
		result = t.NamedType
	}

	if t.NonNull {
		result += "!"
	}

	return result
}

// UnwrapType returns the underlying named type.
func UnwrapType(t *ast.Type) string {
	if t == nil {
		return ""
	}

	if t.Elem != nil {
		return UnwrapType(t.Elem)
	}

	return t.NamedType
}
