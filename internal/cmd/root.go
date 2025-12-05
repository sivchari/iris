// Package cmd provides the CLI commands for iris.
package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/sivchari/iris/internal/client"
	"github.com/sivchari/iris/internal/gql"
	"github.com/sivchari/iris/internal/repl"
)

var (
	endpoint string
	headers  []string
	query    string
	file     string
)

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iris",
		Short: "GraphQL REPL client",
		Long: `iris - GraphQL REPL client

Examples:
  iris -e https://api.example.com/graphql
  iris -e https://api.example.com/graphql -q '{ users { id } }'
  iris -e https://api.example.com/graphql -H "Authorization: Bearer token"
  echo '{ users { id } }' | iris -e https://api.example.com/graphql`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run()
		},
	}

	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", "", "GraphQL endpoint (required)")
	cmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "HTTP header")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Execute query")
	cmd.Flags().StringVarP(&file, "file", "f", "", "Read query from file")

	return cmd
}

// Execute runs the root command.
func Execute() error {
	if err := NewRootCmd().Execute(); err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	return nil
}

func run() error {
	if endpoint == "" {
		return fmt.Errorf("endpoint required (-e)")
	}

	// Create client
	c := client.New(endpoint, parseHeaders()...)

	// CLI mode or REPL mode
	if q := getQuery(); q != "" {
		return runQuery(c, q)
	}

	return runREPL(c)
}

func parseHeaders() []client.Option {
	var opts []client.Option

	for _, h := range headers {
		if parts := strings.SplitN(h, ":", 2); len(parts) == 2 {
			opts = append(opts, client.WithHeader(
				strings.TrimSpace(parts[0]),
				strings.TrimSpace(parts[1]),
			))
		}
	}

	return opts
}

func getQuery() string {
	if query != "" {
		return query
	}

	if file != "" {
		data, err := os.ReadFile(file) //nolint:gosec // file path from user flag
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		return string(data)
	}
	// Check stdin
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		data, _ := io.ReadAll(bufio.NewReader(os.Stdin))

		return string(data)
	}

	return ""
}

func runQuery(c *client.Client, q string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.Execute(ctx, &client.Request{Query: q})
	if err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

	// Output
	var out any

	if len(resp.Errors) > 0 {
		outMap := map[string]any{"errors": resp.Errors}

		if resp.Data != nil {
			var data any
			if err := json.Unmarshal(resp.Data, &data); err == nil {
				outMap["data"] = data
			}
		}

		out = outMap
	} else {
		_ = json.Unmarshal(resp.Data, &out)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encode output: %w", err)
	}

	return nil
}

func runREPL(c *client.Client) error {
	fmt.Printf("Connecting to %s...\n", endpoint)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	schema, err := gql.LoadSchemaFromIntrospection(ctx, c)
	if err != nil {
		return fmt.Errorf("introspection failed: %w", err)
	}

	fmt.Printf("Loaded %d types.\n\n", len(schema.Types))

	r := repl.New(c, schema)
	defer func() { _ = r.Close() }()

	if err := r.Run(); err != nil {
		return fmt.Errorf("repl: %w", err)
	}

	return nil
}
