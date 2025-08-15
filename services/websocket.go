package services

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketService struct {
	clients    map[uint]*Client
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

type Client struct {
	ID       uint
	UserID   uint
	Username string
	Conn     *websocket.Conn
	Send     chan []byte
}

type Message struct {
	Type      string      `json:"type"`
	FromUserID uint       `json:"from_user_id"`
	ToUserID   uint       `json:"to_user_id"`
	Content    string     `json:"content"`
	Timestamp  string     `json:"timestamp"`
	Data       interface{} `json:"data,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		clients:    make(map[uint]*Client),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (s *WebSocketService) StartHub() {
	for {
		select {
		case client := <-s.register:
			s.mutex.Lock()
			s.clients[client.ID] = client
			s.mutex.Unlock()

		case client := <-s.unregister:
			s.mutex.Lock()
			if _, ok := s.clients[client.ID]; ok {
				delete(s.clients, client.ID)
				close(client.Send)
			}
			s.mutex.Unlock()

		case message := <-s.broadcast:
			s.mutex.RLock()
			for _, client := range s.clients {
				// Send private message only to the intended recipient
				if message.Type == "private_message" && client.UserID == message.ToUserID {
					select {
					case client.Send <- s.serializeMessage(message):
					default:
						close(client.Send)
						delete(s.clients, client.ID)
					}
				}
			}
			s.mutex.RUnlock()
		}
	}
}

func (s *WebSocketService) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID uint, username string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:       userID,
		UserID:   userID,
		Username: username,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	s.register <- client

	go s.writePump(client)
	go s.readPump(client)
}

func (s *WebSocketService) writePump(client *Client) {
	defer func() {
		client.Conn.Close()
		s.unregister <- client
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (s *WebSocketService) readPump(client *Client) {
	defer func() {
		s.unregister <- client
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		msg.FromUserID = client.UserID
		s.broadcast <- msg
	}
}

func (s *WebSocketService) serializeMessage(message Message) []byte {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return nil
	}
	return data
}

func (s *WebSocketService) SendPrivateMessage(fromUserID, toUserID uint, content string) {
	message := Message{
		Type:      "private_message",
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Content:    content,
	}
	s.broadcast <- message
}
