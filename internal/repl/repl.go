// Package repl provides the interactive GraphQL shell.
package repl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	prompt "github.com/ktr0731/go-prompt"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/sivchari/iris/internal/client"
	"github.com/sivchari/iris/internal/federation"
	_ "github.com/sivchari/iris/internal/federation/apollo" // Register Apollo Federation provider
	"github.com/sivchari/iris/internal/gql"
)

var errExit = fmt.Errorf("exit")

// REPL is the interactive GraphQL shell.
type REPL struct {
	client     *client.Client
	schema     *ast.Schema
	federation *federation.Info
	completer  *gql.Completer
	prompt     *prompt.Prompt
}

// New creates a new REPL.
func New(c *client.Client, schema *ast.Schema) *REPL {
	r := &REPL{
		client:     c,
		schema:     schema,
		federation: federation.Detect(schema),
		completer:  gql.NewCompleter(schema),
	}

	r.prompt = prompt.New(
		r.executor,
		r.completer.Complete,
		prompt.OptionTitle("iris"),
		prompt.OptionPrefix("iris> "),
		prompt.OptionPrefixTextColor(prompt.Green),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionMaxSuggestion(10),
		prompt.OptionShowCompletionAtStart(),
	)

	return r
}

// Run starts the REPL.
func (r *REPL) Run() error {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println(cyan("iris") + " - GraphQL REPL")
	fmt.Println("Type 'help' for commands, TAB for completion.")

	if r.federation != nil {
		fmt.Printf("%s: %s (subgraph)\n", green("Federation"), r.federation.Provider.Name())
	}

	fmt.Println()

	r.prompt.Run()

	return nil
}

// Close closes the REPL.
func (r *REPL) Close() error {
	return nil
}

func (r *REPL) executor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	if err := r.execute(input); err != nil {
		if errors.Is(err, errExit) {
			fmt.Println("Goodbye!")
			os.Exit(0)
		}

		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintln(os.Stderr, red("Error:"), err)
	}
}

func (r *REPL) execute(input string) error {
	// Raw GraphQL query
	if isGraphQL(input) {
		return r.executeRaw(input)
	}

	// Command
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	cmd, args := parts[0], parts[1:]
	switch cmd {
	case "help", "h", "?":
		return r.cmdHelp()
	case "show":
		return r.cmdShow(args)
	case "desc", "describe":
		return r.cmdDesc(args)
	case "call":
		return r.cmdCall(args)
	case "exit", "quit", "q":
		return errExit
	default:
		return fmt.Errorf("unknown: %s (type 'help')", cmd)
	}
}

func isGraphQL(s string) bool {
	s = strings.TrimSpace(s)

	return strings.HasPrefix(s, "{") ||
		strings.HasPrefix(s, "query") ||
		strings.HasPrefix(s, "mutation")
}

func (r *REPL) executeRaw(query string) error {
	resp, err := r.client.Execute(context.Background(), &client.Request{Query: query})
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	return r.printResponse(resp)
}

func (r *REPL) printResponse(resp *client.Response) error {
	if len(resp.Errors) > 0 {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Println(red("Errors:"))

		for _, e := range resp.Errors {
			fmt.Printf("  - %s\n", e.Message)
		}

		fmt.Println()
	}

	if resp.Data != nil {
		var data any
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		out, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(out))
	}

	return nil
}
