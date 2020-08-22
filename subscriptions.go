package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/graphql-go/graphql"
)

// key is a unique
type key int

const (
	userKey key = iota
)

type ConnectionACKMessage struct {
	OperationID string `json:"id,omitempty"`
	Type        string `json:"type"`
	Payload     struct {
		Query string `json:"query"`
	} `json:"payload,omitempty"`
}

type Subscriber struct {
	ID            int
	Conn          *websocket.Conn
	RequestString string
	OperationID   string
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Subprotocols: []string{"graphql-ws"},
}

var Subscribers sync.Map

var Subscription = graphql.NewObject(graphql.ObjectConfig{
	Name: "Subscription",
	Fields: graphql.Fields{
		"updatedUser": &graphql.Field{
			Type: UserType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				user := p.Context.Value(userKey)
				p.Context.Done()
				return user, nil
			},
		},
	},
})

func SubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to do websocket upgrade: %v", err)
		return
	}
	connectionACK, err := json.Marshal(map[string]string{
		"type": "connection_ack",
	})
	if err != nil {
		log.Printf("failed to marshal ws connection ack: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, connectionACK); err != nil {
		log.Printf("failed to write to ws connection: %v", err)
		return
	}

	go func() {
		for {
			_, p, err := conn.ReadMessage()
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				return
			}
			if err != nil {
				log.Println("failed to read websocket message: %v", err)
				return
			}
			var msg ConnectionACKMessage
			if err := json.Unmarshal(p, &msg); err != nil {
				log.Printf("failed to unmarshal: %v", err)
				return
			}
			if msg.Type == "start" {
				length := 0
				Subscribers.Range(func(key, value interface{}) bool {
					length++
					return true
				})
				var subscriber = Subscriber{
					ID:            length + 1,
					Conn:          conn,
					RequestString: msg.Payload.Query,
					OperationID:   msg.OperationID,
				}
				Subscribers.Store(subscriber.ID, &subscriber)
			}
		}
	}()
}

func publishUserUpdate(user User) {
	Subscribers.Range(func(key, value interface{}) bool {
		subscriber, ok := value.(*Subscriber)
		if !ok {
			return true
		}
		payload := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: subscriber.RequestString,
			Context:       context.WithValue(context.Background(), userKey, user),
		})
		fmt.Print("PAYLOAD", payload)
		message, err := json.Marshal(map[string]interface{}{
			"type":    "data",
			"id":      subscriber.OperationID,
			"payload": payload,
		})
		if err != nil {
			log.Printf("failed to marshal message: %v", err)
			return true
		}
		if err := subscriber.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			if err == websocket.ErrCloseSent {
				Subscribers.Delete(key)
				return true
			}
			log.Printf("failed to write to ws connection: %v", err)
			return true
		}
		return true
	})
}
