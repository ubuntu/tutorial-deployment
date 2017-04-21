package websocket

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	broadcast  chan []byte
	clients    map[*client]bool
	register   chan *client
	unregister chan *client
	stop       chan struct{}
}

// NewHub generate the main Hub object to connect clients to it
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		clients:    make(map[*client]bool),
		register:   make(chan *client),
		unregister: make(chan *client),
		stop:       make(chan struct{}),
	}
}

// Run this our Hub for registering clients
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if _, ok := <-client.send; ok {
					close(client.send)
				}
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					go func() {
						// handle that in a goroutine as the receiver is in the same goroutine
						h.unregister <- client
					}()
				}
			}
		case <-h.stop:
			for c := range h.clients {
				c.conn.Close()
				delete(h.clients, c)
			}
			return
		}
	}
}

// Stop hub and disconnect all clients
func (h *Hub) Stop() {
	close(h.stop)
}

// Send a message to all connected clients
func (h *Hub) Send(msg []byte) {
	h.broadcast <- msg
}
