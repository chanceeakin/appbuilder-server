package main

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/graphql-go/graphql"
)

// SubscriberKey is for unique ctx.Context key/values
type SubscriberKey int

const (
	// UserUpdatedKey is the unique identifier for updated user values
	UserUpdatedKey SubscriberKey = iota
)

// ConnectionACKMessage is the sent message to a given socket subscriber
type ConnectionACKMessage struct {
	OperationID string `json:"id,omitempty"`
	Type        string `json:"type"`
	Payload     struct {
		Query string `json:"query"`
	} `json:"payload,omitempty"`
}

// Subscriber is the struct for each individual socket connection
type Subscriber struct {
	ID            int
	Conn          *websocket.Conn
	RequestString string
	OperationID   string
}

// Upgrader converts an http request into a WS connection
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Subprotocols: []string{"graphql-ws"},
}

// Subscribers is a map of values which is safe for concurrent usage
var Subscribers sync.Map

// SubscriptionHandler covers the http communication for incoming/outgoing websocket connections
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
				log.Info("failed to read websocket message: %v", err)
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

// there's probably a way to do some iota matching on the incoming parameter
func publishUpdate(UpdateKey SubscriberKey, ctxValue interface{}) {
	Subscribers.Range(func(key, value interface{}) bool {
		subscriber, ok := value.(*Subscriber)
		if !ok {
			return true
		}
		payload := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: subscriber.RequestString,
			Context:       context.WithValue(context.Background(), UpdateKey, ctxValue),
		})

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
