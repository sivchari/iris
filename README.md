# iris

iris is an interactive GraphQL REPL client, inspired by [evans](https://github.com/ktr0731/evans) (gRPC client).

## Features

- Interactive REPL with tab completion
- Schema introspection
- Execute queries and mutations interactively
- Custom HTTP headers support
- Pipe and file input support

## Installation

```bash
go install github.com/sivchari/iris/cmd/iris@latest
```

## Usage

### Start REPL

```bash
# Connect to a GraphQL endpoint
iris -e https://api.example.com/graphql

# With authentication header
iris -e https://api.example.com/graphql -H "Authorization: Bearer <token>"
```

### CLI Mode

```bash
# Execute a query directly
iris -e https://api.example.com/graphql -q '{ users { id name } }'

# Read query from file
iris -e https://api.example.com/graphql -f query.graphql

# Pipe query
echo '{ users { id } }' | iris -e https://api.example.com/graphql
```

## REPL Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `help` | `h`, `?` | Show help message |
| `show` | | Show schema info (`types`, `queries`, `mutations`) |
| `desc` | `describe` | Describe a type or field |
| `call` | | Call a query or mutation interactively |
| `exit` | `quit`, `q` | Exit the REPL |

### Examples

```
iris> help
iris> show types
iris> show queries
iris> desc User
iris> desc User.email
iris> call users
iris> { users { id name } }
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--endpoint` | `-e` | GraphQL endpoint (required) |
| `--header` | `-H` | HTTP header (can be specified multiple times) |
| `--query` | `-q` | Execute query directly |
| `--file` | `-f` | Read query from file |

## License

MIT License - see [LICENSE](LICENSE) for details.
