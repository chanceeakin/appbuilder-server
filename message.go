package main

import (
	"fmt"
	"time"

	db "github.com/chanceeakin/appbuilder-server/db"
	"github.com/graphql-go/graphql"
)

const selectAllMessagesStatement = `SELECT id, user_id, message, created_at, updated_at, deleted_at FROM messages;`

const insertOneMessageStatement = `INSERT INTO messages (user_id, message)
VALUES ($1, $2)
RETURNING id, user_id, message, created_at, updated_at, deleted_at;`

const updateOneMessageStatement = `UPDATE messages SET message=$1 WHERE id=$2
RETURNING id, user_id, message, created_at, updated_at, deleted_at;`

const deleteOneMessageStatement = `UPDATE messages set deleted_at=$1 WHERE id=$2 RETURNING id;`

// Message type
type Message struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Message   string     `json:"message"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

/*
   MessageType object type with fields "id" and "name" by using GraphQLObjectTypeConfig:
       - Name: name of object type
       - Fields: a map of fields by using GraphQLFields
   Setup type of field use GraphQLFieldConfig
*/
var MessageType = graphql.NewObject(
	graphql.ObjectConfig{
		Name:        "message",
		Description: "A message sent",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"user_id": &graphql.Field{
				Type: graphql.Int,
			},
			"message": &graphql.Field{
				Type: graphql.String,
			},
			"created_at": &graphql.Field{
				Type: graphql.DateTime,
			},
			"updated_at": &graphql.Field{
				Type: graphql.DateTime,
			},
			"deleted_at": &graphql.Field{
				Type: graphql.DateTime,
			},
		},
	},
)

var createMessage = &graphql.Field{
	Name:        "createMessage",
	Type:        MessageType,
	Description: "Create new message",
	Args: graphql.FieldConfigArgument{
		"user_id": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.Int),
		},
		"message": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
	Resolve: func(params graphql.ResolveParams) (interface{}, error) {
		var val Message
		row := db.DB.QueryRow(insertOneMessageStatement, params.Args["user_id"].(int), params.Args["message"].(string))
		err := row.Scan(&val.ID, &val.UserID, &val.Message, &val.CreatedAt, &val.UpdatedAt, &val.DeletedAt)
		if err != nil {
			return nil, err
		}
		fmt.Println("val", val)
		publishUpdate(NewMessageKey, val)
		return val, nil
	},
}

var updateMessage = &graphql.Field{
	Name:        "updateMessage",
	Type:        MessageType,
	Description: "Update message by id",
	Args: graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.Int),
		},
		"message": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
	Resolve: func(params graphql.ResolveParams) (interface{}, error) {
		var val Message
		row := db.DB.QueryRow(updateOneUserStatement, params.Args["message"].(string), params.Args["id"].(int))
		err := row.Scan(&val.ID, &val.UserID, &val.Message, &val.CreatedAt, &val.UpdatedAt, &val.DeletedAt)
		if err != nil {
			return nil, err
		}
		publishUpdate(NewMessageKey, val)
		return val, nil
	},
}

var deleteMessage = &graphql.Field{
	Name:        "deleteMessage",
	Type:        graphql.Boolean,
	Description: "Delete user by id",
	Args: graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.Int),
		},
	},
	Resolve: func(params graphql.ResolveParams) (interface{}, error) {
		var id int
		err := db.DB.QueryRow(deleteOneMessageStatement, time.Now(), params.Args["id"].(int)).Scan(&id)
		if err != nil {
			return false, err
		}
		return true, nil
	},
}

var getMessages = &graphql.Field{
	Name:        "messages",
	Type:        graphql.NewList(MessageType),
	Description: "Get Messages",
	Resolve: func(params graphql.ResolveParams) (interface{}, error) {
		messages := make([]Message, 0)

		rows, err := db.DB.Query(selectAllMessagesStatement)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			val := Message{}
			err := rows.Scan(&val.ID, &val.UserID, &val.Message, &val.CreatedAt, &val.UpdatedAt, &val.DeletedAt)
			if err != nil {
				return rows, err
			}
			messages = append(messages, val)
		}
		err = rows.Err()
		if err != nil {
			return messages, err
		}
		return messages, nil
	},
}

var newMessage = &graphql.Field{
	Type: MessageType,
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		message := p.Context.Value(NewMessageKey)
		p.Context.Done()
		fmt.Println(message)
		return message, nil
	},
}
