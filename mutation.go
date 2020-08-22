package main

import (
	db "github.com/chanceeakin/appbuilder-server/db"
	"github.com/graphql-go/graphql"
)

const insertOneStatement = `INSERT INTO users (first_name, last_name, email)
VALUES ($1, $2, $3)
returning id, first_name, last_name, email, created_at, updated_at, deleted_at
;`

const updateOneStatement = `UPDATE users SET email=$1, first_name=$2, last_name=$3 WHERE id=$4 
returning id, first_name, last_name, email, created_at, updated_at, deleted_at
;`

const deleteOneStatement = `DELETE FROM users WHERE id=$1 RETURNING id;`

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"createUser": &graphql.Field{
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
				row := db.DB.QueryRow(insertOneStatement, params.Args["first_name"].(string), params.Args["last_name"].(string), params.Args["email"].(string))
				err := row.Scan(&val.ID, &val.FirstName, &val.LastName, &val.Email, &val.CreatedAt, &val.UpdatedAt, &val.DeletedAt)
				if err != nil {
					return nil, err
				}
				publishUserUpdate(val)
				return val, nil
			},
		},

		"updateUser": &graphql.Field{
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
				row := db.DB.QueryRow(updateOneStatement, params.Args["email"].(string), params.Args["first_name"].(string), params.Args["last_name"].(string), params.Args["id"].(int))
				err := row.Scan(&val.ID, &val.FirstName, &val.LastName, &val.Email, &val.CreatedAt, &val.UpdatedAt, &val.DeletedAt)
				if err != nil {
					return nil, err
				}
				publishUserUpdate(val)
				return val, nil
			},
		},

		// 	/* Delete product by id
		// 	   http://localhost:8080/product?query=mutation+_{delete(id:1){id,name,info,price}}
		// 	*/
		"deleteUser": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Delete user by id",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				var id int
				err := db.DB.QueryRow(deleteOneStatement, params.Args["id"].(int)).Scan(&id)
				if err != nil {
					return false, err
				}
				return true, nil
			},
		},
		// "delete": &graphql.Field{
		// 	Type:        productType,
		// 	Description: "Delete product by id",
		// 	Args: graphql.FieldConfigArgument{
		// 		"id": &graphql.ArgumentConfig{
		// 			Type: graphql.NewNonNull(graphql.Int),
		// 		},
		// 	},
		// 	Resolve: func(params graphql.ResolveParams) (interface{}, error) {
		// 		id, _ := params.Args["id"].(int)
		// 		product := Product{}
		// 		for i, p := range products {
		// 			if int64(id) == p.ID {
		// 				product = products[i]
		// 				// Remove from product list
		// 				products = append(products[:i], products[i+1:]...)
		// 			}
		// 		}

		// 		return product, nil
		// 	},
		// },
	},
})
