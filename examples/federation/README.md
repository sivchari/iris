# GraphQL Federation Example

This example demonstrates a GraphQL Federation setup with gqlgen, combining User and Post schemas into a single server.

## Structure

```
examples/federation/
├── go.mod
├── gqlgen.yml           # Single gqlgen configuration
├── graph/
│   ├── users/
│   │   └── schema.graphqls   # User schema with @key directive
│   ├── posts/
│   │   └── schema.graphqls   # Post schema referencing User
│   ├── model/
│   │   └── models_gen.go     # Generated models
│   ├── generated.go          # Generated GraphQL runtime
│   ├── federation.go         # Federation support
│   ├── resolver.go           # Resolver with sample data
│   ├── schema.resolvers.go   # Query resolvers
│   └── entity.resolvers.go   # Entity resolvers for federation
├── main.go
└── README.md
```

## Running the Server

```bash
cd examples/federation
go run .
```

The server starts at http://localhost:8080/

## Example Queries

### Get all users

```graphql
{
  users {
    id
    name
    email
  }
}
```

### Get all posts with author information

```graphql
{
  posts {
    id
    title
    body
    author {
      id
      name
      email
    }
  }
}
```

### Get a single user

```graphql
{
  user(id: "1") {
    id
    name
    email
  }
}
```

### Get a single post

```graphql
{
  post(id: "1") {
    id
    title
    body
    author {
      name
    }
  }
}
```

## Federation Features

This example uses Apollo Federation v2.3 directives:

- `@key`: Defines the primary key for User and Post entities
- `extend type Query`: Extends the Query type across schemas

## Code Generation

To regenerate the GraphQL code after schema changes:

```bash
go run github.com/99designs/gqlgen generate
```
