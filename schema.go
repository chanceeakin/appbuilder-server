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

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"user":     getUser,
			"users":    getUsers,
			"messages": getMessages,
		}})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"createUser":    createUser,
		"updateUser":    updateUser,
		"deleteUser":    deleteUser,
		"createMessage": createMessage,
		"updateMessage": updateMessage,
		"deleteMessage": deleteMessage,
	}})

// Subscription is the graphql object for subscriptions
var Subscription = graphql.NewObject(graphql.ObjectConfig{
	Name: "Subscription",
	Fields: graphql.Fields{
		"updatedUser": updatedUser,
		"newMessage":  newMessage,
	},
})
