package repl

import (
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
)

func TestBuildSelectionString(t *testing.T) {
	t.Parallel()

	r := &REPL{schema: &ast.Schema{}}

	tests := []struct {
		name     string
		fields   []selectedField
		depth    int
		expected string
	}{
		{
			name:     "empty fields",
			fields:   nil,
			depth:    1,
			expected: "",
		},
		{
			name: "simple fields",
			fields: []selectedField{
				{name: "id"},
				{name: "name"},
				{name: "email"},
			},
			depth:    1,
			expected: "{ id name email }",
		},
		{
			name: "nested fields",
			fields: []selectedField{
				{name: "id"},
				{name: "posts", children: []selectedField{
					{name: "id"},
					{name: "title"},
				}},
			},
			depth:    1,
			expected: "{\n    id\n    posts { id title }\n  }",
		},
		{
			name: "deeply nested fields",
			fields: []selectedField{
				{name: "user", children: []selectedField{
					{name: "id"},
					{name: "posts", children: []selectedField{
						{name: "id"},
						{name: "title"},
					}},
				}},
			},
			depth:    1,
			expected: "{\n    user {\n      id\n      posts { id title }\n    }\n  }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := r.buildSelectionString(tt.fields, tt.depth)
			if got != tt.expected {
				t.Errorf("buildSelectionString() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestBuildQueryWithSelection(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Types: map[string]*ast.Definition{
			"User": {
				Kind: ast.Object,
				Name: "User",
				Fields: ast.FieldList{
					{Name: "id", Type: ast.NonNullNamedType("ID", nil)},
					{Name: "name", Type: ast.NonNullNamedType("String", nil)},
				},
			},
		},
	}

	r := &REPL{schema: schema}

	tests := []struct {
		name      string
		opType    string
		field     *ast.FieldDefinition
		args      map[string]any
		selection []selectedField
		contains  []string
	}{
		{
			name:   "query with selected fields",
			opType: "query",
			field: &ast.FieldDefinition{
				Name: "users",
				Type: ast.ListType(ast.NonNullNamedType("User", nil), nil),
			},
			args: nil,
			selection: []selectedField{
				{name: "id"},
				{name: "name"},
			},
			contains: []string{"query {", "users", "{ id name }"},
		},
		{
			name:   "query with args and selection",
			opType: "query",
			field: &ast.FieldDefinition{
				Name: "user",
				Type: ast.NamedType("User", nil),
				Arguments: ast.ArgumentDefinitionList{
					{Name: "id", Type: ast.NonNullNamedType("ID", nil)},
				},
			},
			args: map[string]any{"id": "123"},
			selection: []selectedField{
				{name: "id"},
			},
			contains: []string{"query {", "user(id:", "{ id }"},
		},
		{
			name:   "mutation with selection",
			opType: "mutation",
			field: &ast.FieldDefinition{
				Name: "createUser",
				Type: ast.NamedType("User", nil),
			},
			args: nil,
			selection: []selectedField{
				{name: "id"},
				{name: "name"},
			},
			contains: []string{"mutation {", "createUser", "{ id name }"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := r.buildQueryWithSelection(tt.opType, tt.field, tt.args, tt.selection)

			for _, want := range tt.contains {
				if !containsString(got, want) {
					t.Errorf("buildQueryWithSelection() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
