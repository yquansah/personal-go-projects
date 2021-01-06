package gql

import (
	"github.com/graphql-go/graphql"
	"github.com/yquansah/personal-go-projects/graphql-example/postgres"
)

type Root struct {
	Query *graphql.Object
}

func NewRoot(db *postgres.Db) *Root {
	// Resolver for the database fetch of user information
	resolver := Resolver{db: db}

	root := Root{
		Query: graphql.NewObject(
			graphql.ObjectConfig{
				Name: "Query",
				Fields: graphql.Fields{
					"users": &graphql.Field{
						Type: graphql.NewList(User),
						Args: graphql.FieldConfigArgument{
							"name": &graphql.ArgumentConfig{
								Type: graphql.String,
							},
						},
						Resolve: resolver.UserResolver,
					},
				},
			},
		),
	}

	return &root
}
