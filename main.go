package main

import (
	"context"
	"fmt"
	"strings"

	"encoding/json"
	"net/http"

	"github.com/friendsofgo/graphiql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/urfave/negroni"

	db "github.com/chanceeakin/appbuilder-server/db"
	"github.com/graphql-go/graphql"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type reqBody struct {
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
	Query         string                 `json:"query"`
}

func executeQuery(reqBody reqBody, schema graphql.Schema) (result string) {
	r := graphql.Do(graphql.Params{
		Schema:         schema,
		RequestString:  reqBody.Query,
		VariableValues: reqBody.Variables,
		OperationName:  reqBody.OperationName,
		Context:        context.Background(),
	})
	if len(r.Errors) > 0 {
		fmt.Printf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	rJSON, _ := json.Marshal(r)

	return fmt.Sprintf("%s", rJSON)
}

func gqlHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			http.Error(w, "No query data", 400)
			return
		}

		var rBody reqBody
		err := json.NewDecoder(r.Body).Decode(&rBody)
		if err != nil {
			http.Error(w, "Error parsing JSON request body", 400)
		}

		fmt.Fprintf(w, "%s", executeQuery(rBody, schema))
	})
}

func AutoCORS(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		methods, err := route.GetMethods()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods,
			","))
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	env, err := godotenv.Read()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	initData := db.InitData{Host: env["DB_HOST"], Port: env["DB_PORT"], User: env["DB_USER"], Password: env["DB_PASSWORD"], Dbname: env["DB_NAME"]}
	db.Init(&initData)
	defer db.CleanUp()

	graphiqlHandler, err := graphiql.NewGraphiqlHandler("/graphql")
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter().StrictSlash(false)
	router.HandleFunc("/subscriptions", SubscriptionHandler)

	router.
		Path("/graphql").
		Name("Graphql").
		Methods("GET", "POST").
		Handler(AutoCORS(gqlHandler()))

	router.
		Path("/").
		Name("Graphiql").
		Methods("GET", "POST").
		Handler(AutoCORS(graphiqlHandler))

	n := negroni.Classic() // Includes some default middlewares
	n.UseHandler(router)

	log.Info("Now server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", n))
}
