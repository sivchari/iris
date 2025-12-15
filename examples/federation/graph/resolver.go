package graph

import "github.com/sivchari/iris/examples/federation/graph/model"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	users []*model.User
	posts []*model.Post
}

func NewResolver() *Resolver {
	users := []*model.User{
		{ID: "1", Name: "Alice", Email: "alice@example.com"},
		{ID: "2", Name: "Bob", Email: "bob@example.com"},
		{ID: "3", Name: "Charlie", Email: "charlie@example.com"},
	}

	posts := []*model.Post{
		{ID: "1", Title: "First Post", Body: "This is the first post", AuthorID: "1"},
		{ID: "2", Title: "Second Post", Body: "This is the second post", AuthorID: "1"},
		{ID: "3", Title: "Third Post", Body: "This is the third post", AuthorID: "2"},
	}

	return &Resolver{
		users: users,
		posts: posts,
	}
}
