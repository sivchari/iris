package repl

import (
	"strings"
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
)

func TestBuildSelectionString_Empty(t *testing.T) {
	t.Parallel()

	r := &REPL{schema: &ast.Schema{}}
	got := r.buildSelectionString(nil, 1)

	if got != "" {
		t.Errorf("buildSelectionString() = %q, want %q", got, "")
	}
}

func TestBuildSelectionString_Simple(t *testing.T) {
	t.Parallel()

	r := &REPL{schema: &ast.Schema{}}
	fields := []selectedField{
		{name: "id"},
		{name: "name"},
		{name: "email"},
	}

	got := r.buildSelectionString(fields, 1)
	expected := "{ id name email }"

	if got != expected {
		t.Errorf("buildSelectionString() = %q, want %q", got, expected)
	}
}

func TestBuildSelectionString_Nested(t *testing.T) {
	t.Parallel()

	r := &REPL{schema: &ast.Schema{}}
	fields := []selectedField{
		{name: "id"},
		{name: "posts", children: []selectedField{
			{name: "id"},
			{name: "title"},
		}},
	}

	got := r.buildSelectionString(fields, 1)
	expected := "{\n    id\n    posts { id title }\n  }"

	if got != expected {
		t.Errorf("buildSelectionString() = %q, want %q", got, expected)
	}
}

func TestBuildSelectionString_DeeplyNested(t *testing.T) {
	t.Parallel()

	r := &REPL{schema: &ast.Schema{}}
	fields := []selectedField{
		{name: "user", children: []selectedField{
			{name: "id"},
			{name: "posts", children: []selectedField{
				{name: "id"},
				{name: "title"},
			}},
		}},
	}

	got := r.buildSelectionString(fields, 1)
	expected := "{\n    user {\n      id\n      posts { id title }\n    }\n  }"

	if got != expected {
		t.Errorf("buildSelectionString() = %q, want %q", got, expected)
	}
}

func TestBuildQueryWithSelection_Query(t *testing.T) {
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

	field := &ast.FieldDefinition{
		Name: "users",
		Type: ast.ListType(ast.NonNullNamedType("User", nil), nil),
	}
	selection := []selectedField{{name: "id"}, {name: "name"}}

	got := r.buildQueryWithSelection("query", field, nil, selection)

	for _, want := range []string{"query {", "users", "{ id name }"} {
		if !strings.Contains(got, want) {
			t.Errorf("buildQueryWithSelection() = %q, should contain %q", got, want)
		}
	}
}

func TestBuildQueryWithSelection_Mutation(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Types: map[string]*ast.Definition{
			"User": {Kind: ast.Object, Name: "User"},
		},
	}
	r := &REPL{schema: schema}

	field := &ast.FieldDefinition{
		Name: "createUser",
		Type: ast.NamedType("User", nil),
	}
	selection := []selectedField{{name: "id"}, {name: "name"}}

	got := r.buildQueryWithSelection("mutation", field, nil, selection)

	for _, want := range []string{"mutation {", "createUser", "{ id name }"} {
		if !strings.Contains(got, want) {
			t.Errorf("buildQueryWithSelection() = %q, should contain %q", got, want)
		}
	}
}
