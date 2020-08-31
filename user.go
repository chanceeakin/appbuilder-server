package main

import (
	"database/sql"
	"time"

	db "github.com/chanceeakin/appbuilder-server/db"
	"github.com/graphql-go/graphql"
	log "github.com/sirupsen/logrus"
)

const selectOneUserStatement = `SELECT id, first_name, last_name, email, created_at, updated_at, deleted_at FROM users WHERE id=$1;`

const selectAllUsersStatement = `SELECT id, first_name, last_name, email, created_at, updated_at, deleted_at FROM users;`

const insertOneUserStatement = `INSERT INTO users (first_name, last_name, email)
VALUES ($1, $2, $3)
returning id, first_name, last_name, email, created_at, updated_at, deleted_at
;`

const updateOneUserStatement = `UPDATE users SET email=$1, first_name=$2, last_name=$3 WHERE id=$4 
returning id, first_name, last_name, email, created_at, updated_at, deleted_at
;`

const deleteOneUserStatement = `DELETE FROM users WHERE id=$1 RETURNING id;`

// User type
type User struct {
	ID        int        `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Email     string     `json:"email"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

/*
   Create User object type with fields "id" and "name" by using GraphQLObjectTypeConfig:
       - Name: name of object type
       - Fields: a map of fields by using GraphQLFields
   Setup type of field use GraphQLFieldConfig
*/
var UserType = graphql.NewObject(
	graphql.ObjectConfig{
		Name:        "user",
		Description: "A given user of the application",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"first_name": &graphql.Field{
				Type: graphql.String,
			},
			"last_name": &graphql.Field{
				Type: graphql.String,
			},
			"email": &graphql.Field{
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

var createUser = &graphql.Field{
	Type:        UserType,
	Description: "Create new user",
	Args: graphql.FieldConfigArgument{
		"first_name": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"last_name": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"email": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
	Resolve: func(params graphql.ResolveParams) (interface{}, error) {
		var val User
		row := db.DB.QueryRow(insertOneUserStatement, params.Args["first_name"].(string), params.Args["last_name"].(string), params.Args["email"].(string))
		err := row.Scan(&val.ID, &val.FirstName, &val.LastName, &val.Email, &val.CreatedAt, &val.UpdatedAt, &val.DeletedAt)
		if err != nil {
			return nil, err
		}
		publishUpdate(UserUpdatedKey, val)
		return val, nil
	},
}

var updateUser = &graphql.Field{
	Type:        UserType,
	Description: "Update product by id",
	Args: graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.Int),
		},
		"email": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"first_name": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"last_name": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
	Resolve: func(params graphql.ResolveParams) (interface{}, error) {
		var val User
		row := db.DB.QueryRow(updateOneUserStatement, params.Args["email"].(string), params.Args["first_name"].(string), params.Args["last_name"].(string), params.Args["id"].(int))
		err := row.Scan(&val.ID, &val.FirstName, &val.LastName, &val.Email, &val.CreatedAt, &val.UpdatedAt, &val.DeletedAt)
		if err != nil {
			return nil, err
		}
		publishUpdate(UserUpdatedKey, val)
		return val, nil
	},
}

var deleteUser = &graphql.Field{
	Type:        graphql.Boolean,
	Description: "Delete user by id",
	Args: graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.Int),
		},
	},
	Resolve: func(params graphql.ResolveParams) (interface{}, error) {
		var id int
		err := db.DB.QueryRow(deleteOneUserStatement, params.Args["id"].(int)).Scan(&id)
		if err != nil {
			return false, err
		}
		return true, nil
	},
}

var getUser = &graphql.Field{
	Type: UserType,
	Args: graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
			Type: graphql.Int,
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		idQuery, isOK := p.Args["id"].(int)

		if isOK {
			var returnUser User
			row := db.DB.QueryRow(selectOneUserStatement, idQuery)
			switch err := row.Scan(&returnUser.ID, &returnUser.FirstName, &returnUser.LastName, &returnUser.Email, &returnUser.CreatedAt, &returnUser.UpdatedAt, &returnUser.DeletedAt); err {
			case sql.ErrNoRows:
				log.Println("No rows were returned!")
			case nil:
				return returnUser, nil
			default:
				panic(err)
			}
		}
		return nil, nil
	},
}

var getUsers = &graphql.Field{
	Type:        graphql.NewList(UserType),
	Description: "Get Users",
	Resolve: func(params graphql.ResolveParams) (interface{}, error) {
		users := make([]User, 0)

		rows, err := db.DB.Query(selectAllUsersStatement)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			val := User{}
			err = rows.Scan(
				&val.ID,
				&val.FirstName,
				&val.LastName,
				&val.Email,
				&val.CreatedAt,
				&val.UpdatedAt,
				&val.DeletedAt,
			)
			if err != nil {
				return rows, err
			}
			users = append(users, val)
		}
		err = rows.Err()
		if err != nil {
			return users, err
		}
		return users, nil
	},
}

var updatedUser = &graphql.Field{
	Type: UserType,
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		user := p.Context.Value(UserUpdatedKey)
		p.Context.Done()
		return user, nil
	},
}
