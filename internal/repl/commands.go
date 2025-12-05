package repl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/sivchari/iris/internal/client"
	"github.com/sivchari/iris/internal/gql"
)

// errInputCanceled is returned when user cancels input with Ctrl+D.
var errInputCanceled = fmt.Errorf("input canceled")

// cmdHelp displays available commands.
func (r *REPL) cmdHelp() error {
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	commands := []struct {
		name, aliases, desc string
	}{
		{"help", "h, ?", "Show this help message"},
		{"show", "", "Show schema info (types, queries, mutations)"},
		{"desc", "describe", "Describe a type or field"},
		{"call", "", "Call a query or mutation interactively"},
		{"exit", "quit, q", "Exit the REPL"},
	}

	fmt.Println(cyan("Commands:"))

	for _, c := range commands {
		aliases := ""
		if c.aliases != "" {
			aliases = " (" + c.aliases + ")"
		}

		fmt.Printf("  %s%s - %s\n", cyan(c.name), yellow(aliases), c.desc)
	}

	fmt.Println()
	fmt.Println(cyan("Tips:"))
	fmt.Println("  - Press TAB for auto-completion")
	fmt.Println("  - Type raw GraphQL queries starting with '{' or 'query'")

	return nil
}

// cmdShow displays schema information.
func (r *REPL) cmdShow(args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: show [types|queries|mutations]")

		return nil
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	switch args[0] {
	case "types":
		fmt.Println(cyan("Types:"))

		for _, t := range r.schema.Types {
			if !strings.HasPrefix(t.Name, "__") {
				fmt.Printf("  %s %s\n", yellow(string(t.Kind)), t.Name)
			}
		}
	case "queries":
		r.showFields(cyan("Queries:"), r.schema.Query)
	case "mutations":
		r.showFields(cyan("Mutations:"), r.schema.Mutation)
	default:
		return fmt.Errorf("unknown: %s (use: types, queries, mutations)", args[0])
	}

	return nil
}

func (r *REPL) showFields(title string, def *ast.Definition) {
	yellow := color.New(color.FgYellow).SprintFunc()

	if def == nil {
		fmt.Println("Not defined")

		return
	}

	fmt.Println(title)

	for _, f := range def.Fields {
		if !strings.HasPrefix(f.Name, "__") {
			fmt.Printf("  %s: %s\n", yellow(f.Name), gql.FormatType(f.Type))
		}
	}
}

// cmdDesc describes a type or field.
func (r *REPL) cmdDesc(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: desc <type> [field]")
	}

	typeName, fieldName := parseTypeAndField(args)

	t := r.schema.Types[typeName]
	if t == nil {
		return fmt.Errorf("type not found: %s", typeName)
	}

	if fieldName != "" {
		return r.describeField(t, typeName, fieldName)
	}

	return r.describeType(t)
}

func parseTypeAndField(args []string) (typeName, fieldName string) {
	typeName = args[0]

	if idx := strings.Index(typeName, "."); idx != -1 {
		return typeName[:idx], typeName[idx+1:]
	}

	if len(args) > 1 {
		return typeName, args[1]
	}

	return typeName, ""
}

func (r *REPL) describeField(t *ast.Definition, typeName, fieldName string) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()

	for _, f := range t.Fields {
		if f.Name != fieldName {
			continue
		}

		fmt.Printf("%s.%s: %s\n", cyan(typeName), yellow(f.Name), green(gql.FormatType(f.Type)))

		if f.Description != "" {
			fmt.Printf("  %s\n", gray(f.Description))
		}

		if len(f.Arguments) > 0 {
			r.printFieldArguments(f.Arguments, yellow)
		}

		return nil
	}

	return fmt.Errorf("field not found: %s.%s", typeName, fieldName)
}

func (r *REPL) printFieldArguments(args ast.ArgumentDefinitionList, yellow func(a ...any) string) {
	fmt.Println("  Arguments:")

	for _, a := range args {
		req := ""
		if a.Type.NonNull {
			req = yellow(" (required)")
		}

		fmt.Printf("    %s: %s%s\n", a.Name, gql.FormatType(a.Type), req)
	}
}

func (r *REPL) describeType(t *ast.Definition) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()

	fmt.Printf("%s %s\n", cyan(string(t.Kind)), yellow(t.Name))

	if t.Description != "" {
		fmt.Printf("  %s\n", gray(t.Description))
	}

	switch t.Kind {
	case ast.Object, ast.Interface, ast.InputObject:
		for _, f := range t.Fields {
			if !strings.HasPrefix(f.Name, "__") {
				fmt.Printf("  %s: %s\n", yellow(f.Name), green(gql.FormatType(f.Type)))
			}
		}
	case ast.Enum:
		for _, v := range t.EnumValues {
			fmt.Printf("  %s\n", yellow(v.Name))
		}
	case ast.Union:
		for _, pt := range t.Types {
			fmt.Printf("  %s\n", yellow(pt))
		}
	case ast.Scalar: // No fields to display
	}

	return nil
}

// cmdCall executes a query or mutation interactively.
func (r *REPL) cmdCall(args []string) error {
	if len(args) == 0 {
		return r.showCallable()
	}

	name := args[0]

	// Find in queries
	if f := r.findField(r.schema.Query, name); f != nil {
		return r.executeField("query", f)
	}
	// Find in mutations
	if f := r.findField(r.schema.Mutation, name); f != nil {
		return r.executeField("mutation", f)
	}

	return fmt.Errorf("not found: %s", name)
}

func (r *REPL) showCallable() error {
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Println(cyan("Available:"))

	if r.schema.Query != nil {
		for _, f := range r.schema.Query.Fields {
			if !strings.HasPrefix(f.Name, "__") {
				fmt.Printf("  %s (query)\n", f.Name)
			}
		}
	}

	if r.schema.Mutation != nil {
		for _, f := range r.schema.Mutation.Fields {
			if !strings.HasPrefix(f.Name, "__") {
				fmt.Printf("  %s (mutation)\n", f.Name)
			}
		}
	}

	fmt.Println("\nUsage: call <name>")

	return nil
}

func (r *REPL) findField(def *ast.Definition, name string) *ast.FieldDefinition {
	if def == nil {
		return nil
	}

	for _, f := range def.Fields {
		if f.Name == name {
			return f
		}
	}

	return nil
}

func (r *REPL) executeField(opType string, field *ast.FieldDefinition) error {
	cyan := color.New(color.FgCyan).SprintFunc()

	// Read arguments
	args, err := r.readArgs(field.Arguments)
	if err != nil {
		if errors.Is(err, errInputCanceled) {
			fmt.Println("Canceled.")

			return nil
		}

		return err
	}

	// Build query
	query := r.buildQuery(opType, field, args)

	fmt.Println()
	fmt.Println(cyan("Query:"))
	fmt.Println(query)
	fmt.Println()

	// Execute
	resp, err := r.client.Execute(context.Background(), &client.Request{Query: query})
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	return r.printResponse(resp)
}

func (r *REPL) readArgs(argDefs ast.ArgumentDefinitionList) (map[string]any, error) {
	if len(argDefs) == 0 {
		return make(map[string]any), nil
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()

	fmt.Println(cyan("Arguments:"))

	reader := bufio.NewReader(os.Stdin)
	result := make(map[string]any)

	for _, arg := range argDefs {
		typeStr := gql.FormatType(arg.Type)

		req := ""
		if arg.Type.NonNull {
			req = yellow(" (required)")
		}

		fmt.Printf("  %s %s%s: ", cyan(arg.Name), gray(typeStr), req)

		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, errInputCanceled
		}

		input = strings.TrimSpace(input)
		if input == "" {
			if arg.Type.NonNull && arg.DefaultValue == nil {
				fmt.Println("    Required field.")

				return r.readArgs(argDefs) // Retry
			}

			continue
		}

		result[arg.Name] = r.parseValue(input, arg.Type)
	}

	return result, nil
}

func (r *REPL) parseValue(input string, t *ast.Type) any {
	// Strip quotes if provided
	input = strings.Trim(input, "\"'")

	switch t.NamedType {
	case "Int":
		var v int
		_, _ = fmt.Sscanf(input, "%d", &v)

		return v
	case "Float":
		var v float64
		_, _ = fmt.Sscanf(input, "%f", &v)

		return v
	case "Boolean":
		return strings.EqualFold(input, "true")
	default:
		return input
	}
}

func (r *REPL) buildQuery(opType string, field *ast.FieldDefinition, args map[string]any) string {
	var sb strings.Builder

	sb.WriteString(opType + " {\n  " + field.Name)

	if len(args) > 0 {
		sb.WriteString("(")

		first := true
		for k, v := range args {
			if !first {
				sb.WriteString(", ")
			}

			sb.WriteString(k + ": " + r.formatArg(v, r.findArgType(field, k)))

			first = false
		}

		sb.WriteString(")")
	}

	// Add selection set for object types
	if sel := r.buildSelection(field.Type); sel != "" {
		sb.WriteString(" " + sel)
	}

	sb.WriteString("\n}")

	return sb.String()
}

func (r *REPL) findArgType(field *ast.FieldDefinition, name string) *ast.Type {
	for _, a := range field.Arguments {
		if a.Name == name {
			return a.Type
		}
	}

	return nil
}

func (r *REPL) formatArg(v any, t *ast.Type) string {
	switch val := v.(type) {
	case string:
		// Quote strings and IDs
		if t != nil {
			typeName := gql.UnwrapType(t)
			if typeName != "ID" && typeName != "String" {
				return val // Enum, no quotes
			}
		}

		return fmt.Sprintf("%q", val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (r *REPL) buildSelection(t *ast.Type) string {
	typeName := gql.UnwrapType(t)

	def := r.schema.Types[typeName]
	if def == nil || len(def.Fields) == 0 {
		return ""
	}

	// Select scalar/enum fields
	var fields []string

	for _, f := range def.Fields {
		if strings.HasPrefix(f.Name, "__") {
			continue
		}

		fieldType := gql.UnwrapType(f.Type)

		fieldDef := r.schema.Types[fieldType]
		if fieldDef == nil || fieldDef.Kind == ast.Scalar || fieldDef.Kind == ast.Enum {
			fields = append(fields, f.Name)
		}
	}

	if len(fields) == 0 {
		return ""
	}

	return "{ " + strings.Join(fields, " ") + " }"
}
