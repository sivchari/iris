package gql

import (
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
)

func TestCompleter_completeShow(t *testing.T) {
	c := NewCompleter(&ast.Schema{})

	tests := []struct {
		name   string
		prefix string
		want   []string
	}{
		{
			name:   "empty prefix",
			prefix: "",
			want:   []string{"types", "queries", "mutations", "federation"},
		},
		{
			name:   "prefix t",
			prefix: "t",
			want:   []string{"types"},
		},
		{
			name:   "prefix q",
			prefix: "q",
			want:   []string{"queries"},
		},
		{
			name:   "prefix m",
			prefix: "m",
			want:   []string{"mutations"},
		},
		{
			name:   "prefix f",
			prefix: "f",
			want:   []string{"federation"},
		},
		{
			name:   "no match",
			prefix: "xyz",
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.completeShow(tt.prefix)
			if len(got) != len(tt.want) {
				t.Errorf("completeShow(%q) returned %d suggestions, want %d", tt.prefix, len(got), len(tt.want))

				return
			}

			for i, want := range tt.want {
				if got[i].Text != want {
					t.Errorf("completeShow(%q)[%d].Text = %q, want %q", tt.prefix, i, got[i].Text, want)
				}
			}
		})
	}
}

func TestCompleter_completeTypes(t *testing.T) {
	schema := &ast.Schema{
		Types: map[string]*ast.Definition{
			"User":        {Name: "User", Kind: ast.Object},
			"Post":        {Name: "Post", Kind: ast.Object},
			"Status":      {Name: "Status", Kind: ast.Enum},
			"__Schema":    {Name: "__Schema", Kind: ast.Object},
			"__Type":      {Name: "__Type", Kind: ast.Object},
			"__Directive": {Name: "__Directive", Kind: ast.Object},
		},
	}

	c := NewCompleter(schema)

	tests := []struct {
		name      string
		prefix    string
		wantCount int
	}{
		{
			name:      "empty prefix returns non-introspection types",
			prefix:    "",
			wantCount: 3, // User, Post, Status (excludes __ types)
		},
		{
			name:      "prefix U",
			prefix:    "U",
			wantCount: 1, // User
		},
		{
			name:      "prefix P",
			prefix:    "P",
			wantCount: 1, // Post
		},
		{
			name:      "no match",
			prefix:    "xyz",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.completeTypes(tt.prefix)
			if len(got) != tt.wantCount {
				t.Errorf("completeTypes(%q) returned %d suggestions, want %d", tt.prefix, len(got), tt.wantCount)
			}

			// Verify no __ types are included
			for _, s := range got {
				if len(s.Text) >= 2 && s.Text[:2] == "__" {
					t.Errorf("completeTypes(%q) included introspection type %q", tt.prefix, s.Text)
				}
			}
		})
	}
}

func TestCompleter_completeCall(t *testing.T) {
	schema := &ast.Schema{
		Query: &ast.Definition{
			Fields: []*ast.FieldDefinition{
				{Name: "users"},
				{Name: "user"},
				{Name: "__schema"},
			},
		},
		Mutation: &ast.Definition{
			Fields: []*ast.FieldDefinition{
				{Name: "createUser"},
				{Name: "deleteUser"},
			},
		},
	}

	c := NewCompleter(schema)

	tests := []struct {
		name      string
		prefix    string
		wantCount int
	}{
		{
			name:      "empty prefix",
			prefix:    "",
			wantCount: 4, // users, user, createUser, deleteUser (excludes __schema)
		},
		{
			name:      "prefix u",
			prefix:    "u",
			wantCount: 2, // users, user
		},
		{
			name:      "prefix create",
			prefix:    "create",
			wantCount: 1, // createUser
		},
		{
			name:      "no match",
			prefix:    "xyz",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.completeCall(tt.prefix)
			if len(got) != tt.wantCount {
				t.Errorf("completeCall(%q) returned %d suggestions, want %d", tt.prefix, len(got), tt.wantCount)
			}
		})
	}
}

func TestCompleter_completeGraphQL(t *testing.T) {
	schema := &ast.Schema{
		Query: &ast.Definition{
			Fields: []*ast.FieldDefinition{
				{Name: "users", Type: &ast.Type{NamedType: "User"}},
				{Name: "posts", Type: &ast.Type{NamedType: "Post"}},
				{Name: "__schema"},
			},
		},
	}

	c := NewCompleter(schema)

	tests := []struct {
		name      string
		prefix    string
		wantCount int
	}{
		{
			name:      "empty prefix",
			prefix:    "",
			wantCount: 2, // users, posts
		},
		{
			name:      "prefix u",
			prefix:    "u",
			wantCount: 1, // users
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.completeGraphQL(tt.prefix)
			if len(got) != tt.wantCount {
				t.Errorf("completeGraphQL(%q) returned %d suggestions, want %d", tt.prefix, len(got), tt.wantCount)
			}
		})
	}
}
