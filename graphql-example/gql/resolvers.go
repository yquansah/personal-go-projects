package gql

import (
	"github.com/graphql-go/graphql"
	"github.com/yquansah/personal-go-projects/graphql-example/postgres"
)

type Resolver struct {
	db *postgres.Db
}

func (r *Resolver) UserResolver(p graphql.ResolveParams) (interface{}, error) {
	name, ok := p.Args["name"].(string)
	if !ok {
		return nil, nil
	}

	users := r.db.GetUsersByName(name)
	return users, nil
}
