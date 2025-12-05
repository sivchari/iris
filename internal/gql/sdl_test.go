package gql

import "testing"

func TestTypeRefToString(t *testing.T) {
	tests := []struct {
		name string
		ref  *typeRef
		want string
	}{
		{
			name: "named type",
			ref:  &typeRef{Kind: "SCALAR", Name: "String"},
			want: "String",
		},
		{
			name: "non-null type",
			ref:  &typeRef{Kind: "NON_NULL", OfType: &typeRef{Kind: "SCALAR", Name: "String"}},
			want: "String!",
		},
		{
			name: "list type",
			ref:  &typeRef{Kind: "LIST", OfType: &typeRef{Kind: "SCALAR", Name: "Int"}},
			want: "[Int]",
		},
		{
			name: "non-null list of non-null",
			ref: &typeRef{
				Kind: "NON_NULL",
				OfType: &typeRef{
					Kind: "LIST",
					OfType: &typeRef{
						Kind:   "NON_NULL",
						OfType: &typeRef{Kind: "SCALAR", Name: "String"},
					},
				},
			},
			want: "[String!]!",
		},
		{
			name: "object type",
			ref:  &typeRef{Kind: "OBJECT", Name: "User"},
			want: "User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeRefToString(tt.ref)
			if got != tt.want {
				t.Errorf("typeRefToString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsBuiltinScalar(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"String", true},
		{"Int", true},
		{"Float", true},
		{"Boolean", true},
		{"ID", true},
		{"DateTime", false},
		{"JSON", false},
		{"User", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBuiltinScalar(tt.name)
			if got != tt.want {
				t.Errorf("isBuiltinScalar(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestEscapeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no escape needed",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "escape quotes",
			input: `say "hello"`,
			want:  `say \"hello\"`,
		},
		{
			name:  "escape backslash",
			input: `path\to\file`,
			want:  `path\\to\\file`,
		},
		{
			name:  "escape both",
			input: `"path\to"`,
			want:  `\"path\\to\"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeString(tt.input)
			if got != tt.want {
				t.Errorf("escapeString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatDescription(t *testing.T) {
	tests := []struct {
		name   string
		desc   string
		indent string
		want   string
	}{
		{
			name:   "single line",
			desc:   "A simple description",
			indent: "",
			want:   "\"A simple description\"\n",
		},
		{
			name:   "single line with indent",
			desc:   "A field description",
			indent: "  ",
			want:   "  \"A field description\"\n",
		},
		{
			name:   "multi line",
			desc:   "Line 1\nLine 2",
			indent: "",
			want:   "\"\"\"\nLine 1\nLine 2\n\"\"\"\n",
		},
		{
			name:   "multi line with indent",
			desc:   "Line 1\nLine 2",
			indent: "  ",
			want:   "  \"\"\"\n  Line 1\nLine 2\n  \"\"\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDescription(tt.desc, tt.indent)
			if got != tt.want {
				t.Errorf("formatDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTypeToSDL_Scalar(t *testing.T) {
	t.Run("builtin scalar", func(t *testing.T) {
		got := typeToSDL(&introspectionType{Kind: "SCALAR", Name: "String"})
		if got != "" {
			t.Errorf("typeToSDL() = %q, want empty", got)
		}
	})

	t.Run("custom scalar", func(t *testing.T) {
		got := typeToSDL(&introspectionType{Kind: "SCALAR", Name: "DateTime"})
		if got != "scalar DateTime" {
			t.Errorf("typeToSDL() = %q, want %q", got, "scalar DateTime")
		}
	})
}

func TestTypeToSDL_Enum(t *testing.T) {
	t.Run("enum", func(t *testing.T) {
		typ := &introspectionType{
			Kind: "ENUM",
			Name: "Status",
			EnumValues: []enumValue{
				{Name: "ACTIVE"},
				{Name: "INACTIVE"},
			},
		}
		want := "enum Status {\n  ACTIVE\n  INACTIVE\n}"

		got := typeToSDL(typ)
		if got != want {
			t.Errorf("typeToSDL() = %q, want %q", got, want)
		}
	})

	t.Run("enum with deprecated value", func(t *testing.T) {
		typ := &introspectionType{
			Kind: "ENUM",
			Name: "Status",
			EnumValues: []enumValue{
				{Name: "ACTIVE"},
				{Name: "OLD", IsDeprecated: true, DeprecationReason: "use ACTIVE"},
			},
		}
		want := "enum Status {\n  ACTIVE\n  OLD @deprecated(reason: \"use ACTIVE\")\n}"

		got := typeToSDL(typ)
		if got != want {
			t.Errorf("typeToSDL() = %q, want %q", got, want)
		}
	})
}

func TestTypeToSDL_Object(t *testing.T) {
	t.Run("simple object", func(t *testing.T) {
		typ := &introspectionType{
			Kind: "OBJECT",
			Name: "User",
			Fields: []field{
				{Name: "id", Type: typeRef{Kind: "SCALAR", Name: "ID"}},
				{Name: "name", Type: typeRef{Kind: "SCALAR", Name: "String"}},
			},
		}
		want := "type User {\n  id: ID\n  name: String\n}"

		got := typeToSDL(typ)
		if got != want {
			t.Errorf("typeToSDL() = %q, want %q", got, want)
		}
	})

	t.Run("object with interface", func(t *testing.T) {
		typ := &introspectionType{
			Kind:       "OBJECT",
			Name:       "User",
			Interfaces: []typeRef{{Name: "Node"}},
			Fields:     []field{{Name: "id", Type: typeRef{Kind: "SCALAR", Name: "ID"}}},
		}
		want := "type User implements Node {\n  id: ID\n}"

		got := typeToSDL(typ)
		if got != want {
			t.Errorf("typeToSDL() = %q, want %q", got, want)
		}
	})
}

func TestTypeToSDL_Other(t *testing.T) {
	t.Run("interface", func(t *testing.T) {
		typ := &introspectionType{
			Kind: "INTERFACE",
			Name: "Node",
			Fields: []field{
				{Name: "id", Type: typeRef{Kind: "NON_NULL", OfType: &typeRef{Kind: "SCALAR", Name: "ID"}}},
			},
		}
		want := "interface Node {\n  id: ID!\n}"

		got := typeToSDL(typ)
		if got != want {
			t.Errorf("typeToSDL() = %q, want %q", got, want)
		}
	})

	t.Run("union", func(t *testing.T) {
		typ := &introspectionType{
			Kind:          "UNION",
			Name:          "SearchResult",
			PossibleTypes: []typeRef{{Name: "User"}, {Name: "Post"}},
		}
		want := "union SearchResult = User | Post"

		got := typeToSDL(typ)
		if got != want {
			t.Errorf("typeToSDL() = %q, want %q", got, want)
		}
	})

	t.Run("input object", func(t *testing.T) {
		typ := &introspectionType{
			Kind: "INPUT_OBJECT",
			Name: "CreateUserInput",
			InputFields: []inputValue{
				{Name: "name", Type: typeRef{Kind: "NON_NULL", OfType: &typeRef{Kind: "SCALAR", Name: "String"}}},
				{Name: "email", Type: typeRef{Kind: "SCALAR", Name: "String"}},
			},
		}
		want := "input CreateUserInput {\n  name: String!\n  email: String\n}"

		got := typeToSDL(typ)
		if got != want {
			t.Errorf("typeToSDL() = %q, want %q", got, want)
		}
	})

	t.Run("unknown kind", func(t *testing.T) {
		got := typeToSDL(&introspectionType{Kind: "UNKNOWN", Name: "Foo"})
		if got != "" {
			t.Errorf("typeToSDL() = %q, want empty", got)
		}
	})
}

func TestFieldToSDL(t *testing.T) {
	tests := []struct {
		name  string
		field *field
		want  string
	}{
		{
			name:  "simple field",
			field: &field{Name: "id", Type: typeRef{Kind: "SCALAR", Name: "ID"}},
			want:  "  id: ID\n",
		},
		{
			name: "field with args",
			field: &field{
				Name: "user",
				Args: []inputValue{
					{Name: "id", Type: typeRef{Kind: "NON_NULL", OfType: &typeRef{Kind: "SCALAR", Name: "ID"}}},
				},
				Type: typeRef{Kind: "OBJECT", Name: "User"},
			},
			want: "  user(id: ID!): User\n",
		},
		{
			name: "deprecated field",
			field: &field{
				Name:              "oldField",
				Type:              typeRef{Kind: "SCALAR", Name: "String"},
				IsDeprecated:      true,
				DeprecationReason: "use newField",
			},
			want: "  oldField: String @deprecated(reason: \"use newField\")\n",
		},
		{
			name: "deprecated without reason",
			field: &field{
				Name:         "oldField",
				Type:         typeRef{Kind: "SCALAR", Name: "String"},
				IsDeprecated: true,
			},
			want: "  oldField: String @deprecated\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fieldToSDL(tt.field)
			if got != tt.want {
				t.Errorf("fieldToSDL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDirectiveToSDL(t *testing.T) {
	tests := []struct {
		name string
		dir  *directive
		want string
	}{
		{
			name: "simple directive",
			dir: &directive{
				Name:      "auth",
				Locations: []string{"FIELD_DEFINITION"},
			},
			want: "directive @auth on FIELD_DEFINITION",
		},
		{
			name: "directive with args",
			dir: &directive{
				Name: "auth",
				Args: []inputValue{
					{Name: "role", Type: typeRef{Kind: "SCALAR", Name: "String"}},
				},
				Locations: []string{"FIELD_DEFINITION", "OBJECT"},
			},
			want: "directive @auth(role: String) on FIELD_DEFINITION | OBJECT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := directiveToSDL(tt.dir)
			if got != tt.want {
				t.Errorf("directiveToSDL() = %q, want %q", got, tt.want)
			}
		})
	}
}
