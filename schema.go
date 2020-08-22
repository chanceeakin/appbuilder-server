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
