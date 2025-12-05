package gql

import (
	"fmt"
	"strings"
)

// introspectionToSDL converts an introspection result to SDL format.
func introspectionToSDL(schema *introspectionSchema) string {
	var sb strings.Builder

	writeSchemaDefinition(&sb, schema)
	writeTypes(&sb, schema.Types)
	writeDirectives(&sb, schema.Directives)

	return sb.String()
}

func writeSchemaDefinition(sb *strings.Builder, schema *introspectionSchema) {
	sb.WriteString("schema {\n")

	if schema.QueryType != nil {
		fmt.Fprintf(sb, "  query: %s\n", schema.QueryType.Name)
	}

	if schema.MutationType != nil {
		fmt.Fprintf(sb, "  mutation: %s\n", schema.MutationType.Name)
	}

	if schema.SubscriptionType != nil {
		fmt.Fprintf(sb, "  subscription: %s\n", schema.SubscriptionType.Name)
	}

	sb.WriteString("}\n\n")
}

func writeTypes(sb *strings.Builder, types []introspectionType) {
	for i := range types {
		t := &types[i]
		if strings.HasPrefix(t.Name, "__") {
			continue
		}

		sdl := typeToSDL(t)
		if sdl != "" {
			sb.WriteString(sdl)
			sb.WriteString("\n\n")
		}
	}
}

func writeDirectives(sb *strings.Builder, directives []directive) {
	for i := range directives {
		d := &directives[i]
		if isBuiltinDirective(d.Name) {
			continue
		}

		sb.WriteString(directiveToSDL(d))
		sb.WriteString("\n\n")
	}
}

func isBuiltinDirective(name string) bool {
	switch name {
	case "skip", "include", "deprecated", "specifiedBy":
		return true
	}

	return false
}

func typeToSDL(t *introspectionType) string {
	var sb strings.Builder

	if t.Description != "" {
		sb.WriteString(formatDescription(t.Description, ""))
	}

	switch t.Kind {
	case "SCALAR":
		return scalarToSDL(t)
	case "OBJECT":
		objectToSDL(&sb, t)
	case "INTERFACE":
		interfaceToSDL(&sb, t)
	case "UNION":
		unionToSDL(&sb, t)
	case "ENUM":
		enumToSDL(&sb, t)
	case "INPUT_OBJECT":
		inputObjectToSDL(&sb, t)
	default:
		return ""
	}

	return sb.String()
}

func scalarToSDL(t *introspectionType) string {
	if isBuiltinScalar(t.Name) {
		return ""
	}

	var sb strings.Builder
	if t.Description != "" {
		sb.WriteString(formatDescription(t.Description, ""))
	}

	sb.WriteString(fmt.Sprintf("scalar %s", t.Name))

	return sb.String()
}

func objectToSDL(sb *strings.Builder, t *introspectionType) {
	fmt.Fprintf(sb, "type %s", t.Name)

	if len(t.Interfaces) > 0 {
		ifaces := make([]string, 0, len(t.Interfaces))
		for _, i := range t.Interfaces {
			ifaces = append(ifaces, i.Name)
		}

		sb.WriteString(" implements ")
		sb.WriteString(strings.Join(ifaces, " & "))
	}

	sb.WriteString(" {\n")

	for i := range t.Fields {
		sb.WriteString(fieldToSDL(&t.Fields[i]))
	}

	sb.WriteString("}")
}

func interfaceToSDL(sb *strings.Builder, t *introspectionType) {
	fmt.Fprintf(sb, "interface %s {\n", t.Name)

	for i := range t.Fields {
		sb.WriteString(fieldToSDL(&t.Fields[i]))
	}

	sb.WriteString("}")
}

func unionToSDL(sb *strings.Builder, t *introspectionType) {
	fmt.Fprintf(sb, "union %s = ", t.Name)

	types := make([]string, 0, len(t.PossibleTypes))
	for _, pt := range t.PossibleTypes {
		types = append(types, pt.Name)
	}

	sb.WriteString(strings.Join(types, " | "))
}

func enumToSDL(sb *strings.Builder, t *introspectionType) {
	fmt.Fprintf(sb, "enum %s {\n", t.Name)

	for _, ev := range t.EnumValues {
		writeEnumValue(sb, &ev)
	}

	sb.WriteString("}")
}

func writeEnumValue(sb *strings.Builder, ev *enumValue) {
	if ev.Description != "" {
		sb.WriteString(formatDescription(ev.Description, "  "))
	}

	fmt.Fprintf(sb, "  %s", ev.Name)

	if ev.IsDeprecated {
		if ev.DeprecationReason != "" {
			fmt.Fprintf(sb, " @deprecated(reason: %q)", ev.DeprecationReason)
		} else {
			sb.WriteString(" @deprecated")
		}
	}

	sb.WriteString("\n")
}

func inputObjectToSDL(sb *strings.Builder, t *introspectionType) {
	fmt.Fprintf(sb, "input %s {\n", t.Name)

	for i := range t.InputFields {
		sb.WriteString(inputValueToSDL(&t.InputFields[i], "  "))
	}

	sb.WriteString("}")
}

func fieldToSDL(f *field) string {
	var sb strings.Builder

	if f.Description != "" {
		sb.WriteString(formatDescription(f.Description, "  "))
	}

	sb.WriteString(fmt.Sprintf("  %s", f.Name))

	if len(f.Args) > 0 {
		sb.WriteString("(")

		var args []string

		for _, arg := range f.Args {
			argStr := fmt.Sprintf("%s: %s", arg.Name, typeRefToString(&arg.Type))
			if arg.DefaultValue != nil {
				argStr += fmt.Sprintf(" = %s", *arg.DefaultValue)
			}

			args = append(args, argStr)
		}

		sb.WriteString(strings.Join(args, ", "))
		sb.WriteString(")")
	}

	sb.WriteString(fmt.Sprintf(": %s", typeRefToString(&f.Type)))

	if f.IsDeprecated {
		if f.DeprecationReason != "" {
			sb.WriteString(fmt.Sprintf(" @deprecated(reason: %q)", f.DeprecationReason))
		} else {
			sb.WriteString(" @deprecated")
		}
	}

	sb.WriteString("\n")

	return sb.String()
}

func inputValueToSDL(iv *inputValue, indent string) string {
	var sb strings.Builder

	if iv.Description != "" {
		sb.WriteString(formatDescription(iv.Description, indent))
	}

	sb.WriteString(fmt.Sprintf("%s%s: %s", indent, iv.Name, typeRefToString(&iv.Type)))

	if iv.DefaultValue != nil {
		sb.WriteString(fmt.Sprintf(" = %s", *iv.DefaultValue))
	}

	sb.WriteString("\n")

	return sb.String()
}

func directiveToSDL(d *directive) string {
	var sb strings.Builder

	if d.Description != "" {
		sb.WriteString(formatDescription(d.Description, ""))
	}

	sb.WriteString(fmt.Sprintf("directive @%s", d.Name))

	if len(d.Args) > 0 {
		sb.WriteString("(")

		var args []string

		for _, arg := range d.Args {
			argStr := fmt.Sprintf("%s: %s", arg.Name, typeRefToString(&arg.Type))
			if arg.DefaultValue != nil {
				argStr += fmt.Sprintf(" = %s", *arg.DefaultValue)
			}

			args = append(args, argStr)
		}

		sb.WriteString(strings.Join(args, ", "))
		sb.WriteString(")")
	}

	sb.WriteString(" on ")
	sb.WriteString(strings.Join(d.Locations, " | "))

	return sb.String()
}

func typeRefToString(t *typeRef) string {
	switch t.Kind {
	case "NON_NULL":
		if t.OfType != nil {
			return typeRefToString(t.OfType) + "!"
		}
	case "LIST":
		if t.OfType != nil {
			return "[" + typeRefToString(t.OfType) + "]"
		}
	default:
		return t.Name
	}

	return ""
}

func formatDescription(desc, indent string) string {
	// Use block string for multi-line descriptions
	if strings.Contains(desc, "\n") {
		return fmt.Sprintf("%s\"\"\"\n%s%s\n%s\"\"\"\n", indent, indent, desc, indent)
	}

	return fmt.Sprintf("%s%q\n", indent, escapeString(desc))
}

func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")

	return s
}

func isBuiltinScalar(name string) bool {
	switch name {
	case "String", "Int", "Float", "Boolean", "ID":
		return true
	}

	return false
}
