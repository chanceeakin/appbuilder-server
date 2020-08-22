package main

import (
	"github.com/graphql-go/graphql"
)

var (
	schema graphql.Schema
)

func init() {
	schema, _ = graphql.NewSchema(
		graphql.SchemaConfig{
			Query:        queryType,
			Mutation:     mutationType,
			Subscription: Subscription,
		},
	)
}

// Subscription is the graphql object for subscriptions
var Subscription = graphql.NewObject(graphql.ObjectConfig{
	Name: "Subscription",
	Fields: graphql.Fields{
		"updatedUser": &graphql.Field{
			Type: UserType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				user := p.Context.Value(UserUpdatedKey)
				p.Context.Done()
				return user, nil
			},
		},
	},
})
