package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const (
	redisAddr = "localhost:6379"
)

func main() {
	var port int
	flag.IntVar(&port, "p", 8080, "port to listen on")
	flag.Parse()

	globalCtx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	broadcastService := NewBroadcastService()
	broadcastController := NewBroadcastController(broadcastService, rdb)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	mux.Handle("/ws", broadcastController)

	go func() {
		broadcastController.StartSubscriber(globalCtx)
	}()

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatalf("error starting http server")
	}
}

type BroadcastController struct {
	broadcastService *BroadcastService
	upgrader         websocket.Upgrader
	rdb              *redis.Client
	redisChannelName string
}

func NewBroadcastController(broadcastService *BroadcastService, rdb *redis.Client) *BroadcastController {
	return &BroadcastController{
		broadcastService: broadcastService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		rdb:              rdb,
		redisChannelName: "broadcast-channel",
	}
}

func (b *BroadcastController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := b.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error upgrading connection:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	b.broadcastService.AddConnection(conn)
	defer b.broadcastService.RemoveConnection(conn)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("error reading message:", err)
			break
		}

		err = b.publishMessage(context.Background(), string(message))
		if err != nil {
			log.Println("error publishing message:", err)
		}
	}
}

// StartSubscriber begins a Redis subscription, which will broadcast the message
// received from Redis to all WebSocket clients currently connecting to
// this server.
func (b *BroadcastController) StartSubscriber(ctx context.Context) {
	pubsub := b.rdb.Subscribe(ctx, b.redisChannelName)
	defer pubsub.Close()

	var err error

	ch := pubsub.Channel()
	for msg := range ch {
		err = b.broadcastService.Broadcast([]byte(msg.Payload))
		if err != nil {
			log.Println("error broadcasting message:", err)
		}
	}
}

// publishMessage sends a message to Redis pubsub channel, so that it can be broadcasted
// to connections owned by other WebSocket servers as well.
func (b *BroadcastController) publishMessage(ctx context.Context, msg string) error {
	return b.rdb.Publish(ctx, b.redisChannelName, msg).Err()
}

// BroadcastService sends messages through WebSocket to registered connections.
type BroadcastService struct {
	connections []*websocket.Conn
	mux         sync.RWMutex
}

func NewBroadcastService() *BroadcastService {
	return &BroadcastService{
		connections: make([]*websocket.Conn, 0),
	}
}

func (s *BroadcastService) AddConnection(conn *websocket.Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.connections = append(s.connections, conn)
}

func (s *BroadcastService) RemoveConnection(conn *websocket.Conn) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.connections = slices.DeleteFunc(s.connections, func(c *websocket.Conn) bool {
		return c == conn
	})

	return conn.Close()
}

// Broadcast sends the given message to all registered connections.
// Return error when a message fails to send.
func (s *BroadcastService) Broadcast(message []byte) error {
	s.mux.RLock()
	defer s.mux.RUnlock()

	for _, conn := range s.connections {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return err
		}
	}

	return nil
}
