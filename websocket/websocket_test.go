package websocket

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHubStartStopNoClient(t *testing.T) {
	h, stopHub := createHub()
	if len(h.clients) != 0 {
		t.Errorf("No client should be registered. Got: %+v", h.clients)
	}
	stopHub()
	if len(h.clients) != 0 {
		t.Errorf("No client should be registered. Got: %+v", h.clients)
	}
}

func TestHubRegisterDeregisterClients(t *testing.T) {
	for _, numClient := range []int{0, 1, 2} {
		t.Run(fmt.Sprintf("register and deregister %d clients", numClient), func(t *testing.T) {
			h, stopHub := createHub()
			ts := httptest.NewServer(http.HandlerFunc(h.NewClient))
			defer ts.Close()

			// connect our client(s)
			for i := 0; i < numClient; i++ {
				_, cleanup := addClient(t, ts.URL, h)
				defer cleanup()
			}

			h.muC.RLock()
			if len(h.clients) != numClient {
				t.Errorf("We expected %d clients to get registered. Got: %+v", numClient, h.clients)
			}
			h.muC.RUnlock()
			stopHub()
			if len(h.clients) != 0 {
				t.Errorf("We expected all clients to get deregistered. Got: %+v", h.clients)
			}
		})
	}
}

func TestHubSendOneMessageToClients(t *testing.T) {
	for _, numClient := range []int{0, 1, 2} {
		t.Run(fmt.Sprintf("send message to %d clients", numClient), func(t *testing.T) {
			h, cleanup := createHub()
			defer cleanup()
			ts := httptest.NewServer(http.HandlerFunc(h.NewClient))
			defer ts.Close()

			// connect our client(s)
			var clients []*websocket.Conn
			for i := 0; i < numClient; i++ {
				c, cleanup := addClient(t, ts.URL, h)
				clients = append(clients, c)
				defer cleanup()
			}

			msg := []byte("test")
			h.Send(msg)

			// ensure all clients received the desired message
			for _, c := range clients {
				_, rcv, err := c.ReadMessage()
				if err != nil {
					t.Fatalf("Unexpected received message error: %v", err)
					return
				}
				if string(rcv) != string(msg) {
					t.Errorf("We did send %s, but received %s", msg, rcv)
				}
			}

		})
	}
}

func TestHubSendMultipleMessagesToClients(t *testing.T) {
	for _, numClient := range []int{0, 1, 2} {
		t.Run(fmt.Sprintf("send multiple messages to %d clients", numClient), func(t *testing.T) {
			h, cleanup := createHub()
			defer cleanup()
			ts := httptest.NewServer(http.HandlerFunc(h.NewClient))
			defer ts.Close()

			// connect our client(s)
			var clients []*websocket.Conn
			for i := 0; i < numClient; i++ {
				c, cleanup := addClient(t, ts.URL, h)
				clients = append(clients, c)
				defer cleanup()
			}

			msgs := []string{"test", "test2"}
			for _, msg := range msgs {
				h.Send([]byte(msg))
			}

			// ensure all clients received all the desired messages in correct order
			for _, c := range clients {
				for _, msg := range msgs {
					_, rcv, err := c.ReadMessage()
					if err != nil {
						t.Fatalf("Unexpected received message error: %v", err)
						return
					}

					if string(rcv) != string(msg) {
						t.Errorf("We did send %s, but received %s", msg, rcv)
					}
				}
			}
		})
	}
}

func TestHubSendDefficientClients(t *testing.T) {
	h, cleanup := createHub()
	defer cleanup()
	ts := httptest.NewServer(http.HandlerFunc(h.NewClient))
	defer ts.Close()

	// connect our client
	_, cleanup = addClient(t, ts.URL, h)
	defer cleanup()
	h.muC.RLock()
	var cl *client
	for c := range h.clients {
		cl = c
	}
	h.muC.RUnlock()

	h.unregister <- cl
	// wait for unregistration to proceed
	<-time.After(time.Millisecond)

	h.muC.RLock()
	if len(h.clients) != 0 {
		t.Errorf("We expected all clients to get deregistered. Got: %+v", h.clients)
	}
	h.muC.RUnlock()

	// that shouldn't block
	msg := []byte("test")
	h.Send(msg)
}

// createHub and return a teardown cleanup function
func createHub() (*Hub, func()) {
	h := NewHub()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.Run()
	}()
	return h, func() {
		h.Stop()
		wg.Wait()
	}
}

func addClient(t *testing.T, httpURL string, hub *Hub) (*websocket.Conn, func()) {
	c, _, err := websocket.DefaultDialer.Dial(strings.Replace(httpURL, "http", "ws", 1), nil)
	if err != nil {
		t.Fatalf("Coudn't create websocket dialer: %v", err)
	}

	// wait for registration
	<-time.After(time.Millisecond)
	return c, func() {
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.Close()
	}
}
