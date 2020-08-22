package main

import (
	"database/sql"

	db "github.com/chanceeakin/appbuilder-server/db"
	"github.com/graphql-go/graphql"
	log "github.com/sirupsen/logrus"
)

const selectOneStatement = `SELECT id, first_name, last_name, email, created_at, updated_at, deleted_at FROM users WHERE id=$1;`

const selectAllStatement = `SELECT id, first_name, last_name, email, created_at, updated_at, deleted_at FROM users;`

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"user": &graphql.Field{
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
						row := db.DB.QueryRow(selectOneStatement, idQuery)
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
			},
			"users": &graphql.Field{
				Type:        graphql.NewList(UserType),
				Description: "Get Users",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					users := make([]User, 0)

					rows, err := db.DB.Query(selectAllStatement)
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
			},
		},
	})
