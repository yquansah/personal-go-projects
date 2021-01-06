package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/graphql-go/graphql"
	"github.com/yquansah/personal-go-projects/graphql-example/gql"
	"github.com/yquansah/personal-go-projects/graphql-example/postgres"
	"github.com/yquansah/personal-go-projects/graphql-example/server"
)

func main() {
	router, db := initializeAPI()
	defer db.Close()

	log.Fatal(http.ListenAndServe(":4000", router))
}

func initializeAPI() (*chi.Mux, *postgres.Db) {
	router := chi.NewRouter()

	db, err := postgres.New(
		postgres.ConnString(5432, "localhost", "postgres", "go_graphql_db"),
	)
	if err != nil {
		fmt.Println("IN THIS STATEMENT")
		log.Fatal(err)
	}

	rootQuery := gql.NewRoot(db)

	sc, err := graphql.NewSchema(graphql.SchemaConfig{Query: rootQuery.Query})
	if err != nil {
		log.Fatal(err)
	}

	s := server.Server{
		GqlSchema: &sc,
	}

	router.Use(
		render.SetContentType(render.ContentTypeJSON),
	)

	router.Post("/graphql", s.GraphQL())

	return router, db
}
