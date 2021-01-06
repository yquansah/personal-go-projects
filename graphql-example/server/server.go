package server

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin/render"
	"github.com/graphql-go/graphql"
	"github.com/yquansah/personal-go-projects/graphql-example/gql"
)

type Server struct {
	GqlSchema *graphql.Schema
}

type reqBody struct {
	Query string `json:"query"`
}

func (s *Server) GraphQL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			http.Error(w, "Must provide query in response body", 400)
			return
		}

		var rBody reqBody

		err := json.NewDecoder(r.Body).Decode(&rBody)
		if err != nil {
			http.Error(w, "Error parsing JSON request body", 400)
		}

		result := gql.ExecuteQuery(rBody.Query, *s.GqlSchema)

		render.WriteJSON(w, result)
	}
}
