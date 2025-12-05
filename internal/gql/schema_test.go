package gql

import (
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
)

func TestFormatType(t *testing.T) {
	tests := []struct {
		name string
		typ  *ast.Type
		want string
	}{
		{
			name: "nil",
			typ:  nil,
			want: "",
		},
		{
			name: "simple named type",
			typ:  &ast.Type{NamedType: "String"},
			want: "String",
		},
		{
			name: "non-null named type",
			typ:  &ast.Type{NamedType: "String", NonNull: true},
			want: "String!",
		},
		{
			name: "list type",
			typ:  &ast.Type{Elem: &ast.Type{NamedType: "String"}},
			want: "[String]",
		},
		{
			name: "non-null list type",
			typ:  &ast.Type{Elem: &ast.Type{NamedType: "String"}, NonNull: true},
			want: "[String]!",
		},
		{
			name: "list of non-null type",
			typ:  &ast.Type{Elem: &ast.Type{NamedType: "String", NonNull: true}},
			want: "[String!]",
		},
		{
			name: "non-null list of non-null type",
			typ:  &ast.Type{Elem: &ast.Type{NamedType: "String", NonNull: true}, NonNull: true},
			want: "[String!]!",
		},
		{
			name: "nested list",
			typ:  &ast.Type{Elem: &ast.Type{Elem: &ast.Type{NamedType: "Int"}}},
			want: "[[Int]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatType(tt.typ)
			if got != tt.want {
				t.Errorf("FormatType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUnwrapType(t *testing.T) {
	tests := []struct {
		name string
		typ  *ast.Type
		want string
	}{
		{
			name: "nil",
			typ:  nil,
			want: "",
		},
		{
			name: "simple named type",
			typ:  &ast.Type{NamedType: "User"},
			want: "User",
		},
		{
			name: "non-null named type",
			typ:  &ast.Type{NamedType: "User", NonNull: true},
			want: "User",
		},
		{
			name: "list type",
			typ:  &ast.Type{Elem: &ast.Type{NamedType: "User"}},
			want: "User",
		},
		{
			name: "non-null list of non-null type",
			typ:  &ast.Type{Elem: &ast.Type{NamedType: "User", NonNull: true}, NonNull: true},
			want: "User",
		},
		{
			name: "nested list",
			typ:  &ast.Type{Elem: &ast.Type{Elem: &ast.Type{NamedType: "Int"}}},
			want: "Int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnwrapType(tt.typ)
			if got != tt.want {
				t.Errorf("UnwrapType() = %q, want %q", got, tt.want)
			}
		})
	}
}
